package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Stmt = SwitchStmt{}

// A SwitchStmt represents an expression switch statement.
type SwitchStmt struct {
	*ast.SwitchStmt
	Init Stmt // initialization statement; or nil
	Tag  Expr // tag expression; or nil
	Body BlockStmt
}

func (s SwitchStmt) stmtStep() Evaluable { return s }

func (s SwitchStmt) Eval(vm *VM) {
	vm.pushNewFrame(s)
	defer vm.popFrame() // to handle break statements
	if trace {
		if s.Init != nil {
			vm.traceEval(s.Init.stmtStep())
		}
		if s.Tag != nil {
			vm.traceEval(s.Tag)
		}
		vm.traceEval(s.Body)
	} else {
		if s.Init != nil {
			s.Init.stmtStep().Eval(vm)
		}
		if s.Tag != nil {
			s.Tag.Eval(vm)
		}
		s.Body.Eval(vm)
	}
}
func (s SwitchStmt) String() string {
	return fmt.Sprintf("SwitchStmt(%v,%v,%v)", s.Init, s.Tag, s.Body)
}

func (s SwitchStmt) Flow(g *graphBuilder) (head Step) {
	head = new(pushStackFrameStep)
	g.nextStep(head)
	if s.Init != nil {
		s.Init.Flow(g)
	}
	if s.Tag != nil {
		s.Tag.Flow(g)
	}
	s.Body.Flow(g)
	g.nextStep(new(popStackFrameStep))
	return head
}

var _ Flowable = CaseClause{}

// A CaseClause represents a case of an expression or type switch statement.
type CaseClause struct {
	*ast.CaseClause
	List []Expr // list of expressions; nil means default case
	Body []Stmt
}

func (c CaseClause) stmtStep() Evaluable { return c }

func (c CaseClause) String() string {
	return fmt.Sprintf("CaseClause(%v,%v)", c.List, c.Body)
}
func (c CaseClause) Eval(vm *VM) {
	if c.List == nil {
		// default case
		for _, stmt := range c.Body {
			if trace {
				vm.traceEval(stmt.stmtStep())
			} else {
				stmt.stmtStep().Eval(vm)
			}
		}
		return
	}
	f := vm.frameStack.top()
	var left reflect.Value
	if len(f.operandStack) != 0 {
		left = vm.frameStack.top().pop()
	}
	for _, expr := range c.List {
		right := vm.returnsEval(expr)
		var cond bool
		if left.IsValid() {
			// because value is on the operand stack we compare
			cond = left.Equal(right)
		} else {
			// no operand on stack, treat as boolean expression
			cond = right.Bool()
		}
		if cond {
			vm.pushNewFrame(c)
			defer vm.popFrame()
			for _, stmt := range c.Body {
				if trace {
					vm.traceEval(stmt.stmtStep())
				} else {
					stmt.stmtStep().Eval(vm)
				}
			}
			return
		}
	}
}

func (c CaseClause) Flow(g *graphBuilder) (head Step) {
	return g.current // TODO
}
