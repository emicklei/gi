package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Stmt = SwitchStmt{}

// A SwitchStmt represents an expression switch statement.
type SwitchStmt struct {
	switchPos token.Pos
	init      Stmt // initialization statement; or nil
	tag       Expr // tag expression; or nil
	body      BlockStmt
}

func (s SwitchStmt) stmtStep() Evaluable { return s }

func (s SwitchStmt) Eval(vm *VM) {
	vm.currentFrame.pushEnv()
	defer vm.currentFrame.popEnv()
	if s.init != nil {
		vm.eval(s.init.stmtStep())
	}
	if s.tag != nil {
		vm.eval(s.tag)
	}
	vm.eval(s.body)
}

func (s SwitchStmt) flow(g *graphBuilder) (head Step) {
	if s.init != nil {
		head = s.init.flow(g)
	}
	if s.tag != nil {
		if head == nil {
			head = s.tag.flow(g)
		} else {
			_ = s.tag.flow(g)
		}
	}
	gotoLabel := fmt.Sprintf("switch-end-%d", g.idgen)
	gotoStep := g.newLabeledStep(gotoLabel, s.Pos())
	ref := stmtReference{step: gotoStep} // has no ID
	g.funcStack.top().putGotoReference(gotoLabel, ref)

	for _, stmt := range s.body.list {
		clause := stmt.(CaseClause)

		// check for default case
		if clause.List == nil {
			// compose goto to end of switch
			labelIdent := Ident{name: gotoLabel}
			gotoEnd := BranchStmt{
				tok:    token.GOTO,
				tokPos: clause.Pos(),
				label:  &labelIdent,
			}
			list := append(clause.Body, gotoEnd)

			// if previous case had a fallthrough then set its next
			if len(g.fallthroughStack) > 0 {
				fall := g.fallthroughStack.pop()
				destination := fallThroughDestination{from: fall}
				// put it first in the list so that it will be executed before the case body
				list = append([]Stmt{destination}, list...)
			}

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
			// if previous case had a fallthrough then we need to pop it because there is no body to execute
			if len(g.fallthroughStack) > 0 {
				g.fallthroughStack.pop()
			}
			continue
		}
		// compose condition
		var cond Expr
		// build a chain of OR expressions for each case expression
		for i, expr := range clause.List {
			var nextCond Expr
			if s.tag != nil {
				nextCond = BinaryExpr{
					op:    token.EQL,
					opPos: clause.Pos(),
					x:     s.tag,
					y:     expr,
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
					op:    token.LOR,
					opPos: clause.Pos(),
					x:     cond,
					y:     nextCond,
				}
			}
		}

		// compose goto to end of switch
		labelIdent := Ident{name: gotoLabel}
		gotoEnd := BranchStmt{
			tok:    token.GOTO,
			tokPos: clause.Pos(),
			label:  &labelIdent,
		}
		list := append(clause.Body, gotoEnd)

		// if previous case had a fallthrough then set its next
		if len(g.fallthroughStack) > 0 {
			fall := g.fallthroughStack.pop()
			destination := fallThroughDestination{from: fall}
			// put it first in the list so that it will be executed before the case body
			list = append([]Stmt{destination}, list...)
		}

		// compose if statement for this case
		when := IfStmt{
			ifPos: clause.Pos(),
			cond:  cond,
			body:  &BlockStmt{list: list},
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

var _ Stmt = fallThroughDestination{}

type fallThroughDestination struct {
	from *labeledStep
}

func (c fallThroughDestination) flow(g *graphBuilder) (head Step) {
	to := g.newLabeledStep("fallthrough destination", token.NoPos)
	c.from.SetNext(to)
	g.nextStep(to)
	return to

}
func (c fallThroughDestination) stmtStep() Evaluable {
	return nil
}

func (s SwitchStmt) String() string {
	return fmt.Sprintf("SwitchStmt(%v,%v,%v)", s.init, s.tag, s.body)
}
func (s SwitchStmt) Pos() token.Pos { return s.switchPos }

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
	ref := stmtReference{step: gotoStep} // has no ID
	g.funcStack.top().putGotoReference(gotoLabel, ref)

	nameOfType := Ident{namePos: s.Pos(), name: internalVarName("switch-type-name", g.idgen)}

	nameOfTypeAssignment := AssignStmt{
		tokPos: s.SwitchPos,
		tok:    token.DEFINE,
		lhs:    []Expr{nameOfType},
		rhs:    []Expr{noExpr{}}, // no expression because TypeAssertExpr pushes two values
	}
	nameOfTypeAssignment.flow(g)

	for _, stmt := range s.Body.list {
		clause := stmt.(CaseClause)

		// check for default case
		if clause.List == nil {
			// compose goto to end of switch
			labelIdent := Ident{name: gotoLabel}
			gotoEnd := BranchStmt{
				tok:    token.GOTO,
				tokPos: clause.Pos(),
				label:  &labelIdent,
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
				op:    token.EQL,
				opPos: clause.Pos(),
				x:     nameOfType,
				y:     identAsStringLiteral(expr.(Ident)), // right is the name of the type
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
					op:    token.LOR,
					opPos: clause.Pos(),
					x:     cond,
					y:     nextCond,
				}
			}
		}

		// compose goto to end of switch
		labelIdent := Ident{name: gotoLabel}
		gotoEnd := BranchStmt{
			tok:    token.GOTO,
			tokPos: clause.Pos(),
			label:  &labelIdent,
		}
		list := append(clause.Body, gotoEnd)

		// compose if statement for this case
		when := IfStmt{
			ifPos: clause.Pos(),
			cond:  cond,
			body:  &BlockStmt{list: list},
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
	return newBasicLit(token.NoPos, reflect.ValueOf(id.name))
}
