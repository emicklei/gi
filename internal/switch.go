package internal

import (
	"fmt"
	"go/ast"
	"reflect"

	"github.com/emicklei/dot"
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
	vm.frameStack.top().pushEnv()
	defer vm.frameStack.top().popEnv()
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
	if s.Init != nil {
		head = s.Init.Flow(g)
	}
	if s.Tag != nil {
		s.Tag.Flow(g)
		if head == nil {
			head = g.current
		}
	}
	s.Body.Flow(g)
	if head == nil {
		head = g.current
	}
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
			vm.frameStack.top().pushEnv()
			defer vm.frameStack.top().popEnv()
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
	if c.List != nil {
		for i, expr := range c.List {
			expr.Flow(g)
			if i == 0 {
				head = g.current
			}
			cs := new(caseStep)
			// create detached body flow
			bg := newGraphBuilder(g.goPkg)
			for _, stmt := range c.Body {
				stmt.Flow(bg)
			}
			cs.bodyFlow = bg.head
			g.nextStep(cs)
		}
	} else {
		// default case has no expressions
		for _, stmt := range c.Body {
			stmt.Flow(g)
		}
	}
	if head == nil {
		head = g.current
	}
	// each case clause ends the switch
	g.current = nil
	return
}

type caseStep struct {
	step
	exprFlow Step
	bodyFlow Step
}

func (s *caseStep) Take(vm *VM) Step {
	f := vm.frameStack.top()
	var left reflect.Value
	// is there a tag value on the operand stack?
	if len(f.operandStack) != 0 {
		left = vm.frameStack.top().pop()
	}
	if left.IsValid() {
		if left.Kind() == reflect.Bool {
			if left.Bool() {
				return s.bodyFlow
			}
		} else {
			right := vm.frameStack.top().pop()
			if left.Equal(right) {
				return s.bodyFlow
			}
		}
	}
	return s.next
}

func (s *caseStep) String() string {
	return fmt.Sprintf("%2d: case", s.id)
}

func (s *caseStep) Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node {
	me := s.step.traverse(g, s.String(), "case", visited)
	expr := s.exprFlow.Traverse(g, visited)
	me.Edge(expr, "expr")
	body := s.bodyFlow.Traverse(g, visited)
	me.Edge(body, "body")
	return me
}
