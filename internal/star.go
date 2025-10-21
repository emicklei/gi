package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

type StarExpr struct {
	X Expr
	*ast.StarExpr
}

func (s StarExpr) Eval(vm *VM) {
	v := vm.returnsEval(s.X)
	vm.pushOperand(v.Elem())
}
func (s StarExpr) Assign(vm *VM, value reflect.Value) {
	// TODO
	//v := vm.ReturnsEval(s.X)
	//v.Elem().Set(value)
}
func (s StarExpr) String() string {
	return fmt.Sprintf("StarExpr(%v)", s.X)
}
