package pkg

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

func (f ForStmt) flow(g *graphBuilder) (head Step) {
	push := new(pushEnvironmentStep) // TODO constructor with pos
	push.pos = f.Pos()
	head = push
	g.nextStep(head)
	if f.Init != nil {
		f.Init.flow(g)
	}
	begin := new(conditionalStep)
	if f.Cond != nil {
		begin.conditionFlow = f.Cond.flow(g)
	}
	g.nextStep(begin)
	f.Body.flow(g)
	if f.Post != nil {
		f.Post.flow(g)
	}
	if f.Cond != nil {
		g.nextStep(begin.conditionFlow)
	}
	pop := newPopEnvironmentStep(f.Body.Pos())
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
