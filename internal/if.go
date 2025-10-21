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
	if i.Init != nil {
		vm.eval(i.Init)
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
	begin := g.beginIf(i.Cond)
	if head == nil {
		head = begin.conditionFlow
	}

	// both true and false branch need a new and different stack frame
	truePush := g.newPushStackFrame()
	// both branches will pop and can use the same step
	pop := g.newPopStackFrame()

	g.nextStep(truePush)
	i.Body.Flow(g)
	// after true branch
	g.nextStep(pop)

	// now handle false branch
	if i.Else != nil {
		elsePush := g.newPushStackFrame()
		begin.elseFlow = elsePush
		g.current = elsePush
		i.Else.Flow(g)
		// after false branch
		g.nextStep(pop)
	}
	return head
}
