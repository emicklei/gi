package pkg

import (
	"fmt"
	"go/token"
)

var _ Stmt = IfStmt{}

type IfStmt struct {
	ifPos  token.Pos
	init   Stmt
	cond   Expr
	body   *BlockStmt
	elseif Stmt
}

func (i IfStmt) Eval(vm *VM) {
	// no op
}

func (i IfStmt) flow(g *graphBuilder) (head Step) {
	if i.init != nil {
		head = i.init.flow(g)
	}
	// condition can have its own assigments
	push := newPushEnvironmentStep(i.Pos())
	g.nextStep(push)
	if head == nil {
		head = g.current
	}
	begin := new(conditionalStep)
	begin.conditionFlow = i.cond.flow(g)
	g.nextStep(begin)

	// true branch
	i.body.flow(g)
	pop := newPopEnvironmentStep(i.body.Pos())
	g.nextStep(pop)

	// false branch
	if i.elseif != nil {
		g.current = nil
		elseFlow := i.elseif.flow(g)
		// TODO if the body ends with a branch then pop should should be done before.
		// fmt.Println(g.current)
		begin.elseFlow = elseFlow
		g.nextStep(pop)
	} else {
		begin.elseFlow = pop
	}
	return head
}

func (i IfStmt) stmtStep() Evaluable { return i }

func (i IfStmt) Pos() token.Pos { return i.ifPos }

func (i IfStmt) String() string {
	return fmt.Sprintf("IfStmt(%v, %v, %v, %v)", i.init, i.cond, i.body, i.elseif)
}
