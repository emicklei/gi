package pkg

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
		if trace {
			fmt.Println("TRACE: SelectorExpr.Assign", idn.Name, s.Sel.Name, "=", val, "operands:", vm.currentFrame.operands)
		}

		recv := vm.localEnv().valueLookUp(idn.Name)

		// dereference if pointer to heap value
		if hp, ok := recv.Interface().(*HeapPointer); ok {
			recv = vm.heap.read(hp)
		}
		// can we assign directly to the field?
		fa, ok := recv.Interface().(FieldAssignable)
		if ok {
			fa.Assign(s.Sel.Name, val)
			return
		}
		// TODO missing case?

		vm.fatal(fmt.Sprintf("cannot assign to field %s for receiver: %v (%T)", s.Sel.Name, recv.Interface(), recv.Interface()))
		return
	}
	recv := vm.returnsEval(s.X)

	// dereference if pointer to heap value
	if hp, ok := recv.Interface().(*HeapPointer); ok {
		recv = vm.heap.read(hp)
	}
	if !recv.IsValid() {
		vm.fatal("cannot assign to invalid selector receiver")
	}
	rec, ok := recv.Interface().(CanSelect)
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
	recv := vm.popOperand()
	// check for pointer to heap value
	if hp, ok := recv.Interface().(*HeapPointer); ok {
		recv = vm.heap.read(hp)
	}
	if isUndeclared(recv) {
		// propagate invalid value
		vm.pushOperand(recv)
		return
	}
	rec, ok := recv.Interface().(CanSelect)
	if ok {
		// can be field or method
		sel := rec.Select(s.Sel.Name)
		// check for method
		if _, ok := sel.Interface().(*FuncDecl); ok {
			// method value so push receiver as first argument
			vm.pushOperand(recv)
		}
		vm.pushOperand(sel)
		return
	}

	if recv.Kind() == reflect.Struct {
		field := recv.FieldByName(s.Sel.Name)
		if field.IsValid() {
			vm.pushOperand(field)
			return
		}
	}

	if recv.Kind() == reflect.Pointer {
		nonPtrRecv := recv.Elem()
		if nonPtrRecv.Kind() == reflect.Struct {
			field := nonPtrRecv.FieldByName(s.Sel.Name)
			if field.IsValid() {
				vm.pushOperand(field)
				return
			}
		}
	}

	meth := recv.MethodByName(s.Sel.Name)
	if !meth.IsValid() {
		recvType := recv.Type()
		ptrRecvType := reflect.PointerTo(recvType)
		pmeth, ok := ptrRecvType.MethodByName(s.Sel.Name)
		if ok {
			meth = reflect.ValueOf(pmeth)
			// push pointer to recv as first argument
			if recv.CanAddr() {
				recv = recv.Addr()
			} else {
				// Create a new pointer to a copy
				ptr := reflect.New(recv.Type())
				ptr.Elem().Set(recv)
				recv = ptr
			}
			vm.pushOperand(recv)
			vm.pushOperand(meth)
			return
		}
	} else {
		vm.pushOperand(meth)
		return
	}

	vm.fatal(fmt.Sprintf("method or field \"%s\" not found for receiver: %v (%T)", s.Sel.Name, recv.Interface(), recv.Interface()))
}

func (s SelectorExpr) Flow(g *graphBuilder) (head Step) {
	head = s.X.Flow(g)
	g.next(s)
	return head
}

func (s SelectorExpr) String() string {
	return fmt.Sprintf("SelectorExpr(%v, %v)", s.X, s.SelectorExpr.Sel.Name)
}
