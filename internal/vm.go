package internal

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/emicklei/structexplorer"
)

type stackFrame struct {
	env          Env
	operandStack stack[reflect.Value]
	returnValues []reflect.Value
}

// push adds a value onto the operand stack.
func (f *stackFrame) push(v reflect.Value) {
	f.operandStack.push(v)
}

func (f stackFrame) withOperand(v reflect.Value) stackFrame {
	f.operandStack.push(v)
	return f
}
func (f stackFrame) withReturnValue(v reflect.Value) stackFrame {
	f.returnValues = append(f.returnValues, v)
	return f
}

// pop removes and returns the top value from the operand stack.
func (f *stackFrame) pop() reflect.Value {
	return f.operandStack.pop()
}

type VM struct {
	callStack  stack[*stackFrame] // TODO use value io pointer?
	isStepping bool
	output     *bytes.Buffer
}

func newVM(env Env) *VM {
	vm := &VM{output: new(bytes.Buffer)}
	frame := &stackFrame{env: env}
	vm.callStack.push(frame)
	return vm
}

func (vm *VM) localEnv() Env {
	return vm.callStack.top().env
}

// returnsEval evaluates the argument and returns the popped value that was pushed onto the operand stack.
func (vm *VM) returnsEval(e Evaluable) reflect.Value {
	if trace {
		fmt.Println("vm.returnsEval", e)
	}
	vm.eval(e)
	return vm.callStack.top().pop()
}

// pushOperand pushes a value onto the operand stack as the result of an evaluation.
func (vm *VM) pushOperand(v reflect.Value) {
	// TODO consider add pushOperand to callStack so stackFrame can be value that is replaced on top.

	if trace {
		if v.IsValid() && v.CanInterface() {
			fmt.Printf("vm.pushOperand: %v (%T)\n", v.Interface(), v.Interface())
		} else {
			fmt.Printf("vm.pushOperand: %v\n", v)
		}
	}
	vm.callStack.top().push(v)
}
func (vm *VM) pushNewFrame() *stackFrame {
	frame := &stackFrame{env: vm.localEnv().newChild()}
	vm.callStack.push(frame)
	return frame
}
func (vm *VM) popFrame() *stackFrame {
	return vm.callStack.pop()
}
func (vm *VM) fatal(err any) {
	s := structexplorer.NewService("vm", vm)
	for i, each := range vm.callStack {
		s.Explore(fmt.Sprintf("vm.callStack.%d", i), each, structexplorer.Column(0))
		s.Explore(fmt.Sprintf("vm.callStack.%d.env", i), each.env, structexplorer.Column(1))
		s.Explore(fmt.Sprintf("vm.callStack.%d.operandStack", i), each.operandStack, structexplorer.Column(1))
		s.Explore(fmt.Sprintf("vm.callStack.%d.returnValues", i), each.returnValues, structexplorer.Column(1))
	}
	s.Dump("vm-panic.html")
	if trace {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, "[gi] fatal error:", err)
	os.Exit(1)
}

func (vm *VM) eval(e Evaluable) {
	if trace {
		fmt.Println("vm.eval:", e)
	}
	e.Eval(vm)
}

func (vm *VM) takeAll(head Step) {
	vm.isStepping = true
	here := head
	for here != nil {
		if trace {
			fmt.Println(here)
		}
		here = here.Take(vm)
	}
}

func (vm *VM) pushCallResults(vals []reflect.Value) {
	// Push return values onto the operand stack in reverse order,
	// so the first return value ends up on top of the stack.
	for i := len(vals) - 1; i >= 0; i-- {
		vm.pushOperand(vals[i])
	}
}
