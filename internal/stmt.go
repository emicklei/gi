package internal

import (
	"fmt"
	"go/ast"
	"go/token"
)

var _ Stmt = ExprStmt{}

type ExprStmt struct {
	*ast.ExprStmt
	X Expr
}

func (s ExprStmt) stmtStep() Evaluable { return s }

func (s ExprStmt) Eval(vm *VM) {
	if trace {
		vm.traceEval(s.X)
	} else {
		s.X.Eval(vm)
	}
}

func (s ExprStmt) String() string {
	return fmt.Sprintf("ExprStmt(%v)", s.X)
}

func (s ExprStmt) Flow(g *graphBuilder) (head Step) {
	return s.X.Flow(g)
}

var _ Stmt = DeclStmt{}

type DeclStmt struct {
	*ast.DeclStmt
	Decl Decl
}

func (s DeclStmt) Eval(vm *VM) {
	s.Decl.declStep().Declare(vm)
}

func (s DeclStmt) Flow(g *graphBuilder) (head Step) {
	head = s.Decl.Flow(g)
	g.next(s)
	if head == nil {
		head = g.current
	}
	return
}
func (s DeclStmt) stmtStep() Evaluable { return s }

func (s DeclStmt) String() string {
	return fmt.Sprintf("DeclStmt(%v)", s.Decl)
}

// LabeledStmt represents a labeled statement.
// https://go.dev/ref/spec#Labeled_statements
// https://go.dev/ref/spec#Label_scopes
type LabeledStmt struct {
	*ast.LabeledStmt
	Label *Ident
	Stmt  Stmt
}

func (s LabeledStmt) Eval(vm *VM) {
	if trace {
		vm.traceEval(s.Stmt.stmtStep())
	} else {
		s.Stmt.stmtStep().Eval(vm)
	}
}

func (s LabeledStmt) Flow(g *graphBuilder) (head Step) {
	head = s.Stmt.Flow(g)
	// get statement reference and update its step
	fd := g.funcStack.top()
	ref := fd.labelToStmt[s.Label.Name]
	ref.step.SetNext(head)
	return
}

func (s LabeledStmt) String() string {
	return fmt.Sprintf("LabeledStmt(%v,%v)", s.Label, s.Stmt)
}

func (s LabeledStmt) stmtStep() Evaluable { return s }

var _ Stmt = BranchStmt{}

// BranchStmt represents a break, continue, goto, or fallthrough statement.
type BranchStmt struct {
	*ast.BranchStmt
	Label *Ident
}

func (s BranchStmt) Eval(vm *VM) {
	switch s.Tok {
	case token.GOTO:
		// af := vm.activeFuncStack.top()
		// ref := af.FuncDecl.labelToStmt[s.Label.Name]
		// af.setNextStmtIndex(ref.index)
	default:
		// TODO handle break, continue, fallthrough
	}
}

func (s BranchStmt) Flow(g *graphBuilder) (head Step) {
	switch s.Tok {
	case token.GOTO:
		head = g.newLabeledStep(fmt.Sprintf("goto %s", s.Label.Name))
		g.nextStep(head)
		fd := g.funcStack.top()
		ref, ok := fd.labelToStmt[s.Label.Name]
		if !ok {
			panic(fmt.Sprintf("undefined label: %s", s.Label.Name))
		}
		head.SetNext(ref.step)
		// branch ends the current flow
		g.current = nil
		return
	default:
		// TODO handle break, continue, fallthrough
	}
	return g.current
}

func (s BranchStmt) String() string {
	return fmt.Sprintf("BranchStmt(%v)", s.Label)
}

func (s BranchStmt) stmtStep() Evaluable { return s }

var _ Stmt = DeferStmt{}

type DeferStmt struct {
	*ast.DeferStmt
	Call Expr
}

func (d DeferStmt) String() string {
	return fmt.Sprintf("DeferStmt(%v)", d.Call)
}

func (d DeferStmt) stmtStep() Evaluable { return d }

func (d DeferStmt) Eval(vm *VM) {
	vm.activeFuncStack.top().addDefer(d.Call)
}

func (d DeferStmt) Flow(g *graphBuilder) (head Step) {
	return g.current // TODO
}
