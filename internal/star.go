package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = StarExpr{}
var _ CanAssign = StarExpr{}

type StarExpr struct {
	*ast.StarExpr
	X Expr
}

func (s StarExpr) Eval(vm *VM) {
	v := vm.returnsEval(s.X)
	// Handle VarPointer specially
	if v.Kind() == reflect.Ptr && v.Type().String() == "*internal.VarPointer" {
		if vp, ok := v.Interface().(*VarPointer); ok {
			vm.pushOperand(vp.Deref())
			return
		}
	}
	vm.pushOperand(v.Elem())
}
func (s StarExpr) Flow(g *graphBuilder) (head Step) {
	head = s.X.Flow(g)
	g.next(s)
	return
}

func (s StarExpr) Assign(vm *VM, value reflect.Value) {
	v := vm.returnsEval(s.X)
	// Handle VarPointer specially
	if v.Kind() == reflect.Ptr && v.Type().String() == "*internal.VarPointer" {
		if vp, ok := v.Interface().(*VarPointer); ok {
			vp.Assign(value)
			return
		}
	}
	if v.Kind() != reflect.Ptr {
		vm.fatal(fmt.Sprintf("cannot dereference non-pointer type: %v", v.Kind()))
	}
	if v.IsNil() {
		vm.fatal("cannot dereference nil pointer")
	}
	v.Elem().Set(value)
}

func (s StarExpr) Define(vm *VM, value reflect.Value) {
	// Define through a pointer doesn't make sense in Go
	// This would be like *p := value, which is invalid syntax
	vm.fatal("cannot use := with pointer dereference")
}
func (s StarExpr) String() string {
	return fmt.Sprintf("StarExpr(%v)", s.X)
}
