package pkg

import (
	"fmt"
	"reflect"
	"strings"
)

var frameIdSeq int = 1 // vm is created with frame 0

// stackFrame represents a single frame in the VM's function call stack.
type stackFrame struct {
	id          int  // for debugging only, not used by the VM
	creator     Func // typically a *FuncDecl or *FuncLit
	env         Env  // current environment with name->value mapping
	operands    []reflect.Value
	defers      []funcInvocation
	currentStep Step // for using the VM to debug a function
	returnTo    Step // the step to return to after this function finishes, or nil if this is the top-level frame
}

// reset is called before putting the frame back into the pool.
func (f *stackFrame) reset() {
	f.id = 0
	f.creator = nil
	f.env = nil
	f.operands = f.operands[:0]
	f.defers = f.defers[:0]
	f.currentStep = nil
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
	child.parent = f.env // can be nil
	f.env = child
}

// popEnv reverts to the parent environment for the stack frame.
func (f *stackFrame) popEnv() {
	child := f.env.(*Environment)
	f.env = child.getParent() // can become nil
	// return child to pool
	child.parent = nil
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
