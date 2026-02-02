package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Stmt = ExprStmt{}

type ExprStmt struct {
	X Expr
}

func (s ExprStmt) stmtStep() Evaluable { return s }

func (s ExprStmt) Eval(vm *VM) {
	vm.eval(s.X)
}

func (s ExprStmt) flow(g *graphBuilder) (head Step) {
	return s.X.flow(g)
}

func (s ExprStmt) String() string {
	return fmt.Sprintf("ExprStmt(%v)", s.X)
}

func (s ExprStmt) Pos() token.Pos {
	return s.X.Pos()
}

var _ Stmt = DeclStmt{}

type DeclStmt struct {
	decl Decl
}

func (s DeclStmt) Eval(vm *VM) {
	s.decl.declStep().declare(vm)
}

func (s DeclStmt) flow(g *graphBuilder) (head Step) {
	head = s.decl.flow(g)
	g.next(s)
	if head == nil {
		head = g.current
	}
	return
}

func (s DeclStmt) Pos() token.Pos {
	return s.decl.Pos()
}

func (s DeclStmt) stmtStep() Evaluable { return s }

func (s DeclStmt) String() string {
	return fmt.Sprintf("DeclStmt(%v)", s.decl)
}

// LabeledStmt represents a labeled statement.
// https://go.dev/ref/spec#Labeled_statements
// https://go.dev/ref/spec#Label_scopes
type LabeledStmt struct {
	colonPos  token.Pos
	label     *Ident
	statement Stmt
}

func (s LabeledStmt) Eval(vm *VM) {
	vm.eval(s.statement.stmtStep())
}

func (s LabeledStmt) flow(g *graphBuilder) (head Step) {
	head = s.statement.flow(g)
	// get statement reference and update its step
	fd := g.funcStack.top()
	ref := fd.gotoReference(s.label.Name)
	ref.step.SetNext(head)
	return
}

func (s LabeledStmt) Pos() token.Pos { return s.colonPos }

func (s LabeledStmt) String() string {
	return fmt.Sprintf("LabeledStmt(%v,%v)", s.label, s.statement)
}

func (s LabeledStmt) stmtStep() Evaluable { return s }

var _ Stmt = BranchStmt{}

// BranchStmt represents a break, continue, goto, or fallthrough statement.
type BranchStmt struct {
	TokPos token.Pos   // position of Tok
	Tok    token.Token // keyword token (BREAK, CONTINUE, GOTO, FALLTHROUGH)
	Label  *Ident
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

func (s BranchStmt) flow(g *graphBuilder) (head Step) {
	switch s.Tok {
	case token.GOTO:
		head = g.newLabeledStep(fmt.Sprintf("goto %s", s.Label.Name), s.Pos())
		g.nextStep(head)
		fd := g.funcStack.top()
		ref := fd.gotoReference(s.Label.Name)
		head.SetNext(ref.step)
		// branch ends the current flow
		g.current = nil
		return
	default:
		// TODO handle break, continue, fallthrough
	}
	return g.current
}

func (s BranchStmt) Pos() token.Pos { return s.TokPos }

func (s BranchStmt) String() string {
	return fmt.Sprintf("BranchStmt(%v)", s.Label)
}

func (s BranchStmt) stmtStep() Evaluable { return s }

var _ Stmt = DeferStmt{}

type DeferStmt struct {
	DeferPos token.Pos
	Call     Expr
	// detached flow
	callGraph Step
}

func (d DeferStmt) Eval(vm *VM) {
	frame := vm.currentFrame
	// create a new env and copy the current argument values
	env := frame.env.newChild() // TODO needed?
	call := d.Call.(CallExpr)
	vals := make([]reflect.Value, len(call.Args))
	for i, arg := range call.Args { // TODO variadic
		vals[i] = vm.returnsEval(arg)
	}
	frame.env.markSharedReferenced()
	invocation := funcInvocation{
		flow:      d.callGraph,
		env:       env,
		arguments: vals,
	}
	frame.defers = append(frame.defers, invocation)
}

func (d DeferStmt) flow(g *graphBuilder) (head Step) {
	g.next(d)
	return g.current
}

func (d DeferStmt) Pos() token.Pos { return d.DeferPos }

func (d DeferStmt) String() string {
	return fmt.Sprintf("DeferStmt(%v)", d.Call)
}

func (d DeferStmt) stmtStep() Evaluable { return d }

var _ Stmt = BlockStmt{}

type BlockStmt struct {
	LbracePos token.Pos // position of "{"
	List      []Stmt
}

func (b BlockStmt) Eval(vm *VM) {
	for _, stmt := range b.List {
		vm.eval(stmt.stmtStep())
	}
}

func (b BlockStmt) flow(g *graphBuilder) (head Step) {
	head = g.current
	for i, stmt := range b.List {
		if i == 0 {
			head = stmt.flow(g)
			continue
		}
		_ = stmt.flow(g)
	}
	return
}

func (b BlockStmt) stmtStep() Evaluable { return b }

func (b BlockStmt) Pos() token.Pos { return b.LbracePos }

func (b BlockStmt) String() string {
	return fmt.Sprintf("BlockStmt(len=%d)", len(b.List))
}
