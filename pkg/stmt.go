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
	ref := fd.gotoReference(s.label.name)
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
	tokPos token.Pos   // position of Tok
	tok    token.Token // keyword token (BREAK, CONTINUE, GOTO, FALLTHROUGH)
	label  *Ident
}

func (s BranchStmt) Eval(vm *VM) {} // no-op; flow is handled in graph building

func (s BranchStmt) flow(g *graphBuilder) (head Step) {
	switch s.tok {
	case token.GOTO:
		head = g.newLabeledStep(fmt.Sprintf("goto %s", s.label.name), s.Pos())
		g.nextStep(head)
		fd := g.funcStack.top()
		ref := fd.gotoReference(s.label.name)
		head.SetNext(ref.step)
		// branch ends the current flow
		g.current = nil
		return
	case token.BREAK:
		target := g.breakStack.top().(*labeledStep) // safe?
		target.SetPos(s.Pos())
		g.nextStep(target)
		g.current = nil
		return
	case token.CONTINUE:
		target := g.continueStack.top().(*labeledStep) // safe?
		target.SetPos(s.Pos())
		g.nextStep(target)
		g.current = nil
		return
	default:
		g.fatal("TODO handle break, continue, fallthrough")
	}
	return g.current
}

func (s BranchStmt) Pos() token.Pos { return s.tokPos }

func (s BranchStmt) String() string {
	return fmt.Sprintf("BranchStmt(%v)", s.label)
}

func (s BranchStmt) stmtStep() Evaluable { return s }

var _ Stmt = DeferStmt{}

type DeferStmt struct {
	deferPos token.Pos
	call     Expr
	// detached flow
	callGraph Step
}

func (d DeferStmt) Eval(vm *VM) {
	frame := vm.currentFrame
	// create a new env and copy the current argument values
	env := frame.env.newChild() // TODO needed?
	call := d.call.(CallExpr)
	vals := make([]reflect.Value, len(call.args))
	for i, arg := range call.args { // TODO variadic
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

func (d DeferStmt) Pos() token.Pos { return d.deferPos }

func (d DeferStmt) String() string {
	return fmt.Sprintf("DeferStmt(%v)", d.call)
}

func (d DeferStmt) stmtStep() Evaluable { return d }

var _ Stmt = BlockStmt{}

type BlockStmt struct {
	lbracePos token.Pos // position of "{"
	list      []Stmt
}

func (b BlockStmt) Eval(vm *VM) {
	for _, stmt := range b.list {
		vm.eval(stmt.stmtStep())
	}
}

func (b BlockStmt) flow(g *graphBuilder) (head Step) {
	head = g.current
	for i, stmt := range b.list {
		if i == 0 {
			head = stmt.flow(g)
			continue
		}
		_ = stmt.flow(g)
	}
	return
}

func (b BlockStmt) stmtStep() Evaluable { return b }

func (b BlockStmt) Pos() token.Pos { return b.lbracePos }

func (b BlockStmt) String() string {
	return fmt.Sprintf("BlockStmt(len=%d)", len(b.list))
}
