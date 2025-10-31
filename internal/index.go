package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = IndexExpr{}
var _ CanAssign = IndexExpr{}

type IndexExpr struct {
	*ast.IndexExpr
	X     Expr
	Index Expr
}

// func (i IndexExpr) EvalKind() reflect.Kind {
// 	return i.X.EvalKind()
// }

func (i IndexExpr) Eval(vm *VM) {
	var index, target reflect.Value
	if vm.isStepping {
		index = vm.frameStack.top().pop()
		target = vm.frameStack.top().pop()
	} else {
		index = vm.returnsEval(i.Index)
		target = vm.returnsEval(i.X)
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
	expected(target, "map or slice or array")
}

func (i IndexExpr) Assign(vm *VM, value reflect.Value) {
	target := vm.returnsEval(i.X)
	index := vm.returnsEval(i.Index)
	if target.Kind() == reflect.Map {
		target.SetMapIndex(index, value)
		return
	}
	if target.Kind() == reflect.Slice || target.Kind() == reflect.Array {
		reflect.ValueOf(target.Interface()).Index(int(index.Int())).Set(value)
		return
	}
	expected(target, "map or slice or array")
}

func (i IndexExpr) Define(vm *VM, value reflect.Value) {
	fmt.Println("IndexExpr.Define", i, value)
}

func (i IndexExpr) Flow(g *graphBuilder) (head Step) {
	head = i.X.Flow(g)
	i.Index.Flow(g)
	g.next(i)
	return head
}

func (i IndexExpr) String() string {
	return fmt.Sprintf("IndexExpr(%v, %v)", i.X, i.Index)
}
