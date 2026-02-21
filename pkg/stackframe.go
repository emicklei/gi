package pkg

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"
)

// stackFrame represents a single frame in the VM's function call stack.
type stackFrame struct {
	id       int  // for debugging only, not used by the VM
	creator  Func // typically a *FuncDecl or *FuncLit
	env      Env  // current environment with name->value mapping
	operands []reflect.Value
	defers   []funcInvocation
	step     Step // for using the VM to debug a function
	returnTo Step // the step to return to after this function finishes, or nil if this is the top-level frame
}

// reset is called before putting the frame back into the pool.
func (f *stackFrame) reset() {
	f.id = 0
	f.creator = nil
	f.env = nil
	f.operands = f.operands[:0]
	f.defers = f.defers[:0]
	f.step = nil
	f.returnTo = nil
}

// push adds a value onto the operand stack.
func (f *stackFrame) push(v reflect.Value) {
	f.operands = append(f.operands, v)
}

// pop removes and returns the top value from the operand list.
func (f *stackFrame) pop() reflect.Value {
	v := f.operands[len(f.operands)-1]
	f.operands = f.operands[:len(f.operands)-1]
	return v
}

// pushEnv creates and activates a new child environment for the stack frame.
func (f *stackFrame) pushEnv() {
	child := envPool.Get().(*Environment)
	child.parentEnv = f.env // can be nil
	f.env = child
}

// popEnv reverts to the parent environment for the stack frame.
func (f *stackFrame) popEnv() {
	child := f.env.(*Environment)
	f.env = child.parent() // can become nil
	// return child to pool
	child.parentEnv = nil
	clear(child.valueTable)
	envPool.Put(child)
}

func (f *stackFrame) takeDeferList(vm *VM) {
	for i := len(f.defers) - 1; i >= 0; i-- {
		invocation := f.defers[i]
		// push all argument values as operands on the stack
		// make sure first value is on top of the operand stack
		for i := len(invocation.arguments) - 1; i >= 0; i-- {
			vm.pushOperand(invocation.arguments[i])
		}
		vm.takeAllStartingAt(invocation.flow)
	}
}

func (f *stackFrame) String() string {
	if f == nil {
		return "stackFrame(<nil>)"
	}
	buf := strings.Builder{}
	if f.creator != nil {
		fmt.Fprintf(&buf, "%v ", f.creator)
	} else {
		fmt.Fprintf(&buf, "? ")
	}
	fmt.Fprintf(&buf, "%v ", f.env)
	fmt.Fprintf(&buf, "ops=%v ", f.operands)
	return buf.String()
}

// func (f *stackFrame) buildDeferGraph() (head Step) {
// 	g := newGraphBuilder(nil)
// 	operands := make([]reflect.Value, 0)
// 	for i := len(f.defers) - 1; i >= 0; i-- {
// 		invocation := f.defers[i]
// 		// push all argument values as operands on the stack
// 		// make sure first value is on top of the operand stack
// 		for i := len(invocation.arguments) - 1; i >= 0; i-- {
// 			operands = append(operands, invocation.arguments[i])
// 		}
// 		push := &pushOperandsStep{
// 			deferPos: invocation.flow.pos(),
// 			operands: operands,
// 		}
// 		g.nextStep(push)
// 		if head == nil {
// 			head = push
// 		}
// 		g.nextStep(invocation.flow)
// 	}
// }

type pushOperandsStep struct {
	step
	deferPos token.Pos
	operands []reflect.Value
}

func (p *pushOperandsStep) take(vm *VM) Step {
	for _, operand := range p.operands {
		vm.pushOperand(operand)
	}
	return p.next
}
func (p *pushOperandsStep) pos() token.Pos {
	return p.deferPos
}
func (p *pushOperandsStep) String() string {
	return fmt.Sprintf("~pushOperand(%d)", len(p.operands))
}
