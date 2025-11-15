package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = StarExpr{}
var _ CanAssign = StarExpr{}

type StarExpr struct {
	StarPos token.Pos
	X       Expr
}

func (s StarExpr) Eval(vm *VM) {
	v := vm.frameStack.top().pop()
	// Check if this is a heap pointer
	if hp, ok := v.Interface().(*HeapPointer); ok {
		// Dereference from heap
		vm.pushOperand(vm.heap.read(hp))
		return
	}
	// (*int64)(nil)
	if v.Kind() == reflect.Func {
		if idn, ok := s.X.(Ident); ok {
			v = vm.localEnv().valueLookUp("*" + idn.Name)
			vm.pushOperand(v)
			return
		}
	}
	// Regular pointer dereference - validate it's a pointer
	if v.Kind() != reflect.Pointer {
		vm.fatal(fmt.Sprintf("cannot dereference non-pointer type: %v", v.Kind()))
	}
	if v.IsNil() {
		vm.fatal("cannot dereference nil pointer")
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
	// Check if this is a heap pointer
	if hp, ok := v.Interface().(*HeapPointer); ok {
		vm.heap.write(hp, value)
		return
	}
	if v.Kind() != reflect.Pointer {
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

func (s StarExpr) Pos() token.Pos { return s.StarPos }

func (s StarExpr) String() string {
	return fmt.Sprintf("StarExpr(%v)", s.X)
}

var _ Expr = ParenExpr{}

type ParenExpr struct {
	LParen token.Pos
	X      Expr
}

func (e ParenExpr) Eval(vm *VM) {} // noop
func (e ParenExpr) Flow(g *graphBuilder) (head Step) {
	return e.X.Flow(g)
}
func (e ParenExpr) Pos() token.Pos { return e.LParen }
func (e ParenExpr) String() string {
	return fmt.Sprintf("ParenExpr(%v)", e.X)
}
