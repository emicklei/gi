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
	head = newPushEnvironmentStep(f.Pos())
	g.nextStep(head)
	if f.init != nil {
		f.init.flow(g)
	}
	begin := new(conditionalStep)
	if f.cond != nil {
		begin.conditionFlow = f.cond.flow(g)
	}
	g.nextStep(begin)

	// need to know where to continue for 'continue' statements in the body
	// continue with the condition of the loop unless there is no condition
	cont := g.newLabeledStep("continue", token.NoPos) // TODO give unique name
	g.continueStack.push(cont)
	defer g.continueStack.pop()

	// need to know the end of the loop for 'break' statements in the body
	// and for the else branch of the condition
	end := newPopEnvironmentStep(f.body.Pos())
	g.breakStack.push(end)
	defer g.breakStack.pop()
	begin.elseFlow = end

	f.body.flow(g)
	if f.post != nil {
		postHead := f.post.flow(g)
		cont.SetNext(postHead)
	}

	// goto the condition of the loop unless there is no condition
	if f.cond == nil {
		// if there is no post statement, we can jump directly to the beginning of the loop
		if f.post == nil {
			cont.SetNext(begin)
		}
		g.nextStep(begin)
	} else {
		// if there is no post statement, we can jump directly to the condition of the loop
		if f.post == nil {
			cont.SetNext(begin)
		}
		g.nextStep(begin.conditionFlow)
	}
	g.current = end
	return
}

func (f ForStmt) Pos() token.Pos {
	return f.forPos
}

func (f ForStmt) stmtStep() Evaluable { return f }

func (f ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%v)", f.cond)
}
