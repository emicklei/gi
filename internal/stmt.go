package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Stmt = ExprStmt{}

type ExprStmt struct {
	*ast.ExprStmt
	X Expr
}

func (s ExprStmt) stmtStep() Evaluable { return s }

func (s ExprStmt) Eval(vm *VM) {
	if trace {
		vm.eval(s.X)
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

func (s LabeledStmt) String() string {
	return fmt.Sprintf("LabeledStmt(%v)", s.Label)
}

func (s LabeledStmt) stmtStep() Evaluable { return s }

func (s LabeledStmt) Eval(vm *VM) {
	vm.eval(s.Stmt.stmtStep())
}

var _ Stmt = BranchStmt{}

// BranchStmt represents a break, continue, goto, or fallthrough statement.
type BranchStmt struct {
	*ast.BranchStmt
	Label *Ident
}

func (s BranchStmt) Eval(vm *VM) {}

func (s BranchStmt) String() string {
	return fmt.Sprintf("BranchStmt(%v)", s.Label)
}

func (s BranchStmt) stmtStep() Evaluable { return s }

func (s BranchStmt) Flow(g *graphBuilder) (head Step) {
	return head // TODO
}

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
	vm.pushNewFrame()
	defer vm.popFrame() // to handle break statements
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

func (s SwitchStmt) Flow(g *graphBuilder) (head Step) {
	head = g.newPushStackFrame()
	g.nextStep(head)
	if s.Init != nil {
		s.Init.Flow(g)
	}
	if s.Tag != nil {
		s.Tag.Flow(g)
	}
	s.Body.Flow(g)
	g.nextStep(g.newPopStackFrame())
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
			vm.eval(stmt.stmtStep())
		}
		return
	}
	f := vm.callStack.top()
	var left reflect.Value
	if len(f.operandStack) != 0 {
		left = vm.callStack.top().pop()
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
			vm.pushNewFrame()
			defer vm.popFrame()
			for _, stmt := range c.Body {
				vm.eval(stmt.stmtStep())
			}
			return
		}
	}
}

func (c CaseClause) Flow(g *graphBuilder) (head Step) {
	return g.current // TODO
}

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
	if d.Call == nil {
		return
	}
	// TODO: keep defer stack in the current frame?
	defer vm.eval(d.Call)
}

func (d DeferStmt) Flow(g *graphBuilder) (head Step) {
	return g.current // TODO
}
