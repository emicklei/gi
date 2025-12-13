package internal

import (
	"fmt"
	"go/token"
)

var _ Stmt = ForStmt{}

type ForStmt struct {
	ForPos token.Pos
	Init   Stmt
	Cond   Expr
	Post   Stmt
	Body   *BlockStmt
}

func (f ForStmt) Eval(vm *VM) {} // noop

func (f ForStmt) Flow(g *graphBuilder) (head Step) {
	push := new(pushEnvironmentStep) // TODO constructor with pos
	push.pos = f.Pos()
	head = push
	g.nextStep(head)
	if f.Init != nil {
		f.Init.Flow(g)
	}
	begin := new(conditionalStep)
	if f.Cond != nil {
		begin.conditionFlow = f.Cond.Flow(g)
	}
	g.nextStep(begin)
	f.Body.Flow(g)
	if f.Post != nil {
		f.Post.Flow(g)
	}
	if f.Cond != nil {
		g.nextStep(begin.conditionFlow)
	}
	pop := new(popEnvironmentStep)
	begin.elseFlow = pop
	g.current = pop
	return
}

func (f ForStmt) Pos() token.Pos {
	return f.ForPos
}

func (f ForStmt) stmtStep() Evaluable { return f }

func (f ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%v)", f.Cond)
}
