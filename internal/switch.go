package internal

import (
	"fmt"
	"go/ast"
	"go/token"
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
		if head == nil {
			head = s.Tag.Flow(g)
		} else {
			_ = s.Tag.Flow(g)
		}
	}
	gotoLabel := fmt.Sprintf("switch-end-%d", g.idgen)
	gotoStep := g.newLabeledStep(gotoLabel)
	ref := statementReference{step: gotoStep} // has no ID
	g.funcStack.top().labelToStmt[gotoLabel] = ref

	for _, stmt := range s.Body.List {
		clause := stmt.(CaseClause)

		// check for default case
		if clause.List == nil {
			// compose goto to end of switch
			gotoEnd := BranchStmt{
				BranchStmt: &ast.BranchStmt{Tok: token.GOTO, TokPos: clause.Pos()},
				Label:      &Ident{Ident: &ast.Ident{Name: gotoLabel}},
			}
			list := append(clause.Body, gotoEnd)
			for i, stmt := range list {
				if i == 0 {
					first := stmt.Flow(g)
					if head == nil {
						head = first
					}
					continue
				}
				_ = stmt.Flow(g)
			}
			// switch clause ends here
			g.current = nil
			continue
		}

		// non-default case
		// non-default case
		if len(clause.Body) == 0 {
			continue
		}
		// compose condition
		var cond Expr
		// build a chain of OR expressions for each case expression
		for i, expr := range clause.List {
			var nextCond BinaryExpr
			if _, ok := expr.(BasicLit); ok {
				nextCond = BinaryExpr{
					BinaryExpr: &ast.BinaryExpr{Op: token.EQL, OpPos: clause.Pos()},
					X:          s.Tag,
					Y:          expr,
				}
			}
			if _, ok := expr.(Ident); ok {
				nextCond = BinaryExpr{
					BinaryExpr: &ast.BinaryExpr{Op: token.EQL, OpPos: clause.Pos()},
					X:          s.Tag,
					Y:          expr,
				}
			}
			if bin, ok := expr.(BinaryExpr); ok {
				nextCond = bin
			}
			if i == 0 {
				cond = nextCond
			} else {
				cond = BinaryExpr{
					BinaryExpr: &ast.BinaryExpr{Op: token.LOR, OpPos: clause.Pos()},
					X:          cond,
					Y:          nextCond,
				}
			}
		}

		// compose goto to end of switch
		gotoEnd := BranchStmt{
			BranchStmt: &ast.BranchStmt{Tok: token.GOTO, TokPos: clause.Pos()},
			Label:      &Ident{Ident: &ast.Ident{Name: gotoLabel}},
		}
		list := append(clause.Body, gotoEnd)

		// compose if statement for this case
		when := IfStmt{
			IfStmt: &ast.IfStmt{If: clause.Pos()},
			Cond:   cond,
			Body:   &BlockStmt{List: list},
		}
		whenFlow := when.Flow(g)
		if head == nil {
			head = whenFlow
		}
		// switch clause does not end here, can fallthrough
	}
	g.nextStep(gotoStep)
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
func (c CaseClause) Eval(vm *VM) {}

func (c CaseClause) Eval2(vm *VM) {
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
	// no flow for case clause itself
	return nil
}

type TypeSwitchStmt struct {
	*ast.TypeSwitchStmt
	Init   Stmt // initialization statement; or nil
	Assign Stmt // assignment statement; or nil
	Body   *BlockStmt
}

func (s TypeSwitchStmt) stmtStep() Evaluable { return s }

func (s TypeSwitchStmt) Eval(vm *VM) {}

func (s TypeSwitchStmt) String() string {
	return fmt.Sprintf("TypeSwitchStmt(%v,%v,%v)", s.Init, s.Assign, s.Body)
}

func (s TypeSwitchStmt) Flow(g *graphBuilder) (head Step) {
	if s.Init != nil {
		head = s.Init.Flow(g)
	}
	if s.Assign != nil {
		if head == nil {
			head = s.Assign.Flow(g)
		} else {
			_ = s.Assign.Flow(g)
		}
	}
	// body has CaseClauses, see SwitchStmt.Flow
	return head
}
