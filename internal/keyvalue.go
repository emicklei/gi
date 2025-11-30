package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = KeyValueExpr{}

type KeyValueExpr struct {
	Colon token.Pos // position of ":"
	Key   Expr
	Value Expr
}

func (e KeyValueExpr) Eval(vm *VM) {
	key := vm.callStack.top().pop()
	val := vm.callStack.top().pop()
	vm.pushOperand(reflect.ValueOf(keyValue{Key: key, Value: val}))
}

func (e KeyValueExpr) Flow(g *graphBuilder) (head Step) {
	// Value first so that key ends up on top of the stack
	head = e.Value.Flow(g)

	switch k := e.Key.(type) {
	case Ident:
		// use as string selector
		key := newBasicLit(e.Colon, reflect.ValueOf(k.Name))
		key.Flow(g)
	case BasicLit, Expr:
		e.Key.Flow(g)
	default:
		g.fatal(fmt.Sprintf("unhandled key type: %T", e.Key))
	}

	g.next(e)
	return head
}

func (e KeyValueExpr) Pos() token.Pos { return e.Colon }

func (e KeyValueExpr) String() string {
	return fmt.Sprintf("KeyValueExpr(%v,%v)", e.Key, e.Value)
}

type keyValue struct {
	Key   reflect.Value
	Value reflect.Value
}

func (k keyValue) String() string {
	return fmt.Sprintf("KeyValue(%v,%v)", k.Key, k.Value)
}
