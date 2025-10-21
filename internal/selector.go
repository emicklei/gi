package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = SelectorExpr{}

type SelectorExpr struct {
	*ast.SelectorExpr
	X Expr
}

func (s SelectorExpr) Eval(vm *VM) {
	var recv reflect.Value
	if vm.isStepping {
		recv = vm.callStack.top().pop()
	} else {
		recv = vm.returnsEval(s.X)
	}
	if !recv.IsValid() {
		// propagate invalid value
		vm.pushOperand(recv)
		return
	}
	rec, ok := recv.Interface().(FieldSelectable)
	if ok {
		sel := rec.Select(s.Sel.Name)
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
