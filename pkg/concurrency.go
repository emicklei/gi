package pkg

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var _ Evaluable = (*ChanType)(nil)
var _ Flowable = (*ChanType)(nil)
var _ CanMake = (*ChanType)(nil)

type ChanType struct {
	beginPos  token.Pos   // position of "chan" keyword or "<-" (whichever comes first)
	dir       ast.ChanDir // channel direction
	valueType Expr        // value type
}

func (c ChanType) eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(c))
}
func (c ChanType) flow(g *graphBuilder) (head Step) {
	g.next(c)
	return g.current
}
func (c ChanType) makeValue(vm *VM, buffer int, elements []reflect.Value) reflect.Value {
	typ := makeType(vm, c.valueType)
	dir := reflect.ChanDir(c.dir)
	ch := reflect.ChanOf(dir, typ)
	return reflect.MakeChan(ch, int(buffer))
}
func (c ChanType) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	// TODO
	return composite
}
func (c ChanType) pos() token.Pos {
	return c.beginPos
}
func (c ChanType) String() string {
	return fmt.Sprintf("ChanType(%v,%v)", c.dir, c.valueType)
}

var _ Expr = SendStmt{}
var _ Stmt = SendStmt{}

type SendStmt struct {
	arrowPos token.Pos
	chann    Expr
	value    Expr
}

func (s SendStmt) eval(vm *VM) {
	// stack: value, chan
	val := vm.popOperand()
	ch := vm.popOperand()
	ch.Send(val)
}

func (s SendStmt) flow(g *graphBuilder) (head Step) {
	head = s.chann.flow(g)
	s.value.flow(g)
	g.next(s)
	return head
}

func (s SendStmt) stmtStep() Evaluable { return s }

func (s SendStmt) pos() token.Pos {
	return s.arrowPos
}

func (s SendStmt) String() string {
	return fmt.Sprintf("SendStmt(%v <- %v)", s.chann, s.value)
}
