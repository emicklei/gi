package internal

import (
	"fmt"
	"go/ast"
)

var _ Expr = SelectorExpr{}

type SelectorExpr struct {
	*ast.SelectorExpr
	X Expr
}

func (s SelectorExpr) Eval(vm *VM) {
	recv := vm.frameStack.top().pop()
	// TODO value sementics for address-of
	if addrOf, ok := recv.Interface().(AddressOf); ok {
		recv = addrOf.Value
	}
	if !recv.IsValid() {
		// propagate invalid value
		vm.pushOperand(recv)
		return
	}
	rec, ok := recv.Interface().(FieldSelectable)
	if ok {
		sel := rec.Select(s.Sel.Name)
		if !sel.IsValid() {
			vm.fatal(fmt.Sprintf("field %s not found for receiver: %v (%T)", s.Sel.Name, recv.Interface(), recv.Interface()))
		}
		vm.pushOperand(sel)
		return
	}
	meth := recv.MethodByName(s.Sel.Name)
	if !meth.IsValid() {
		vm.fatal(fmt.Sprintf("method %s not found for receiver: %v (%T)", s.Sel.Name, recv.Interface(), recv.Interface()))
	}
	vm.pushOperand(meth)
}

func (s SelectorExpr) Flow(g *graphBuilder) (head Step) {
	head = s.X.Flow(g)
	g.next(s)
	return head
}

func (s SelectorExpr) String() string {
	return fmt.Sprintf("SelectorExpr(%v, %v)", s.X, s.SelectorExpr.Sel.Name)
}
