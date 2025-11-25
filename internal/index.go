package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = IndexExpr{}
var _ CanAssign = IndexExpr{}

type IndexExpr struct {
	Lbrack token.Pos // position of "["
	X      Expr
	Index  Expr
}

func (i IndexExpr) Eval(vm *VM) {
	if len(vm.callStack.top().operands) == 0 {
		vm.eval(i.Index)
		vm.eval(i.X)
	}
	index := vm.callStack.top().pop()
	target := vm.callStack.top().pop()
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}
	if target.Kind() == reflect.Map {
		v := target.MapIndex(index)
		vm.pushOperand(v)
		return
	}
	if target.Kind() == reflect.Slice || target.Kind() == reflect.Array {
		vm.pushOperand(target.Index(int(index.Int())))
		return
	}
	if target.Kind() == reflect.String {
		v := reflect.ValueOf(target.String()[int(index.Int())])
		vm.pushOperand(v)
		return
	}
	vm.fatal(fmt.Sprintf("expected map or slice or array, got %s", target.Kind()))
}

// assign:
// to variable
// to map index
// to slice or array index
// to struct field

type variableAssigner struct {
	ident Ident
}
type mapIndexAssigner struct {
	container reflect.Value
	key       reflect.Value
}
type sliceOrArrayIndexAssigner struct {
	container reflect.Value
	index     int
}
type structFieldAssigner struct {
	container reflect.Value
	fieldName string
}

// find the object to assign the value to

// func (vm *VM) findAssigner(e Expr) CanAssign {
// 	if ident, ok := e.(Ident); ok {
// 		return ident
// 	}
// }

func (i IndexExpr) Assign(vm *VM, value reflect.Value) {
	vm.printStack()
	index := vm.callStack.top().pop()
	target := vm.callStack.top().pop()
	if target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	if target.Kind() == reflect.Map {
		target.SetMapIndex(index, value)
		return
	}
	if target.Kind() == reflect.Slice || target.Kind() == reflect.Array {
		target.Index(int(index.Int())).Set(value)
		return
	}
	expected(target, "map or slice or array")
}

func (i IndexExpr) Define(vm *VM, value reflect.Value) {
	vm.fatal("not yet implemented: IndexExpr.Define")
}

func (i IndexExpr) Flow(g *graphBuilder) (head Step) {
	head = i.X.Flow(g)
	i.Index.Flow(g)
	g.next(i)
	return head
}

func (i IndexExpr) Pos() token.Pos { return i.Lbrack }

func (i IndexExpr) String() string {
	return fmt.Sprintf("IndexExpr(%v, %v)", i.X, i.Index)
}
