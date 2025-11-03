package internal

import (
	"fmt"
	"go/ast"
)

var _ Stmt = ForStmt{}

type ForStmt struct {
	*ast.ForStmt
	Init Stmt
	Cond Expr
	Post Stmt
	Body *BlockStmt
}

func (f ForStmt) stmtStep() Evaluable { return f }

func (f ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%v)", f.Cond)
}
func (f ForStmt) Eval(vm *VM) {} // noop

func (f ForStmt) Flow(g *graphBuilder) (head Step) {
	head = new(pushStackFrameStep)
	g.nextStep(head)
	if f.Init != nil {
		f.Init.Flow(g)
	}
	begin := g.beginIf(f.Cond)
	f.Body.Flow(g)
	f.Post.Flow(g)
	g.nextStep(begin.conditionFlow)
	pop := new(popStackFrameStep)
	begin.elseFlow = pop
	g.current = pop
	return
}
