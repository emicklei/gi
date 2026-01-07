package pkg

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
	vm.currentFrame.pushEnv()
	defer vm.currentFrame.popEnv()
	if s.Init != nil {
		vm.eval(s.Init.stmtStep())
	}
	if s.Tag != nil {
		vm.eval(s.Tag)
	}
	vm.eval(s.Body)
}
func (s SwitchStmt) String() string {
	return fmt.Sprintf("SwitchStmt(%v,%v,%v)", s.Init, s.Tag, s.Body)
}

func (s SwitchStmt) flow(g *graphBuilder) (head Step) {
	if s.Init != nil {
		head = s.Init.flow(g)
	}
	if s.Tag != nil {
		if head == nil {
			head = s.Tag.flow(g)
		} else {
			_ = s.Tag.flow(g)
		}
	}
	gotoLabel := fmt.Sprintf("switch-end-%d", g.idgen)
	gotoStep := g.newLabeledStep(gotoLabel, s.Pos())
	ref := statementReference{step: gotoStep} // has no ID
	g.funcStack.top().putGotoReference(gotoLabel, ref)

	for _, stmt := range s.Body.List {
		clause := stmt.(CaseClause)

		// check for default case
		if clause.List == nil {
			// compose goto to end of switch
			labelIdent := Ident{Name: gotoLabel}
			gotoEnd := BranchStmt{
				Tok:    token.GOTO,
				TokPos: clause.Pos(),
				Label:  &labelIdent,
			}
			list := append(clause.Body, gotoEnd)
			for i, stmt := range list {
				if i == 0 {
					first := stmt.flow(g)
					if head == nil {
						head = first
					}
					continue
				}
				_ = stmt.flow(g)
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
			var nextCond Expr
			if s.Tag != nil {
				nextCond = BinaryExpr{
					Op:    token.EQL,
					OpPos: clause.Pos(),
					X:     s.Tag,
					Y:     expr,
				}
			}
			if bin, ok := expr.(BinaryExpr); ok {
				nextCond = bin
			}
			if nextCond == nil {
				nextCond = expr
			}
			if i == 0 {
				cond = nextCond
			} else {
				cond = BinaryExpr{
					Op:    token.LOR,
					OpPos: clause.Pos(),
					X:     cond,
					Y:     nextCond,
				}
			}
		}

		// compose goto to end of switch
		labelIdent := Ident{Name: gotoLabel}
		gotoEnd := BranchStmt{
			Tok:    token.GOTO,
			TokPos: clause.Pos(),
			Label:  &labelIdent,
		}
		list := append(clause.Body, gotoEnd)

		// compose if statement for this case
		when := IfStmt{
			IfPos: clause.Pos(),
			Cond:  cond,
			Body:  &BlockStmt{List: list},
		}
		whenFlow := when.flow(g)
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
	CasePos token.Pos // position of "case" or "default" keyword
	List    []Expr    // list of expressions; nil means default case
	Body    []Stmt
}

func (c CaseClause) Eval(vm *VM) {}

func (c CaseClause) flow(g *graphBuilder) (head Step) {
	// no flow for case clause itself
	return nil
}

func (c CaseClause) Pos() token.Pos { return c.CasePos }

func (c CaseClause) stmtStep() Evaluable { return c }

func (c CaseClause) String() string {
	return fmt.Sprintf("CaseClause(%v,%v)", c.List, c.Body)
}

type TypeSwitchStmt struct {
	SwitchPos token.Pos
	Init      Stmt // initialization statement; or nil
	Assign    Stmt // x := y.(type) or y.(type)
	Body      *BlockStmt
}

func (s TypeSwitchStmt) Eval(vm *VM) {}

func (s TypeSwitchStmt) flow(g *graphBuilder) (head Step) {
	if s.Init != nil {
		head = s.Init.flow(g)
	}
	if s.Assign != nil {
		assignFlow := s.Assign.flow(g)
		if head == nil {
			head = assignFlow
		}
		// if Assign is not an assignment statement, we need to pop the value from stack
		// that was pushed by the TypeAssertExpr in the Assign expression
		if _, ok := s.Assign.(AssignStmt); !ok {
			g.nextStep(new(popOperandStep))
		}
	}
	gotoLabel := fmt.Sprintf("type-switch-end-%d", g.idgen)
	gotoStep := g.newLabeledStep(gotoLabel, s.Pos())
	ref := statementReference{step: gotoStep} // has no ID
	g.funcStack.top().putGotoReference(gotoLabel, ref)

	nameOfType := Ident{NamePos: s.Pos(), Name: internalVarName("switch-type-name", g.idgen)}

	nameOfTypeAssignment := AssignStmt{
		TokPos: s.SwitchPos,
		Tok:    token.DEFINE,
		Lhs:    []Expr{nameOfType},
		Rhs:    []Expr{noExpr{}}, // no expression because TypeAssertExpr pushes two values
	}
	nameOfTypeAssignment.flow(g)

	for _, stmt := range s.Body.List {
		clause := stmt.(CaseClause)

		// check for default case
		if clause.List == nil {
			// compose goto to end of switch
			labelIdent := Ident{Name: gotoLabel}
			gotoEnd := BranchStmt{
				Tok:    token.GOTO,
				TokPos: clause.Pos(),
				Label:  &labelIdent,
			}
			list := append(clause.Body, gotoEnd)
			for i, stmt := range list {
				if i == 0 {
					first := stmt.flow(g)
					if head == nil {
						head = first
					}
					continue
				}
				_ = stmt.flow(g)
			}
			// switch clause ends here
			g.current = nil
			continue
		}

		// non-default case
		if len(clause.Body) == 0 {
			continue
		}
		// compose condition
		var cond Expr
		// build a chain of OR expressions for each case expression
		for i, expr := range clause.List {
			var nextCond Expr
			nextCond = BinaryExpr{
				Op:    token.EQL,
				OpPos: clause.Pos(),
				X:     nameOfType,
				Y:     identAsStringLiteral(expr.(Ident)), // right is the name of the type
			}

			if bin, ok := expr.(BinaryExpr); ok {
				nextCond = bin
			}
			if nextCond == nil {
				nextCond = expr
			}
			if i == 0 {
				cond = nextCond
			} else {
				cond = BinaryExpr{
					Op:    token.LOR,
					OpPos: clause.Pos(),
					X:     cond,
					Y:     nextCond,
				}
			}
		}

		// compose goto to end of switch
		labelIdent := Ident{Name: gotoLabel}
		gotoEnd := BranchStmt{
			Tok:    token.GOTO,
			TokPos: clause.Pos(),
			Label:  &labelIdent,
		}
		list := append(clause.Body, gotoEnd)

		// compose if statement for this case
		when := IfStmt{
			IfPos: clause.Pos(),
			Cond:  cond,
			Body:  &BlockStmt{List: list},
		}
		whenFlow := when.flow(g)
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

func (s TypeSwitchStmt) stmtStep() Evaluable { return s }

func (s TypeSwitchStmt) Pos() token.Pos { return s.SwitchPos }

func (s TypeSwitchStmt) String() string {
	return fmt.Sprintf("TypeSwitchStmt(%v,%v,%v)", s.Init, s.Assign, s.Body)
}

func identAsStringLiteral(id Ident) BasicLit {
	return newBasicLit(token.NoPos, reflect.ValueOf(id.Name))
}
