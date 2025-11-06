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
func (c ChanType) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	return composite
}
func (c ChanType) Pos() token.Pos {
	return c.Begin
}
func (c ChanType) String() string {
	return fmt.Sprintf("ChanType(%v,%v)", c.Dir, c.Value)
}
