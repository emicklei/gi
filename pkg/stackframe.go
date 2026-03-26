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
	callee   Func // typically a *FuncDecl or *FuncLit
	env      Env  // current environment with name->value mapping
	operands []reflect.Value
	// results  []reflect.Value // for storing return values of the function
	defers   []funcInvocation
	step     Step // for using the VM to debug a function
	returnTo Step // the step to return to after this function finishes, or nil if this is the top-level frame
}

// reset is called before putting the frame back into the pool.
func (f *stackFrame) reset() {
	f.id = 0
	f.callee = nil
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

	// cannot recycle env if heappointer is referencing it
	if !child.isShared {
		child.parentEnv = nil
		clear(child.values)
		envPool.Put(child)
	}
}

var _ Stmt = (*pushArgumentsStmt)(nil)

type pushArgumentsStmt struct {
	args []reflect.Value
	env  Env // TODO why is this needed
}

func (p *pushArgumentsStmt) eval(vm *VM) {
	vm.currentFrame.env = p.env
	// push all argument values as operands on the stack
	// make sure first value is on top of the operand stack
	for i := len(p.args) - 1; i >= 0; i-- {
		vm.pushOperand(p.args[i])
	}
}
func (p *pushArgumentsStmt) flow(g *graphBuilder) (head Step) {
	g.next(p)
	return g.current
}
func (p *pushArgumentsStmt) stmtStep() Evaluable {
	return p
}
func (p *pushArgumentsStmt) pos() token.Pos {
	return token.NoPos
}
func (p *pushArgumentsStmt) String() string {
	return fmt.Sprintf("pushArguments(len=%d)", len(p.args))
}

func (f *stackFrame) String() string {
	if f == nil {
		return "stackFrame(<nil>)"
	}
	buf := strings.Builder{}
	if f.callee != nil {
		fmt.Fprintf(&buf, "%v ", f.callee)
	} else {
		fmt.Fprintf(&buf, "? ")
	}
	fmt.Fprintf(&buf, "%v ", f.env)
	fmt.Fprintf(&buf, "ops=%v ", f.operands)
	return buf.String()
}
