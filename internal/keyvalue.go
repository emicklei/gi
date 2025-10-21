package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = KeyValueExpr{}

type KeyValueExpr struct {
	*ast.KeyValueExpr
	Key   Expr
	Value Expr
}

func (e KeyValueExpr) String() string {
	return fmt.Sprintf("KeyValueExpr(%v,%v)", e.Key, e.Value)
}

func (e KeyValueExpr) Eval(vm *VM) {
	var key reflect.Value
	if vm.isStepping {
		key = vm.callStack.top().pop()
	} else {
		switch k := e.Key.(type) {
		case Ident:
			key = reflect.ValueOf(k.Name)
		case BasicLit:
			key = vm.returnsEval(k)
		default:
			vm.fatal("unhandled key type:" + fmt.Sprintf("%T", e.Key))
		}
	}
	var val reflect.Value
	if vm.isStepping {
		val = vm.callStack.top().pop()
	} else {
		val = vm.returnsEval(e.Value)
	}
	vm.pushOperand(reflect.ValueOf(KeyValue{Key: key, Value: val}))
}

func (e KeyValueExpr) Flow(g *graphBuilder) (head Step) {
	// Value first so that it ends up on top of the stack
	head = e.Value.Flow(g)
	e.Key.Flow(g)
	g.next(e)
	return head
}

type KeyValue struct {
	Key   reflect.Value
	Value reflect.Value
}

func (k KeyValue) String() string {
	return fmt.Sprintf("KeyValue(%v,%v)", k.Key, k.Value)
}
