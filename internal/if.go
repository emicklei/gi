package internal

import (
	"fmt"
	"go/ast"
)

var _ Stmt = IfStmt{}

type IfStmt struct {
	*ast.IfStmt
	Init Expr
	Cond Expr
	Body *BlockStmt
	Else Stmt // else if ...
}

func (i IfStmt) stmtStep() Evaluable { return i }

func (i IfStmt) String() string {
	return fmt.Sprintf("IfStmt(%v, %v, %v, %v)", i.Init, i.Cond, i.Body, i.Else)
}

func (i IfStmt) Eval(vm *VM) {
	if trace {
		if i.Init != nil {
			vm.traceEval(i.Init)
		}
		rv := vm.returnsEval(i.Cond)
		if rv.Bool() {
			vm.traceEval(i.Body)
			return
		}
		if i.Else != nil {
			vm.traceEval(i.Else.stmtStep())
			return
		}
	} else {
		if i.Init != nil {
			i.Init.Eval(vm)
		}
		rv := vm.returnsEval(i.Cond)
		if rv.Bool() {
			i.Body.Eval(vm)
			return
		}
		if i.Else != nil {
			i.Else.stmtStep().Eval(vm)
			return
		}
	}
}

func (i IfStmt) Flow(g *graphBuilder) (head Step) {
	if i.Init != nil {
		head = i.Init.Flow(g)
	}
	// condition can have its own assigments
	g.nextStep(new(pushEnvironmentStep))
	if head == nil {
		head = g.current
	}
	begin := new(conditionalStep)
	begin.conditionFlow = i.Cond.Flow(g)
	g.nextStep(begin)

	// true branch
	i.Body.Flow(g)
	pop := new(popEnvironmentStep)
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
