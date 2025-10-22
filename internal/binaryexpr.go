package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var _ Flowable = BinaryExpr{}
var _ Expr = BinaryExpr{}

type BinaryExpr struct {
	*ast.BinaryExpr
	X Expr // left
	Y Expr // right
}

// func (s BinaryExpr) CanEval(vm *VM) bool {
// 	if vm.isStepping {
// 		right := vm.callStack.top().peek(0)
// 		left := vm.callStack.top().peek(1)
// 		return left.IsValid() && right.IsValid()
// 	}
// 	return s.X.CanEval(vm) && s.Y.CanEval(vm)
// }

func (s BinaryExpr) Eval(vm *VM) {
	var left, right reflect.Value
	if vm.isStepping {
		// see Flow for the order
		right = vm.callStack.top().pop()
		left = vm.callStack.top().pop()
	} else {
		left = vm.returnsEval(s.X)
		// propagate invalid value. this happens when the expression is
		// used in a package variable or constant declaration
		if !left.IsValid() {
			vm.pushOperand(left)
			return
		}
		right = vm.returnsEval(s.Y)
		// propagate invalid value. this happens when the expression is
		// used in a package variable or constant declaration
		if !right.IsValid() {
			vm.pushOperand(right)
			return
		}
	}
	v := BinaryExprValue{
		left:  left,
		op:    s.Op,
		right: right,
	}
	vm.pushOperand(v.Eval())
}

func (s BinaryExpr) Flow(g *graphBuilder) (head Step) {
	head = s.X.Flow(g)
	s.Y.Flow(g)
	g.next(s)
	return head
}

func (s BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr(%v %v %v)", s.X, s.Op, s.Y)
}

type BinaryExprValue struct {
	left  reflect.Value
	op    token.Token
	right reflect.Value
}

func (b BinaryExprValue) IsValid() bool {
	return b.left.IsValid() && b.right.IsValid()
}

func (b BinaryExprValue) Eval() reflect.Value {
	switch b.left.Kind() {
	case reflect.Int:
		res := b.IntEval(b.left.Int())
		if res.CanInt() {
			return reflect.ValueOf(int(res.Int()))
		} else {
			return res
		}
	case reflect.Uint:
		res := b.UIntEval(b.left.Uint())
		if res.CanInt() {
			return reflect.ValueOf(uint(res.Int()))
		} else {
			return res
		}
	case reflect.Int8:
		res := b.IntEval(b.left.Int())
		if res.CanInt() {
			return reflect.ValueOf(int8(res.Int()))
		} else {
			return res
		}
	case reflect.Uint8:
		res := b.UIntEval(b.left.Uint())
		if res.CanInt() {
			return reflect.ValueOf(uint8(res.Int()))
		} else {
			return res
		}
	case reflect.Int16:
		res := b.IntEval(b.left.Int())
		if res.CanInt() {
			return reflect.ValueOf(int16(res.Int()))
		} else {
			return res
		}
	case reflect.Uint16:
		res := b.UIntEval(b.left.Uint())
		if res.CanInt() {
			return reflect.ValueOf(uint16(res.Int()))
		} else {
			return res
		}
	case reflect.Int32:
		res := b.IntEval(b.left.Int())
		if res.CanInt() {
			return reflect.ValueOf(int32(res.Int()))
		} else {
			return res
		}
	case reflect.Uint32:
		res := b.UIntEval(b.left.Uint())
		if res.CanInt() {
			return reflect.ValueOf(uint32(res.Int()))
		} else {
			return res
		}
	case reflect.Int64:
		return b.IntEval(b.left.Int())
	case reflect.Uint64:
		return reflect.ValueOf(b.UIntEval(b.left.Uint()).Uint())
	// non-ints
	case reflect.Float64:
		return b.FloatEval(b.left.Float())
	case reflect.String:
		return b.StringEval(b.left.String())
	case reflect.Bool:
		return b.BoolEval(b.left.Bool())
	case reflect.Pointer:
		return b.PointerEval(b.left)
	}
	panic("not implemented: BinaryExprValue.Eval:" + b.left.Kind().String())
}

func (b BinaryExprValue) PointerEval(left reflect.Value) reflect.Value {
	switch b.op {
	case token.EQL:
		if left.Interface() == untypedNil && b.right.Interface() == untypedNil {
			return reflectTrue
		}
		if left.Interface() != b.right.Interface() {
			return reflectFalse
		}
		return reflectCondition(left.Pointer() == b.right.Pointer())
	case token.NEQ:
		if left.Interface() == untypedNil {
			return reflectCondition(b.right.Interface() != untypedNil)
		}
		if b.right.Interface() == untypedNil {
			return reflectCondition(left.Interface() != untypedNil)
		}
		// both non untypedNil
		return reflectCondition(left.Elem() != b.right.Elem())
	}
	panic("not implemented: BinaryExprValue.PointerEval:" + b.right.Kind().String())
}

func (b BinaryExprValue) BoolEval(left bool) reflect.Value {
	switch b.op {
	case token.LAND:
		return reflect.ValueOf(left && b.right.Bool())
	case token.LOR:
		return reflect.ValueOf(left || b.right.Bool())
	case token.EQL:
		return reflect.ValueOf(left == b.right.Bool())
	case token.NEQ:
		return reflect.ValueOf(left != b.right.Bool())
	}
	panic("not implemented: BinaryExprValue.BoolEval:" + b.right.Kind().String())
}

func (b BinaryExprValue) StringEval(left string) reflect.Value {
	switch b.op {
	case token.ADD:
		return reflect.ValueOf(left + b.right.String())
	case token.EQL:
		return reflect.ValueOf(left == b.right.String())
	case token.NEQ:
		return reflect.ValueOf(left != b.right.String())
	}
	panic("not implemented: BinaryExprValue.StringEval:" + b.right.Kind().String())
}

func (b BinaryExprValue) IntEval(left int64) reflect.Value {
	switch b.right.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return b.IntOpInt(left, b.right.Int())
	case reflect.Float64:
		return b.FloatOpFloat(float64(left), b.right.Float())
	case reflect.Complex128:
		return b.ComplexOpComplex(complex(float64(left), 0), b.right.Complex())
	}
	panic("not implemented: BinaryExprValue.IntEval:" + b.right.Kind().String())
}

func (b BinaryExprValue) UIntEval(left uint64) reflect.Value {
	switch b.right.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return b.UIntOpUInt(left, b.right.Uint())
	}
	panic("not implemented: BinaryExprValue.UIntEval:" + b.right.Kind().String())
}

func (b BinaryExprValue) FloatEval(left float64) reflect.Value {
	switch b.right.Kind() {
	case reflect.Float64:
		return b.FloatOpFloat(left, b.right.Float())
	case reflect.Int:
		return b.FloatOpFloat(left, float64(b.right.Int()))
	}
	panic("not implemented: BinaryExprValue.FloatEval:" + b.right.Kind().String())
}

func (b BinaryExprValue) FloatOpFloat(left float64, right float64) reflect.Value {
	switch b.op {
	case token.ADD:
		return reflect.ValueOf(left + right)
	case token.SUB:
		return reflect.ValueOf(left - right)
	case token.MUL:
		return reflect.ValueOf(left * right)
	case token.QUO:
		return reflect.ValueOf(left / right)
	}
	panic("not implemented: BinaryExprValue.FloatOpFloat:" + b.op.String())
}

func (b BinaryExprValue) ComplexOpComplex(left, right complex128) reflect.Value {
	switch b.op {
	case token.ADD:
		return reflect.ValueOf(left + right)
	}
	panic("not implemented: BinaryExprValue.ComplexOpComplex:" + b.op.String())
}

func (b BinaryExprValue) IntOpInt(left int64, right int64) reflect.Value {
	switch b.op {
	case token.ADD:
		return reflect.ValueOf(left + right)
	case token.SUB:
		return reflect.ValueOf(left - right)
	case token.MUL:
		return reflect.ValueOf(left * right)
	case token.QUO:
		return reflect.ValueOf(left / right)
	case token.REM:
		return reflect.ValueOf(left % right)
	case token.AND:
		return reflect.ValueOf(left & right)
	case token.OR:
		return reflect.ValueOf(left | right)
	case token.XOR:
		return reflect.ValueOf(left ^ right)
	case token.SHL:
		// right must be unsigned
		return reflect.ValueOf(left << uint64(right))
	case token.SHR:
		// right must be unsigned
		return reflect.ValueOf(left >> uint64(right))
	case token.AND_NOT:
		return reflect.ValueOf(left &^ right)
	case token.EQL:
		return reflect.ValueOf(left == right)
	case token.NEQ:
		return reflect.ValueOf(left != right)
	case token.LSS:
		return reflect.ValueOf(left < right)
	case token.LEQ:
		return reflect.ValueOf(left <= right)
	case token.GTR:
		return reflect.ValueOf(left > right)
	case token.GEQ:
		return reflect.ValueOf(left >= right)
	}
	panic("not implemented: BinaryExprValue.IntOpInt:" + b.op.String())
}

func (b BinaryExprValue) UIntOpUInt(left uint64, right uint64) reflect.Value {
	switch b.op {
	case token.ADD:
		return reflect.ValueOf(left + right)
	case token.SUB:
		return reflect.ValueOf(left - right)
	case token.MUL:
		return reflect.ValueOf(left * right)
	case token.QUO:
		return reflect.ValueOf(left / right)
	case token.REM:
		return reflect.ValueOf(left % right)
	case token.AND:
		return reflect.ValueOf(left & right)
	case token.OR:
		return reflect.ValueOf(left | right)
	case token.XOR:
		return reflect.ValueOf(left ^ right)
	case token.SHL:
		return reflect.ValueOf(left << right)
	case token.SHR:
		return reflect.ValueOf(left >> right)
	case token.AND_NOT:
		return reflect.ValueOf(left &^ right)
	case token.EQL:
		return reflect.ValueOf(left == right)
	case token.NEQ:
		return reflect.ValueOf(left != right)
	case token.LSS:
		return reflect.ValueOf(left < right)
	case token.LEQ:
		return reflect.ValueOf(left <= right)
	case token.GTR:
		return reflect.ValueOf(left > right)
	case token.GEQ:
		return reflect.ValueOf(left >= right)
	}
	panic("not implemented: BinaryExprValue.UIntOpUInt:" + b.op.String())
}
