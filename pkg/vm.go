package pkg

import (
	"bytes"
	"fmt"
	"go/token"
	"io"
	"os"
	"reflect"
	"sync"
)

// framePool is a pool of stackFrame values for reuse.
var framePool = sync.Pool{
	New: func() any {
		return &stackFrame{
			operands: make([]reflect.Value, 0, 8),
		}
	},
}

// Runtime represents a virtual machine that can execute Go code.
type VM struct {
	callStack    stack[*stackFrame]
	currentFrame *stackFrame // optimization
	heap         *Heap
	output       *bytes.Buffer  // for testing only
	fileSet      *token.FileSet // optional file set for position info
	isStepping   bool           // whether the VM is currently stepping through code
}

func NewVM(pkg *Package) *VM {
	vm := &VM{
		output:    new(bytes.Buffer),
		callStack: make(stack[*stackFrame], 0, 16),
		heap:      newHeap()}
	// happens in tests
	if pkg.Package != nil {
		vm.fileSet = pkg.Fset
	}
	frame := framePool.Get().(*stackFrame)
	frame.env = pkg.env
	vm.callStack.push(frame)
	vm.currentFrame = frame
	return vm
}

// currentEnv returns the current environment from the top stack frame.
func (vm *VM) currentEnv() Env {
	return vm.currentFrame.env
}

// returnsEval evaluates the argument and returns the popped value that was pushed onto the operand stack.
func (vm *VM) returnsEval(e Evaluable) reflect.Value {
	vm.eval(e)
	return vm.popOperand()
}

// pushOperand pushes a value onto the operand stack as the result of an evaluation.
func (vm *VM) pushOperand(v reflect.Value) {
	if trace {
		before := len(vm.currentFrame.operands)
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("~~ frame[%d].push [%d->%d]: untyped nil\n", vm.currentFrame.id, before, before+1)
			} else if isUndeclared(v) {
				fmt.Printf("~~ frame[%d].push [%d->%d]: %v (undeclared)\n", vm.currentFrame.id, before, before+1, v)
			} else {
				fmt.Printf("~~ frame[%d].push [%d->%d]: %v (%T)\n", vm.currentFrame.id, before, before+1, v.Interface(), v.Interface())
			}
		} else {
			fmt.Printf("~~ frame[%d].push [%d->%d]: %v\n", vm.currentFrame.id, before, before+1, v)
		}
	}
	vm.currentFrame.push(v)
}

// pushOperands pushes multiple values onto the operand stack in reverse order,
// so the first value ends up on top of the stack.
func (vm *VM) pushOperands(vals ...reflect.Value) {
	for i := len(vals) - 1; i >= 0; i-- {
		vm.pushOperand(vals[i])
	}
}

// popOperand pops a value from the operand stack.
func (vm *VM) popOperand() reflect.Value {
	if trace {
		if len(vm.currentFrame.operands) == 0 {
			vm.fatalf("no operand left on the stack")
		}
	}
	popped := vm.currentFrame.pop()
	if trace {
		before := len(vm.currentFrame.operands)
		fmt.Printf("~~ frame[%d].pop [%d->%d]:%s\n", vm.currentFrame.id, before, before-1, stringOf(popped))
	}
	return popped
}

func (vm *VM) pushNewFrame(f Func) {
	frame := framePool.Get().(*stackFrame)
	frame.id = frameIdSeq
	frameIdSeq++
	frame.creator = f
	env := envPool.Get().(*Environment)
	env.parent = vm.currentEnv()
	frame.env = env

	// remember return
	if vm.isStepping && vm.currentFrame.step != nil {
		vm.currentFrame.returnTo = vm.currentFrame.step.Next()
	}
	vm.callStack.push(frame)
	vm.currentFrame = frame
	if trace {
		fmt.Printf("vm.pushNewFrame[%d]:%s\n", frame.id, stringOf(f))
	}
}

func (vm *VM) popFrame() {
	if trace {
		if len(vm.callStack) == 0 {
			vm.fatalf("no frame left on the call stack")
		}
	}
	frame := vm.callStack.pop()
	if trace {
		fmt.Printf("vm.popFrame[%d]\n", frame.id)
	}
	if len(vm.callStack) > 0 {
		vm.currentFrame = vm.callStack.top()
		if vm.isStepping {
			//consume return
			vm.currentFrame.step = vm.currentFrame.returnTo
			vm.currentFrame.returnTo = nil
		}
	} else {
		vm.currentFrame = nil
	}

	// return env to pool
	env := frame.env.(*Environment)
	env.parent = nil
	// do not recycle environments that contain values referenced by a heap pointer
	if !env.hasHeapPointer {
		clear(env.valueTable)
		envPool.Put(env)
	}
	frame.reset()
	framePool.Put(frame)
}

// fatalf reports a fatal error and stops execution.
func (vm *VM) fatalf(format string, a ...any) {
	line := fmt.Sprintf(format, a...)
	fmt.Fprintln(os.Stderr, "[gi] fatal error: "+line)
	fmt.Fprintln(os.Stderr, "")
	vm.printStack()
	panic(line)
}

func (vm *VM) eval(e Evaluable) {
	if trace {
		fmt.Fprintln(os.Stderr, "vm.eval:", e)
	}
	e.Eval(vm)
}

func (vm *VM) takeAllStartingAt(head Step) {
	// TODO stepping will be the default behavior

	if vm.isStepping {
		vm.currentFrame.step = head
		return
	}
	here := head
	for here != nil {
		if trace {
			fmt.Printf("%v @ %s\n", here, sourceLocation(vm.fileSet, here.Pos()))
		}
		here = here.take(vm)
	}
}

// take one step
func (vm *VM) Step() error {
	frame := vm.currentFrame
	here := frame.step
	// EOF means function is done
	if here == nil {
		return io.EOF
	}
	// take the step and return the next or nil
	next := here.take(vm)
	// proceed with next if in same frame
	// if not the currentStep is reset for the new frame
	if vm.currentFrame == frame {
		frame.step = next
	}
	return nil
}
func (vm *VM) Location() string {
	s := vm.currentFrame.step
	if s == nil {
		return "no current step"
	}
	loc := "no fileset"
	if vm.fileSet != nil {
		if s.Pos() == token.NoPos {
			return "no position info"
		}
		loc = sourceLocation(vm.fileSet, s.Pos())
	}
	return fmt.Sprintf("%v @ %s", s, loc)
}
func (vm *VM) Setup(pkg *Package, funcName string, args []any) {
	vm.isStepping = false // Temporary
	pkg.initialize(vm)

	fun := vm.currentEnv().valueLookUp(funcName)
	call := CallExpr{}
	if args != nil {
		// add noop expressions as arguments; the values will be pushed on the operand stack
		for range len(args) {
			call.args = append(call.args, noExpr{})
		}
		// push arguments as parameters on the operand stack, in reverse order
		for i := len(args) - 1; i >= 0; i-- {
			vm.pushOperand(reflect.ValueOf(args[i]))
		}
	}
	// until w have breakpoints
	vm.isStepping = true
	call.handleFuncDecl(vm, fun.Interface().(*FuncDecl))
}

func (vm *VM) printStack() {
	if len(vm.callStack) == 0 {
		fmt.Println("vm.ops: <empty>")
		return
	}
	frame := vm.currentFrame
	if env, ok := frame.env.(*PkgEnvironment); ok {
		for i, decl := range env.declarations {
			fmt.Printf("pkg.decl[%d]: %v\n", i, decl)
			if cd, ok := decl.(ConstDecl); ok {
				for s, spec := range cd.specs {
					for n, idn := range spec.names {
						fmt.Printf("  const.spec[%d][%d]: %v\n", s, n, idn.name)
					}
				}
			}
		}
		for i, method := range env.inits {
			fmt.Printf("pkg.init[%d]: %v\n", i, method)
		}
		for i, method := range env.methods {
			fmt.Printf("pkg.method[%d]: %v\n", i, method)
		}
	}
	if env, ok := frame.env.(*Environment); ok {
		for k, v := range env.valueTable {
			if v.IsValid() && v.CanInterface() {
				if v == reflectNil {
					fmt.Printf("vm.env[%s]: untyped nil\n", k)
					continue
				}
				if isUndeclared(v) {
					fmt.Printf("vm.env[%s]: undeclared value\n", k)
					continue
				}
				fmt.Printf("vm.env.%s = %s (%T)\n", k, stringOf(v.Interface()), v.Interface())
			} else {
				fmt.Printf("vm.env.%s = %s\n", k, stringOf(v))
			}
		}
	}
	for i := 0; i < len(frame.operands); i++ {
		v := frame.operands[i]
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("vm.ops.%d: untyped nil\n", i)
				continue
			}
			if isUndeclared(v) {
				fmt.Printf("vm.ops.%d: undeclared value\n", i)
				continue
			}
			fmt.Printf("vm.ops.%d: %s (%T)\n", i, stringOf(v.Interface()), v.Interface())
		} else {
			fmt.Printf("vm.ops.%d: %s\n", i, stringOf(v))
		}
	}
}
