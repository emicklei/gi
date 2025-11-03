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
	// Check if this is a heap pointer
	if hp, ok := v.Interface().(HeapPointer); ok {
		// Dereference from heap
		vm.pushOperand(vm.readHeap(hp))
	} else {
		// Regular pointer dereference
		vm.pushOperand(v.Elem())
	}
}
func (s StarExpr) Flow(g *graphBuilder) (head Step) {
	head = s.X.Flow(g)
	g.next(s)
	return
}

func (s StarExpr) Assign(vm *VM, value reflect.Value) {
	v := vm.returnsEval(s.X)
	// Check if this is a heap pointer
	if hp, ok := v.Interface().(HeapPointer); ok {
		// Write to heap
		vm.writeHeap(hp, value)
	} else {
		// Regular pointer assignment
		v.Elem().Set(value)
	}
}
func (s StarExpr) String() string {
	return fmt.Sprintf("StarExpr(%v)", s.X)
}
