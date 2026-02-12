package pkg

import (
	"bytes"
	"fmt"
	"go/token"
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
	currentStep  Step           // for using the VM to debug a function
}

func NewVM(pkg *Package) *VM {
	vm := &VM{
		output:    new(bytes.Buffer),
		callStack: make(stack[*stackFrame], 0, 16),
		heap:      newHeap()}
	frame := framePool.Get().(*stackFrame)
	frame.env = pkg.env
	// happens in tests
	if pkg.Package != nil {
		vm.fileSet = pkg.Fset
	}
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
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("vm.push: untyped nil\n")
			} else if isUndeclared(v) {
				fmt.Printf("vm.push: %v (undeclared)\n", v)
			} else {
				fmt.Printf("vm.push: %v (%T)\n", v.Interface(), v.Interface())
			}
		} else {
			fmt.Printf("vm.push: %v\n", v)
		}
	}
	vm.currentFrame.push(v)
}

// popOperand pops a value from the operand stack.
func (vm *VM) popOperand() reflect.Value {
	return vm.currentFrame.pop()
}

func (vm *VM) pushNewFrame(f Func) {
	if trace {
		fmt.Println("vm.pushNewFrame:", f)
	}
	frame := framePool.Get().(*stackFrame)
	frame.creator = f
	//frame.returnTo = vm.currentStep.Next()
	env := envPool.Get().(*Environment)
	env.parent = vm.currentEnv()
	frame.env = env
	vm.callStack.push(frame)
	vm.currentFrame = frame
}

func (vm *VM) popFrame() {
	if trace {
		fmt.Println("vm.popFrame")
	}
	frame := vm.callStack.pop()
	if len(vm.callStack) > 0 {
		vm.currentFrame = vm.callStack.top()
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

	// reset references
	frame.operands = frame.operands[:0]
	frame.env = nil
	frame.creator = nil
	frame.defers = frame.defers[:0]
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
	here := head
	for here != nil {
		if trace {
			fmt.Printf("%v @ %s\n", here, sourceLocation(vm.fileSet, here.Pos()))
		}
		here = here.take(vm)
	}
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
