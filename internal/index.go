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
	index := vm.frameStack.top().pop()
	target := vm.frameStack.top().pop()
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

func (i IndexExpr) Assign(vm *VM, value reflect.Value) {
	target := vm.returnsEval(i.X)
	index := vm.returnsEval(i.Index)
	if target.Kind() == reflect.Map {
		target.SetMapIndex(index, value)
		return
	}
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
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
