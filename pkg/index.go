package pkg

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
	if len(vm.currentFrame.operands) == 0 {
		vm.eval(i.Index)
		vm.eval(i.X)
	}
	index := vm.popOperand()
	target := vm.popOperand()
	if target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	switch target.Kind() {
	case reflect.Map:
		v := target.MapIndex(index)
		vm.pushOperand(v)
	case reflect.Slice, reflect.Array:
		vm.pushOperand(target.Index(int(index.Int())))
	case reflect.String:
		v := reflect.ValueOf(target.String()[int(index.Int())])
		vm.pushOperand(v)
	default:
		vm.fatalf("expected string,map,slice or array, got %s", target.Kind())
	}
}

func (i IndexExpr) assign(vm *VM, value reflect.Value) {
	index := vm.popOperand()
	target := vm.popOperand()
	if target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	switch target.Kind() {
	case reflect.Map:
		target.SetMapIndex(index, value)
	case reflect.Slice, reflect.Array:
		target.Index(int(index.Int())).Set(value)
	default:
		vm.fatal("expected map or slice or array")
	}
}

func (i IndexExpr) define(vm *VM, value reflect.Value) {
	vm.fatalf("not yet implemented: IndexExpr.Define")
}

func (i IndexExpr) flow(g *graphBuilder) (head Step) {
	head = i.X.flow(g)
	i.Index.flow(g)
	g.next(i)
	return head
}

func (i IndexExpr) Pos() token.Pos { return i.Lbrack }

func (i IndexExpr) String() string {
	return fmt.Sprintf("IndexExpr(%v, %v)", i.X, i.Index)
}
