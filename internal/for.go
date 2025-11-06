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

func (f ForStmt) Eval(vm *VM) {} // noop

func (f ForStmt) Flow(g *graphBuilder) (head Step) {
	head = new(pushEnvironmentStep)
	g.nextStep(head)
	if f.Init != nil {
		f.Init.Flow(g)
	}
	begin := new(conditionalStep)
	begin.conditionFlow = f.Cond.Flow(g)
	g.nextStep(begin)
	f.Body.Flow(g)
	f.Post.Flow(g)
	g.nextStep(begin.conditionFlow)
	pop := new(popEnvironmentStep)
	begin.elseFlow = pop
	g.current = pop
	return
}

func (f ForStmt) stmtStep() Evaluable { return f }

func (f ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%v)", f.Cond)
}
