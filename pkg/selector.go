package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = SelectorExpr{}
var _ CanAssign = SelectorExpr{}

type SelectorExpr struct {
	selector *Ident
	x        Expr
}

func (s SelectorExpr) define(vm *VM, val reflect.Value) {}

func (s SelectorExpr) assign(vm *VM, val reflect.Value) {
	if idn, ok := s.x.(Ident); ok {

		// need to pop from stack? TODO
		if trace {
			fmt.Println("TRACE: SelectorExpr.Assign", idn.name, s.selector.name, "=", val, "operands:", vm.currentFrame.operands)
		}

		recv := vm.localEnv().valueLookUp(idn.name)

		// dereference if pointer to heap value
		if hp, ok := recv.Interface().(*HeapPointer); ok {
			recv = vm.heap.read(hp)
		}
		// can we assign directly to the field?
		fa, ok := recv.Interface().(FieldAssignable)
		if ok {
			fa.fieldAssign(s.selector.name, val)
			return
		}

		vm.fatal(fmt.Sprintf("cannot assign to field %v for receiver: %v (%T)", s, recv.Interface(), recv.Interface()))
		return
	}
	recv := vm.returnsEval(s.x)

	// dereference if pointer to heap value
	if hp, ok := recv.Interface().(*HeapPointer); ok {
		recv = vm.heap.read(hp)
	}
	if !recv.IsValid() {
		vm.fatal("cannot assign to invalid selector receiver")
	}
	rec, ok := recv.Interface().(CanSelect)
	if ok {
		sel := rec.selectFieldOrMethod(s.selector.name)
		if !sel.IsValid() {
			vm.fatal(fmt.Sprintf("field %s not found for receiver: %v (%T)", s.selector.name, recv.Interface(), recv.Interface()))
		}
		if !sel.CanSet() {
			vm.fatal(fmt.Sprintf("field %s is not settable for receiver: %v (%T)", s.selector.name, recv.Interface(), recv.Interface()))
		}
		sel.Set(val)
		return
	}
	vm.fatal(fmt.Sprintf("cannot assign to method %s for receiver: %v (%T)", s.selector.name, recv.Interface(), recv.Interface()))
}

func (s SelectorExpr) Eval(vm *VM) {
	recv := vm.popOperand()
	// check for pointer to heap value
	if hp, ok := recv.Interface().(*HeapPointer); ok {
		recv = vm.heap.read(hp)
	}

	// interpreted receiver that can select fields or methods
	rec, ok := recv.Interface().(CanSelect)
	if ok {
		// can be field or method
		sel := rec.selectFieldOrMethod(s.selector.name)
		// check for method
		if _, ok := sel.Interface().(*FuncDecl); ok {
			// method value so push receiver as first argument
			vm.pushOperand(recv)
		}
		vm.pushOperand(sel)
		return
	}

	if recv.Kind() == reflect.Struct {
		field := recv.FieldByName(s.selector.name)
		if field.IsValid() {
			vm.pushOperand(field)
			return
		}
	}

	if recv.Kind() == reflect.Pointer {
		nonPtrRecv := recv.Elem()
		if nonPtrRecv.Kind() == reflect.Struct {
			field := nonPtrRecv.FieldByName(s.selector.name)
			if field.IsValid() {
				vm.pushOperand(field)
				return
			}
		}
	}

	meth := recv.MethodByName(s.selector.name)
	if meth.IsValid() {
		vm.pushOperand(meth)
		return
	}

	if isUndeclared(recv) {
		// propagate invalid value
		vm.pushOperand(recv)
		return
	}

	// Sel.Name is a method of receiver's pointer type ?
	recvType := recv.Type()
	ptrRecvType := reflect.PointerTo(recvType)
	pmeth, ok := ptrRecvType.MethodByName(s.selector.name)
	if ok {
		meth := reflect.ValueOf(pmeth)
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

	if ext, ok := recv.Interface().(ExtendedValue); ok {
		// *FuncDecl
		m, ok := ext.typ.methods[s.selector.name]
		if ok {
			// method value so push receiver as first argument
			vm.pushOperand(recv)
			vm.pushOperand(reflect.ValueOf(m))
			return
		}
	}

	vm.fatal(fmt.Sprintf("method or field \"%s\" not found for receiver: %v (%T)", s.selector.name, recv.Interface(), recv.Interface()))
}

func (s SelectorExpr) flow(g *graphBuilder) (head Step) {
	head = s.x.flow(g)
	g.next(s)
	return head
}

func (s SelectorExpr) Pos() token.Pos { return s.selector.namePos }

func (s SelectorExpr) String() string {
	return fmt.Sprintf("SelectorExpr(%v, %v)", s.x, s.selector.name)
}
