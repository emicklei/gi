package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

type BinaryExprFunc func(x, y reflect.Value) reflect.Value

var _ Flowable = BinaryExpr2{}

type BinaryExpr2 struct {
	OpPos      token.Pos
	Op         token.Token
	X          Expr // left
	Y          Expr // right
	binaryFunc BinaryExprFunc
}

func (b BinaryExpr2) Flow(g *graphBuilder) (head Step) {
	head = b.X.Flow(g)
	b.Y.Flow(g)
	g.next(b)
	return head
}

func (b BinaryExpr2) Eval(vm *VM) {
	y := vm.popOperand()
	if y == reflectUndeclared {
		vm.pushOperand(y)
		return
	}
	x := vm.popOperand()
	if x == reflectUndeclared {
		vm.pushOperand(x)
		return
	}
	vm.pushOperand(b.binaryFunc(x, y))
}

func (b BinaryExpr2) Pos() token.Pos {
	return b.OpPos
}

func (b BinaryExpr2) String() string {
	return fmt.Sprintf("BinaryExpr2(%v,%v,%v)", b.X, b.Op, b.Y)
}
