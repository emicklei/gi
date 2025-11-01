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
	begin := g.beginIf(i.Cond)
	if head == nil {
		head = begin.conditionFlow
	}

	// both true and false branch need a new and different stack frame
	truePush := new(pushStackFrameStep)
	// both branches will pop and can use the same step
	pop := new(popStackFrameStep)

	g.nextStep(truePush)
	i.Body.Flow(g)
	// after true branch
	g.nextStep(pop)

	// now handle false branch
	if i.Else != nil {
		elsePush := new(pushStackFrameStep)
		begin.elseFlow = elsePush
		g.current = elsePush
		i.Else.Flow(g)
		// after false branch
		g.nextStep(pop)
	}
	return head
}
