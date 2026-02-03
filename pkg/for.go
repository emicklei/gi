package pkg

import (
	"fmt"
	"go/token"
)

var _ Stmt = ForStmt{}

type ForStmt struct {
	forPos token.Pos
	init   Stmt
	cond   Expr
	post   Stmt
	body   *BlockStmt
}

func (f ForStmt) Eval(vm *VM) {} // noop

func (f ForStmt) flow(g *graphBuilder) (head Step) {
	push := new(pushEnvironmentStep) // TODO constructor with pos
	push.pos = f.Pos()
	head = push
	g.nextStep(head)
	if f.init != nil {
		f.init.flow(g)
	}
	begin := new(conditionalStep)
	if f.cond != nil {
		begin.conditionFlow = f.cond.flow(g)
	}
	g.nextStep(begin)
	f.body.flow(g)
	if f.post != nil {
		f.post.flow(g)
	}
	if f.cond != nil {
		g.nextStep(begin.conditionFlow)
	}
	pop := newPopEnvironmentStep(f.body.Pos())
	begin.elseFlow = pop
	g.current = pop
	return
}

func (f ForStmt) Pos() token.Pos {
	return f.forPos
}

func (f ForStmt) stmtStep() Evaluable { return f }

func (f ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%v)", f.cond)
}
