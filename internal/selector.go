package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = SelectorExpr{}
var _ CanAssign = SelectorExpr{}

type SelectorExpr struct {
	*ast.SelectorExpr
	X Expr
}

func (s SelectorExpr) Define(vm *VM, val reflect.Value) {}

func (s SelectorExpr) Assign(vm *VM, val reflect.Value) {
	if idn, ok := s.X.(Ident); ok {

		// need to pop from stack? TODO

		recv := vm.localEnv().valueLookUp(idn.Name)

		rec, ok := recv.Interface().(FieldAssignable)
		if ok {
			rec.Assign(s.Sel.Name, val)
		}
		return
	}

	recv := vm.returnsEval(s.X)
	// check for pointer to heap value
	if hp, ok := recv.Interface().(HeapPointer); ok {
		recv = vm.heap.read(hp)
	}
	if !recv.IsValid() {
		vm.fatal("cannot assign to invalid selector receiver")
	}
	rec, ok := recv.Interface().(FieldSelectable)
	if ok {
		sel := rec.Select(s.Sel.Name)
		if !sel.IsValid() {
			vm.fatal(fmt.Sprintf("field %s not found for receiver: %v (%T)", s.Sel.Name, recv.Interface(), recv.Interface()))
		}
		if !sel.CanSet() {
			vm.fatal(fmt.Sprintf("field %s is not settable for receiver: %v (%T)", s.Sel.Name, recv.Interface(), recv.Interface()))
		}
		sel.Set(val)
		return
	}
	vm.fatal(fmt.Sprintf("cannot assign to method %s for receiver: %v (%T)", s.Sel.Name, recv.Interface(), recv.Interface()))
}

func (s SelectorExpr) Eval(vm *VM) {
	recv := vm.frameStack.top().pop()
	// check for pointer to heap value
	if hp, ok := recv.Interface().(HeapPointer); ok {
		recv = vm.heap.read(hp)
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
