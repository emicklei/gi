package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Stmt = AssignStmt{}

type AssignStmt struct {
	TokPos      token.Pos   // position of Tok
	Tok         token.Token // assignment token, DEFINE
	Lhs         []Expr
	Rhs         []Expr
	rhsBinFuncs []BinaryExprFunc // computed at build time
}

func (a AssignStmt) Eval(vm *VM) {
	var lastVal reflect.Value
	for i := 0; i < len(a.Lhs); i++ {
		each := a.Lhs[i]
		var v reflect.Value
		// handle "ok" idiom for map index expressions
		if len(vm.callStack.top().operands) == 0 {
			if !lastVal.IsValid() {
				panic("internal error: missing value for assignment")
			}
			v = reflect.ValueOf(!lastVal.IsZero())
		} else {
			v = vm.callStack.top().pop()
			lastVal = v
		}
		a.apply(each, vm, v)
	}
	if len(a.Lhs) < len(a.Rhs) {
		_ = vm.callStack.top().pop()
	}
}

func (a AssignStmt) apply(each Expr, vm *VM, v reflect.Value) {
	target, ok_ := each.(CanAssign)
	if !ok_ {
		vm.fatal(fmt.Sprintf("cannot assign %v to a %T", v.Interface(), each))
	}
	switch a.Tok {
	case token.DEFINE: // :=
		target.Define(vm, v)
	case token.ASSIGN: // =
		target.Assign(vm, v)
	case token.ADD_ASSIGN: // +=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.ADD, right: v}.Eval()
		target.Assign(vm, result)
	case token.SUB_ASSIGN: // -=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.SUB, right: v}.Eval()
		target.Assign(vm, result)
	case token.MUL_ASSIGN: // *=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.MUL, right: v}.Eval()
		target.Assign(vm, result)
	case token.QUO_ASSIGN: // /=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.QUO, right: v}.Eval()
		target.Assign(vm, result)
	case token.REM_ASSIGN: // %=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.REM, right: v}.Eval()
		target.Assign(vm, result)
	case token.AND_ASSIGN: // &=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.AND, right: v}.Eval()
		target.Assign(vm, result)
	case token.OR_ASSIGN: // |=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.OR, right: v}.Eval()
		target.Assign(vm, result)
	case token.XOR_ASSIGN: // ^=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.XOR, right: v}.Eval()
		target.Assign(vm, result)
	case token.SHL_ASSIGN: // <<=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.SHL, right: v}.Eval()
		target.Assign(vm, result)
	case token.SHR_ASSIGN: // >>=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.SHR, right: v}.Eval()
		target.Assign(vm, result)
	case token.AND_NOT_ASSIGN: // &^=
		current := vm.returnsEval(each)
		result := BinaryExprValue{left: current, op: token.AND_NOT, right: v}.Eval()
		target.Assign(vm, result)
	default:
		panic("unsupported assignment " + a.Tok.String())
	}
}

// pairwise flow
func (a AssignStmt) Flow(g *graphBuilder) (head Step) {
	for i := len(a.Lhs) - 1; i >= 0; i-- {
		left := a.Lhs[i]
		left.Flow(g)
		// step back to previous, the last node must not be evaluated
		g.stepBack()
		if head == nil && g.current != nil {
			head = g.current
		}
		// right side may be shorter (e.g. x, y = f())
		if i < len(a.Rhs) {
			right := a.Rhs[i]
			rightFlow := right.Flow(g)
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

func (a AssignStmt) Pos() token.Pos { return a.TokPos }

func (a AssignStmt) stmtStep() Evaluable { return a }

func (a AssignStmt) String() string {
	return fmt.Sprintf("AssignStmt(%v %s %v)", a.Lhs, a.Tok, a.Rhs)
}
