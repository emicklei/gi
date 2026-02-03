package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Stmt = AssignStmt{}

type AssignStmt struct {
	tokPos      token.Pos   // position of Tok
	tok         token.Token // assignment token, DEFINE
	lhs         []Expr
	rhs         []Expr
	rhsBinFuncs []BinaryExprFunc // computed at build time
}

func (a AssignStmt) Eval(vm *VM) {
	var lastVal reflect.Value
	for i := 0; i < len(a.lhs); i++ {
		each := a.lhs[i]
		var v reflect.Value
		// handle "ok" idiom for map index expressions
		if len(vm.currentFrame.operands) == 0 {
			if !lastVal.IsValid() {
				panic("internal error: missing value for assignment")
			}
			v = reflect.ValueOf(!lastVal.IsZero())
		} else {
			v = vm.popOperand()
			lastVal = v
		}
		a.apply(each, vm, v)
	}
	if len(a.lhs) < len(a.rhs) {
		_ = vm.popOperand()
	}
}

func (a AssignStmt) apply(each Expr, vm *VM, v reflect.Value) {
	target, ok_ := each.(CanAssign)
	if !ok_ {
		vm.fatal(fmt.Sprintf("cannot assign %v to a %T", v.Interface(), each))
	}
	switch a.tok {
	case token.DEFINE: // :=
		target.define(vm, v)
	case token.ASSIGN: // =
		target.assign(vm, v)
	case token.ADD_ASSIGN: // +=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.ADD, right: v}.eval()
		target.assign(vm, result)
	case token.SUB_ASSIGN: // -=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.SUB, right: v}.eval()
		target.assign(vm, result)
	case token.MUL_ASSIGN: // *=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.MUL, right: v}.eval()
		target.assign(vm, result)
	case token.QUO_ASSIGN: // /=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.QUO, right: v}.eval()
		target.assign(vm, result)
	case token.REM_ASSIGN: // %=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.REM, right: v}.eval()
		target.assign(vm, result)
	case token.AND_ASSIGN: // &=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.AND, right: v}.eval()
		target.assign(vm, result)
	case token.OR_ASSIGN: // |=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.OR, right: v}.eval()
		target.assign(vm, result)
	case token.XOR_ASSIGN: // ^=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.XOR, right: v}.eval()
		target.assign(vm, result)
	case token.SHL_ASSIGN: // <<=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.SHL, right: v}.eval()
		target.assign(vm, result)
	case token.SHR_ASSIGN: // >>=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.SHR, right: v}.eval()
		target.assign(vm, result)
	case token.AND_NOT_ASSIGN: // &^=
		current := vm.returnsEval(each)
		result := binaryExprValue{left: current, op: token.AND_NOT, right: v}.eval()
		target.assign(vm, result)
	default:
		panic("unsupported assignment " + a.tok.String())
	}
}

// pairwise flow
func (a AssignStmt) flow(g *graphBuilder) (head Step) {
	for i := len(a.lhs) - 1; i >= 0; i-- {
		left := a.lhs[i]
		left.flow(g)
		// step back to previous, the last node must not be evaluated
		g.stepBack()
		if head == nil && g.current != nil {
			head = g.current
		}
		// right side may be shorter (e.g. x, y = f())
		if i < len(a.rhs) {
			right := a.rhs[i]
			rightFlow := right.flow(g)
			if head == nil {
				head = rightFlow
			}
		}
	}
	g.next(a)
	if head == nil {
		head = g.current
	}
	return head
}

func (a AssignStmt) Pos() token.Pos { return a.tokPos }

func (a AssignStmt) stmtStep() Evaluable { return a }

func (a AssignStmt) String() string {
	return fmt.Sprintf("AssignStmt(%v %s %v)", a.lhs, a.tok, a.rhs)
}
