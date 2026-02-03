package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = StarExpr{}
var _ CanAssign = StarExpr{}

type StarExpr struct {
	starPos token.Pos
	x       Expr
}

func (s StarExpr) Eval(vm *VM) {
	v := vm.popOperand()
	// Check if this is a heap pointer
	if hp, ok := v.Interface().(*HeapPointer); ok {
		// Dereference from heap
		vm.pushOperand(vm.heap.read(hp))
		return
	}
	// needed?
	if v.Kind() == reflect.Func {
		if idn, ok := s.x.(Ident); ok {
			v = vm.localEnv().valueLookUp("*" + idn.name)
			vm.pushOperand(v)
			return
		}
	}
	// (*int64)(nil)
	if v.Kind() == reflect.Struct {
		if _, ok := v.Interface().(builtinType); ok {
			vm.pushOperand(v)
			return
		} else if _, ok := v.Interface().(ExtendedType); ok {
			vm.pushOperand(v)
			return
		} else {
			vm.fatal(fmt.Sprintf("unhandled struct type: %v (%T)", v.Interface(), v.Interface()))
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
func (s StarExpr) flow(g *graphBuilder) (head Step) {
	head = s.x.flow(g)
	g.next(s)
	return
}

func (s StarExpr) assign(vm *VM, value reflect.Value) {
	v := vm.returnsEval(s.x)
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
	uv := v.Elem()
	uv.Set(value.Convert(uv.Type()))
}

func (s StarExpr) define(vm *VM, value reflect.Value) {
	// Define through a pointer doesn't make sense in Go
	// This would be like *p := value, which is invalid syntax
	vm.fatal("cannot use := with pointer dereference")
}

func (s StarExpr) Pos() token.Pos { return s.starPos }

func (s StarExpr) String() string {
	return fmt.Sprintf("StarExpr(%v)", s.x)
}

var _ Expr = ParenExpr{}

type ParenExpr struct {
	LParen token.Pos
	X      Expr
}

func (e ParenExpr) Eval(vm *VM) {} // noop
func (e ParenExpr) flow(g *graphBuilder) (head Step) {
	return e.X.flow(g)
}
func (e ParenExpr) Pos() token.Pos { return e.LParen }
func (e ParenExpr) String() string {
	return fmt.Sprintf("ParenExpr(%v)", e.X)
}
