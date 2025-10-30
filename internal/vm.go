package internal

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/emicklei/structexplorer"
)

var framePool = sync.Pool{
	New: func() any {
		return &stackFrame{}
	},
}

type stackFrame struct {
	creator      Evaluable
	env          Env
	operandStack []reflect.Value
	returnValues []reflect.Value
}

// push adds a value onto the operand stack.
func (f *stackFrame) push(v reflect.Value) {
	f.operandStack = append(f.operandStack, v)
}

// pop removes and returns the top value from the operand stack.
func (f *stackFrame) pop() reflect.Value {
	v := f.operandStack[len(f.operandStack)-1]
	f.operandStack = f.operandStack[:len(f.operandStack)-1]
	return v
}

func (f *stackFrame) String() string {
	buf := strings.Builder{}
	if f.creator != nil {
		fmt.Fprintf(&buf, "%v ", f.creator)
	} else {
		fmt.Fprintf(&buf, "? ")
	}
	fmt.Fprintf(&buf, "%v ", f.env)
	fmt.Fprintf(&buf, "ops=%v ", f.operandStack)
	fmt.Fprintf(&buf, "rets=%v", f.returnValues)
	return buf.String()
}

// VM represents a virtual machine that can execute Go code.
type VM struct {
	frameStack      stack[*stackFrame]
	isStepping      bool                   // for debugging use
	activeFuncStack stack[*activeFuncDecl] // active function declarations, TODO store in stackFrame?
	output          *bytes.Buffer          // for testing only
}

func newVM(env Env) *VM {
	vm := &VM{output: new(bytes.Buffer)}
	frame := framePool.Get().(*stackFrame)
	frame.env = env
	vm.frameStack.push(frame)
	return vm
}

// localEnv returns the current environment from the top stack frame.
func (vm *VM) localEnv() Env {
	return vm.frameStack.top().env
}

// returnsEval evaluates the argument and returns the popped value that was pushed onto the operand stack.
func (vm *VM) returnsEval(e Evaluable) reflect.Value {
	if trace {
		vm.traceEval(e)
	} else {
		e.Eval(vm)
	}
	return vm.frameStack.top().pop()
}

func (vm *VM) returnsType(e Evaluable) reflect.Type {
	if id, ok := e.(Ident); ok {
		typ, ok := builtinTypesMap[id.Name]
		if ok {
			return typ
		}
	}
	if star, ok := e.(StarExpr); ok {
		nonStarType := vm.returnsType(star.X)
		return reflect.PointerTo(nonStarType)
	}
	vm.fatal(fmt.Sprintf("todo returnsType	for %v", e))
	return nil
}

// pushOperand pushes a value onto the operand stack as the result of an evaluation.
func (vm *VM) pushOperand(v reflect.Value) {
	if trace {
		if v.IsValid() && v.CanInterface() {
			fmt.Printf("vm.push: %v (%T)\n", v.Interface(), v.Interface())
		} else {
			fmt.Printf("vm.push: %v\n", v)
		}
	}
	vm.frameStack.top().push(v)
}

func (vm *VM) pushNewFrame(e Evaluable) {
	frame := framePool.Get().(*stackFrame)
	frame.creator = e
	frame.env = vm.localEnv().newChild()
	vm.frameStack.push(frame)
}

func (vm *VM) popFrame() {
	frame := vm.frameStack.pop()
	// reset references
	frame.operandStack = frame.operandStack[:0]
	frame.returnValues = frame.returnValues[:0]
	frame.env = nil
	frame.creator = nil
	framePool.Put(frame)
}

func (vm *VM) fatal(err any) {
	fmt.Fprintln(os.Stderr, "[gi] fatal error:", err)
	fmt.Fprintln(os.Stderr, "")
	// dump the callstack
	for i := len(vm.frameStack) - 1; i >= 0; i-- {
		frame := vm.frameStack[i]
		fmt.Fprintln(os.Stderr, "[gi]", frame)
	}
	s := structexplorer.NewService("vm", vm)
	for i, each := range vm.frameStack {
		s.Explore(fmt.Sprintf("vm.callStack.%d", i), each, structexplorer.Column(0))
		s.Explore(fmt.Sprintf("vm.callStack.%d.env", i), each.env, structexplorer.Column(1))
		s.Explore(fmt.Sprintf("vm.callStack.%d.operandStack", i), each.operandStack, structexplorer.Column(1))
		s.Explore(fmt.Sprintf("vm.callStack.%d.returnValues", i), each.returnValues, structexplorer.Column(1))
	}
	s.Dump("gi-vm-panic.html")
	if trace {
		panic(err)
	}
	os.Exit(1)
}

func (vm *VM) traceEval(e Evaluable) {
	fmt.Fprintln(os.Stderr, "vm.eval:", e)
	e.Eval(vm)
}

func (vm *VM) takeAll(head Step) {
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
