package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = StarExpr{}

type StarExpr struct {
	*ast.StarExpr
	X Expr
}

func (s StarExpr) Eval(vm *VM) {
	v := vm.returnsEval(s.X)
	vm.pushOperand(v.Elem())
}
func (s StarExpr) Flow(g *graphBuilder) (head Step) {
	head = s.X.Flow(g)
	g.next(s)
	return
}

func (s StarExpr) Assign(vm *VM, value reflect.Value) {
	// TODO
	//v := vm.ReturnsEval(s.X)
	//v.Elem().Set(value)
}
func (s StarExpr) String() string {
	return fmt.Sprintf("StarExpr(%v)", s.X)
}
