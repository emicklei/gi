package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = UnaryExpr{}

type UnaryExpr struct {
	OpPos token.Pos   // position of Op
	Op    token.Token // operator
	X     Expr
}

func (u UnaryExpr) Eval(vm *VM) {
	// Special case: if taking address of an identifier, create a reference to the variable
	if u.Op == token.AND {
		if ident, ok := u.X.(Ident); ok {
			// Pop the value that was already pushed by the identifier evaluation
			_ = vm.callStack.top().pop()
			// Create a heap pointer that references the environment variable
			env := vm.localEnv().valueOwnerOf(ident.Name)
			if env != nil {
				env.markContainsHeapPointer()
				value := env.valueLookUp(ident.Name)
				hp := vm.heap.allocHeapVar(env, ident.Name, value.Type())
				vm.pushOperand(reflect.ValueOf(hp))
				return
			}
		}
		if _, ok := u.X.(CompositeLit); ok {
			v := vm.callStack.top().pop()
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
			return
		}
		vm.fatal("UnaryExpr.Eval todo")
	}

	v := vm.callStack.top().pop()
	// propagate undeclared value. this happens when the expression is
	// used in a package variable or constant declaration
	if v == reflectUndeclared {
		vm.pushOperand(v)
		return
	}
	switch v.Kind() {
	case reflect.Bool:
		switch u.Op {
		case token.NOT:
			vm.pushOperand(reflect.ValueOf(!v.Bool()))
		default:
			vm.fatal("missing unary operation on bool:" + u.Op.String())
		}
	case reflect.Int:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(int(-v.Int())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(int(^v.Int())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(int(v.Int())))
		default:
			vm.fatal("missing unary operation on int:" + u.Op.String())
		}
	case reflect.Int8:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(int8(-v.Int())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(int8(^v.Int())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(int8(v.Int())))
		default:
			vm.fatal("missing unary operation on int8:" + u.Op.String())
		}
	case reflect.Int16:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(int16(-v.Int())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(int16(^v.Int())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(int16(v.Int())))
		default:
			vm.fatal("missing unary operation on int16:" + u.Op.String())
		}
	case reflect.Int32:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(int32(-v.Int())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(int32(^v.Int())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(int32(v.Int())))
		default:
			vm.fatal("missing unary operation on int32:" + u.Op.String())
		}
	case reflect.Int64:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(-v.Int()))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(^v.Int()))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(int32(v.Int())))
		default:
			vm.fatal("missing unary operation on int64:" + u.Op.String())
		}
	case reflect.Uint64:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(uint64(-v.Uint())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(uint64(^v.Uint())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(uint64(v.Uint())))
		default:
			vm.fatal("missing unary operation on uint64:" + u.Op.String())
		}
	case reflect.Uint32:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(uint32(-v.Uint())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(uint32(^v.Uint())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(uint32(v.Uint())))
		default:
			vm.fatal("missing unary operation on uint32:" + u.Op.String())
		}
	case reflect.Uint16:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(uint16(-v.Uint())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(uint16(^v.Uint())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(uint16(v.Uint())))
		default:
			vm.fatal("missing unary operation on uint16:" + u.Op.String())
		}
	case reflect.Uint8:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(uint8(-v.Uint())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(uint8(^v.Uint())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(uint8(v.Uint())))
		default:
			vm.fatal("missing unary operation on uint8:" + u.Op.String())
		}
	case reflect.Uint:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(uint(-v.Uint())))
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		case token.XOR:
			vm.pushOperand(reflect.ValueOf(uint(^v.Uint())))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(uint(v.Uint())))
		default:
			vm.fatal("missing unary operation on uint:" + u.Op.String())
		}
	case reflect.Float64:
		switch u.Op {
		case token.SUB:
			vm.pushOperand(reflect.ValueOf(-v.Float()))
		case token.ADD:
			vm.pushOperand(reflect.ValueOf(v.Float()))
		default:
			vm.fatal("missing unary operation on float64:" + u.Op.String())
		}
	case reflect.Struct:
		switch u.Op {
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		default:
			vm.fatal("missing unary operation on struct:" + u.Op.String())
		}
	case reflect.Chan:
		switch u.Op {
		case token.ARROW: // receive
			val, ok := v.Recv()
			if !ok {
				vm.pushOperand(reflect.Zero(v.Type()))
			} else {
				vm.pushOperand(val)
			}
		default:
			vm.fatal("missing unary operation on chan:" + u.Op.String())
		}
	default:
		// Handle any other types (string, slice, map, etc.)
		switch u.Op {
		case token.AND:
			hp := vm.heap.allocHeapValue(v)
			vm.pushOperand(reflect.ValueOf(hp))
		default:
			vm.fatal("not implemented: UnaryExpr.Eval:" + v.Kind().String())
		}
	}
}

func (u UnaryExpr) Flow(g *graphBuilder) (head Step) {
	head = u.X.Flow(g)
	g.next(u)
	return head
}

func (u UnaryExpr) String() string {
	return fmt.Sprintf("UnaryExpr(%s %s)", u.Op, u.X)
}

func (u UnaryExpr) Pos() token.Pos { return u.OpPos }
