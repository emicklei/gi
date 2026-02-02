package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

type UnaryExprFunc func(reflect.Value) reflect.Value

var _ Expr = UnaryExpr{}

type UnaryExpr struct {
	opPos     token.Pos   // position of Op
	op        token.Token // operator
	x         Expr
	unaryFunc UnaryExprFunc // optional function to perform the unary operation
}

func (u UnaryExpr) Eval(vm *VM) {
	// TODO handle duplicate code (stack work)

	if u.unaryFunc != nil {
		v := vm.popOperand() // x value

		// propagate undeclared value. this happens when the expression is
		// used in a package variable or constant declaration
		if isUndeclared(v) {
			vm.pushOperand(v)
			return
		}

		result := u.unaryFunc(v)
		vm.pushOperand(result)
		return
	}

	// Special case: if taking address of an identifier, create a reference to the variable
	if u.op == token.AND {
		if ident, ok := u.x.(Ident); ok {
			// Pop the value that was already pushed by the identifier evaluation
			_ = vm.popOperand()
			// Create a heap pointer that references the environment variable
			env := vm.localEnv().valueOwnerOf(ident.Name)
			if env != nil {
				env.markSharedReferenced()
				value := env.valueLookUp(ident.Name)
				hp := vm.heap.allocHeapVar(env, ident.Name, value.Type())
				vm.pushOperand(reflect.ValueOf(hp))
				return
			}
		}
		if _, ok := u.x.(CompositeLit); ok {
			v := vm.popOperand()
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
			return
		}
		vm.fatal("UnaryExpr.Eval todo")
	}

	v := vm.popOperand()
	// propagate undeclared value. this happens when the expression is
	// used in a package variable or constant declaration
	if isUndeclared(v) {
		vm.pushOperand(v)
		return
	}
	switch v.Kind() {

	case reflect.Struct:
		switch u.op {
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		default:
			vm.fatal("missing unary operation on struct:" + u.op.String())
		}
	case reflect.Chan:
		switch u.op {
		case token.ARROW: // receive
			val, ok := v.Recv()
			if !ok {
				vm.pushOperand(reflect.Zero(v.Type()))
			} else {
				vm.pushOperand(val)
			}
		default:
			vm.fatal("missing unary operation on chan:" + u.op.String())
		}
	default:
		// Handle any other types (string, slice, map, etc.)
		switch u.op {
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		default:
			vm.fatal("not implemented: UnaryExpr.Eval:" + v.Kind().String())
		}
	}
}

func (u UnaryExpr) flow(g *graphBuilder) (head Step) {
	head = u.x.flow(g)
	g.next(u)
	return head
}

func (u UnaryExpr) String() string {
	return fmt.Sprintf("UnaryExpr(%s %s)", u.op, u.x)
}

func (u UnaryExpr) Pos() token.Pos { return u.opPos }
