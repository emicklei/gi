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
func (f ForStmt) Eval(vm *VM) {
	vm.pushNewFrame(f)
	if trace {
		if f.Init != nil {
			vm.traceEval(f.Init.stmtStep())
		}
		for vm.returnsEval(f.Cond).Bool() {
			vm.traceEval(f.Body.stmtStep())
			vm.traceEval(f.Post.stmtStep())
		}
	} else {
		if f.Init != nil {
			f.Init.stmtStep().Eval(vm)
		}
		for vm.returnsEval(f.Cond).Bool() {
			f.Body.stmtStep().Eval(vm)
			f.Post.stmtStep().Eval(vm)
		}
	}
	vm.popFrame()
}

func (f ForStmt) Flow(g *graphBuilder) (head Step) {
	head = g.newPushStackFrame()
	g.nextStep(head)
	if f.Init != nil {
		f.Init.Flow(g)
	}
	begin := g.beginIf(f.Cond)
	f.Body.Flow(g)
	f.Post.Flow(g)
	g.nextStep(begin.conditionFlow)
	g.endIf(begin)
	return
}
