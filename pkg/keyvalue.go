package pkg

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

func (k KeyValueExpr) Eval(vm *VM) {
	key := vm.popOperand()
	val := vm.popOperand()
	vm.pushOperand(reflect.ValueOf(keyValue{Key: key, Value: val}))
}

func (k KeyValueExpr) Flow(g *graphBuilder) (head Step) {
	// Value first so that key ends up on top of the stack
	head = k.Value.Flow(g)

	key := k.Key
	if id, ok := key.(Ident); ok {
		// wrap so that we can evaluate it properly
		key = identKey{Ident: id}
	}
	key.Flow(g)

	g.next(k)
	return head
}

func (k KeyValueExpr) Pos() token.Pos { return k.Colon }

func (k KeyValueExpr) String() string {
	return fmt.Sprintf("KeyValueExpr(%v,%v)", k.Key, k.Value)
}

// identKey is a wrapper to evaluate struct keys that are selectors.
type identKey struct {
	Ident Ident
}

func (i identKey) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(i.Ident))
}
func (i identKey) Flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}
func (i identKey) Pos() token.Pos { return i.Ident.Pos() }

func (i identKey) String() string {
	return fmt.Sprintf("identKey(%v)", i.Ident)
}

type keyValue struct {
	Key   reflect.Value
	Value reflect.Value
}

func (k keyValue) String() string {
	return fmt.Sprintf("KeyValue(%v,%v)", k.Key, k.Value)
}
