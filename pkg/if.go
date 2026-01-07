package pkg

import (
	"fmt"
	"go/token"
)

var _ Stmt = IfStmt{}

type IfStmt struct {
	IfPos token.Pos
	Init  Stmt
	Cond  Expr
	Body  *BlockStmt
	Else  Stmt // else if ...
}

func (i IfStmt) Eval(vm *VM) {
	if i.Init != nil {
		vm.eval(i.Init.stmtStep())
	}
	rv := vm.returnsEval(i.Cond)
	if rv.Bool() {
		vm.eval(i.Body)
		return
	}
	if i.Else != nil {
		vm.eval(i.Else.stmtStep())
		return
	}
}

func (i IfStmt) Flow(g *graphBuilder) (head Step) {
	if i.Init != nil {
		head = i.Init.Flow(g)
	}
	// condition can have its own assigments
	push := newPushEnvironmentStep(i.Pos())
	g.nextStep(push)
	if head == nil {
		head = g.current
	}
	begin := new(conditionalStep)
	begin.conditionFlow = i.Cond.Flow(g)
	g.nextStep(begin)

	// true branch
	i.Body.Flow(g)
	pop := newPopEnvironmentStep(i.Body.Pos())
	g.nextStep(pop)

	// false branch
	if i.Else != nil {
		g.current = nil
		elseFlow := i.Else.Flow(g)
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

func (i IfStmt) Pos() token.Pos { return i.IfPos }

func (i IfStmt) String() string {
	return fmt.Sprintf("IfStmt(%v, %v, %v, %v)", i.Init, i.Cond, i.Body, i.Else)
}
