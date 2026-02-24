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
	pkg          *Package
	callStack    stack[*stackFrame]
	frameIdSeq   int
	currentFrame *stackFrame // optimization
	heap         *Heap
	output       *bytes.Buffer // for testing only
}

func NewVM(pkg *Package) *VM {
	vm := &VM{
		pkg:        pkg,
		frameIdSeq: 1, // vm is created with frame 0 on stack
		output:     new(bytes.Buffer),
		callStack:  make(stack[*stackFrame], 0, 16),
		heap:       newHeap(),
	}
	// frame := framePool.Get().(*stackFrame)
	// frame.env = pkg.env
	// vm.callStack.push(frame)
	// vm.currentFrame = frame
	return vm
}

// currentEnv returns the current environment from the top stack frame
// or the package environment if no frame is available.
func (vm *VM) currentEnv() Env {
	if vm.currentFrame == nil {
		return vm.pkg.env
	}
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
				fmt.Printf("~~ frame.%d.push [%d->%d]: untyped nil\n", vm.currentFrame.id, before, before+1)
			} else if isUndeclared(v) {
				fmt.Printf("~~ frame.%d.push [%d->%d]: %v (undeclared)\n", vm.currentFrame.id, before, before+1, v)
			} else {
				fmt.Printf("~~ frame.%d.push [%d->%d]: %v (%T)\n", vm.currentFrame.id, before, before+1, v.Interface(), v.Interface())
			}
		} else {
			fmt.Printf("~~ frame.%d.push [%d->%d]: %v\n", vm.currentFrame.id, before, before+1, v)
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
		fmt.Printf("~~ frame.%d.pop [%d->%d]: %s (%T)\n", vm.currentFrame.id, before+1, before, stringOf(popped), popped.Interface())
	}
	return popped
}

func (vm *VM) pushNewFrame(creator Func) {
	frame := framePool.Get().(*stackFrame)
	frame.id = vm.frameIdSeq
	vm.frameIdSeq++
	frame.callee = creator
	env := envPool.Get().(*Environment)
	env.parentEnv = vm.currentEnv()
	frame.env = env

	// remember return
	if vm.currentFrame != nil && vm.currentFrame.step != nil {
		frame.returnTo = vm.currentFrame.step.Next()
		if trace {
			fmt.Printf("vm.pushNewFrame.%d: set returnTo to %v\n", frame.id, frame.returnTo)
		}
	}
	vm.callStack.push(frame)
	vm.currentFrame = frame
	if trace {
		fmt.Printf("vm.pushNewFrame.%d:%s\n", frame.id, stringOf(creator))
	}
}

func (vm *VM) popFrame() {
	if trace {
		if len(vm.callStack) == 0 {
			vm.fatalf("no frame left on the call stack")
		}
		fmt.Printf("vm.popFrame.%d\n", vm.callStack.top().id)
	}
	frame := vm.callStack.pop()
	if len(vm.callStack) > 0 {
		vm.currentFrame = vm.callStack.top()
		vm.currentFrame.step = frame.returnTo
		if trace {
			fmt.Printf("vm.currentFrame.%d.step set to returnTo: %v\n", vm.currentFrame.id, vm.currentFrame.step)
		}
	} else {
		vm.currentFrame = nil
	}

	// return env to pool
	env, ok := frame.env.(*Environment)
	// skip PkgEnvironment
	if ok {
		env.parentEnv = nil
		// do not recycle environments that contain values referenced by a heap pointer
		if !env.hasHeapPointer {
			clear(env.valueTable)
			envPool.Put(env)
		}
	}
	// return frame to pool
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
	e.eval(vm)
}

func (vm *VM) takeAllStartingAt(head Step) {
	vm.currentFrame.step = head
}

// Next takes the current step and advances to the next step, returning an error if there are no more steps to take (i.e., EOF).
// Pre: vm.currentFrame not nil
func (vm *VM) Next() error {
	frame := vm.currentFrame
	// EOF means function is done
	if frame.step == nil {
		return io.EOF
	}
	if trace {
		if vm.pkg == nil || vm.pkg.Fset == nil {
			fmt.Printf("%v @ <no fileset>\n", frame.step)
		} else {
			fmt.Printf("%v @ %v\n", frame.step, cursor(vm.pkg.Fset, frame.step.pos()))
		}
	}
	frame.step.take(vm)
	return nil
}

func (vm *VM) stepThrough(flow Step) {
	vm.currentFrame.step = flow
	for here := vm.currentFrame.step; here != nil; {
		here.take(vm)
		if vm.currentFrame.step == here {
			// if the step did not change, then move to the next step.
			vm.currentFrame.step = here.Next()
		}
	}
}

// Launch sets up the VM for execution of the given function name with the provided arguments.
func (vm *VM) Launch(functionName string, args []any) {
	vm.launch(functionName, args)
}

func (vm *VM) callPackageFunction(functionName string, args []any) ([]any, error) {
	vm.launch(functionName, args)
	for {
		if err := vm.Next(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error during execution: %v", err)
		}
	}

	// collect non-reflection return values
	top := vm.currentFrame
	vals := []any{}
	for len(top.operands) > 0 {
		val := top.pop()
		vals = append(vals, val.Interface())
	}
	return vals, nil
}

// launch sets the call flow.
func (vm *VM) launch(functionName string, args []any) {
	vm.pushNewFrame(nil)
	// make sure PkgEnvironment is active
	vm.currentFrame.env = vm.pkg.env

	initPkg := newFuncStep(token.NoPos, "initpkg", func(vm *VM) {
		vm.pushNewFrame(nil)
		// make sure PkgEnvironment is active
		vm.currentFrame.env = vm.pkg.env
		vm.currentFrame.step = vm.pkg.callGraph
	})

	pushArgs := newFuncStep(token.NoPos, fmt.Sprintf("push args for %s.%s", vm.pkg.Name, functionName), func(vm *VM) {
		// push arguments as parameters on the operand stack, in reverse order
		for i := len(args) - 1; i >= 0; i-- {
			vm.pushOperand(reflect.ValueOf(args[i]))
		}
	})

	// add noop expressions as arguments; the values will be pushed on the operand stack
	callArgs := make([]Expr, len(args))
	for i := range len(args) {
		callArgs[i] = noExpr{}
	}
	// make a CallExpr and reuse its logic to set up the call
	call := CallExpr{
		fun:  Ident{name: functionName},
		args: callArgs,
	}
	gb := newGraphBuilder(vm.pkg.Package)
	gb.nextStep(initPkg)
	gb.nextStep(pushArgs)
	call.flow(gb)

	vm.currentFrame.step = initPkg // head of flow
}

func (vm *VM) printStack() {
	if len(vm.callStack) == 0 {
		fmt.Println("vm.ops: <empty>")
		return
	}
	frame := vm.currentFrame
	if env, ok := frame.env.(*PkgEnvironment); ok {
		for i, decl := range env.declarations {
			fmt.Printf("pkg.decl.%d: %v\n", i, decl)
			if cd, ok := decl.(ConstDecl); ok {
				for s, spec := range cd.specs {
					for n, idn := range spec.names {
						fmt.Printf("  const.spec.%d.%d: %v\n", s, n, idn.name)
					}
				}
			}
		}
		for i, method := range env.inits {
			fmt.Printf("pkg.init.%d: %v\n", i, method)
		}
		for i, method := range env.methods {
			fmt.Printf("pkg.method.%d: %v\n", i, method)
		}
	}
	if env, ok := frame.env.(*Environment); ok {
		for k, v := range env.valueTable {
			if v.IsValid() && v.CanInterface() {
				if v == reflectNil {
					fmt.Printf("vm.env.%s: untyped nil\n", k)
					continue
				}
				if isUndeclared(v) {
					fmt.Printf("vm.env.%s: undeclared value\n", k)
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
