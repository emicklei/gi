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

func (f ForStmt) eval(vm *VM) {} // noop

func (f ForStmt) flow(g *graphBuilder) (head Step) {
	return f.flowWithOptions(g, false)
}

func (f ForStmt) flowWithOptions(g *graphBuilder, skipNewEnvironment bool) (head Step) {
	if !skipNewEnvironment {
		head = newPushEnvironmentStep(f.pos())
		g.nextStep(head)
	}
	if f.init != nil {
		initFlow := f.init.flow(g)
		if head == nil {
			head = initFlow
		}
	}
	begin := new(conditionalStep)
	if f.cond != nil {
		begin.conditionFlow = f.cond.flow(g)
	}
	if head == nil {
		head = begin
	}
	g.nextStep(begin)

	// need to know where to continue for 'continue' statements in the body
	// continue with the condition of the loop unless there is no condition
	cont := g.newLabeledStep("~continue", token.NoPos)
	g.continueStack.push(cont)
	defer g.continueStack.pop()

	// need to know the end of the loop for 'break' statements in the body
	// and for the else branch of the condition
	braek := g.newLabeledStep("~break", token.NoPos)
	g.breakStack.push(braek)
	defer g.breakStack.pop()

	if !skipNewEnvironment {
		// body runs in separate env in which loop vars are copied so they have their own unique address
		g.nextStep(newPushEnvironmentStep(f.body.lbracePos))
		g.nextStep(newFuncStep(f.body.lbracePos, "~parent->child", func(vm *VM) {
			vm.currentFrame.env.parent().copyValues(vm.currentFrame.env)
		}))
	}
	f.body.flow(g)
	if !skipNewEnvironment {
		// put back copied loop vars so any modifications are visible to the loop condition and loop post.
		g.nextStep(newFuncStep(f.body.lbracePos, "~child->parent", func(vm *VM) {
			vm.currentFrame.env.copyValues(vm.currentFrame.env.parent())
		}))
		g.nextStep(g.newPopEnvironmentStep(f.body.lbracePos))
	}

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
	var end Step
	if !skipNewEnvironment {
		// leave the scope of the loop
		end = g.newPopEnvironmentStep(f.body.pos())
	} else {
		end = g.newLabeledStep("~end", token.NoPos) // noop
	}
	// break goes to the end of the loop
	braek.SetNext(end)
	// else branch of the condition goes to the end of the loop
	begin.elseFlow = end
	g.current = end
	return
}

func (f ForStmt) pos() token.Pos {
	return f.forPos
}

func (f ForStmt) stmtStep() Evaluable { return f }

func (f ForStmt) String() string {
	return fmt.Sprintf("ForStmt(%v)", f.cond)
}
