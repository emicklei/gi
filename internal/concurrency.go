package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var _ Evaluable = ChanType{}
var _ Flowable = ChanType{}
var _ CanInstantiate = ChanType{}

type ChanType struct {
	Begin token.Pos   // position of "chan" keyword or "<-" (whichever comes first)
	Arrow token.Pos   // position of "<-" (token.NoPos if there is no "<-")
	Dir   ast.ChanDir // channel direction
	Value Expr        // value type
}

func (c ChanType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(c))
}
func (c ChanType) Flow(g *graphBuilder) (head Step) {
	g.next(c)
	return g.current
}
func (c ChanType) Instantiate(vm *VM, buffer int, constructorArgs []reflect.Value) reflect.Value {
	typ := vm.returnsType(c.Value)
	dir := reflect.ChanDir(c.Dir)
	ch := reflect.ChanOf(dir, typ)
	return reflect.MakeChan(ch, int(buffer))
}
func (c ChanType) LiteralCompose(composite reflect.Value, elementType reflect.Type, values []reflect.Value) reflect.Value {
	return composite
}
func (c ChanType) Pos() token.Pos {
	return c.Begin
}
func (c ChanType) String() string {
	return fmt.Sprintf("ChanType(%v,%v)", c.Dir, c.Value)
}

var _ Expr = SendStmt{}
var _ Stmt = SendStmt{}

type SendStmt struct {
	Arrow token.Pos
	Chan  Expr
	Value Expr
}

func (s SendStmt) Eval(vm *VM) {
	// stack: value, chan
	val := vm.frameStack.top().pop()
	ch := vm.frameStack.top().pop()
	ch.Send(val)
}

func (s SendStmt) Flow(g *graphBuilder) (head Step) {
	head = s.Chan.Flow(g)
	s.Value.Flow(g)
	g.next(s)
	return head
}

func (s SendStmt) stmtStep() Evaluable { return s }

func (s SendStmt) Pos() token.Pos {
	return s.Arrow
}

func (s SendStmt) String() string {
	return fmt.Sprintf("SendStmt(%v <- %v)", s.Chan, s.Value)
}
