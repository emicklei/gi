package internal

import (
	"fmt"
	"os"

	"github.com/emicklei/dot"
	"golang.org/x/tools/go/packages"
)

// graphBuilder helps building a control flow graph by keeping track of the current step.
type graphBuilder struct {
	goPkg     *packages.Package // for type information
	head      Step              // the entry point to the flow graph
	current   Step              // the current step to attach the next step to
	dotFile   string            // for overriding the default graph.dot file (only used when calling dotify)
	funcStack stack[FuncDecl]   // to keep track of current function for branch statements
}

func newGraphBuilder(goPkg *packages.Package) *graphBuilder {
	return &graphBuilder{goPkg: goPkg}
}

// next adds a new step after the current one and makes it the current step.
func (g *graphBuilder) next(e Evaluable) {
	g.nextStep(newStep(e))
}

// nextStep adds the given step after the current one and makes it the current step.
func (g *graphBuilder) nextStep(next Step) {
	if g.current != nil {
		if g.current.Next() != nil {
			panic(fmt.Sprintf("current %s already has a next %s, wanted %s\n", g.current, g.current.Next(), next))
		}
		g.current.SetNext(next)
	} else {
		g.head = next
	}
	g.current = next
}

// beginIf creates a conditional step with the given condition
// and makes it the current step. It returns the created conditional step to set the else branch later.
// TODO maybe inline this
func (g *graphBuilder) beginIf(cond Expr) *conditionalStep {
	head := cond.Flow(g)
	c := &conditionalStep{
		conditionFlow: head,
		step:          newStep(nil),
	}
	g.nextStep(c)
	return c
}

// TODO maybe inline this
func (g *graphBuilder) endIf(begin *conditionalStep) {
	pop := g.newPopStackFrame()
	begin.elseFlow = pop
	g.current = pop
}

// newPushStackFrameStep creates a step that pushes a new stack frame.
func (g *graphBuilder) newPushStackFrame() *pushStackFrameStep {
	return &pushStackFrameStep{step: newStep(nil)}
}

// newPopStackFrameStep creates a step that pops the current stack frame.
func (g *graphBuilder) newPopStackFrame() *popStackFrameStep {
	return &popStackFrameStep{step: newStep(nil)}
}

func (g *graphBuilder) dotFilename() string {
	if g.dotFile != "" {
		return g.dotFile
	}
	return "graph.dot"
}

// dotify writes the current graph to a file.
func (g *graphBuilder) dotify() {
	if g.current == nil {
		return
	}
	d := dot.NewGraph(dot.Directed)
	visited := map[int]dot.Node{}
	g.head.Traverse(d, visited)
	os.WriteFile(g.dotFilename(), []byte(d.String()), 0644)
}
