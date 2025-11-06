package internal

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

// graphBuilder helps building a control flow graph by keeping track of the current step.
type graphBuilder struct {
	idgen     int
	goPkg     *packages.Package // for type information
	head      Step              // the entry point to the flow graph
	current   Step              // the current step to attach the next step to
	funcStack stack[FuncDecl]   // to keep track of current function for branch statements
}

func newGraphBuilder(goPkg *packages.Package) *graphBuilder {
	return &graphBuilder{goPkg: goPkg}
}

// next adds a new step after the current one and makes it the current step.
func (g *graphBuilder) next(e Evaluable) {
	g.nextStep(g.newStep(e))
}

// newStep creates a new step for the given Evaluable but does not add it to the current flow.
func (g *graphBuilder) newStep(e Evaluable) *evaluableStep {
	if e == nil {
		panic("call to newStep without Evaluable")
	}
	g.idgen++
	es := new(evaluableStep)
	es.id = g.idgen
	es.Evaluable = e
	return es
}

// newLabeledStep creates a labeled step but does not add it to the current flow.
func (g *graphBuilder) newLabeledStep(label string) Step {
	return &labeledStep{label: label}
}

// nextStep adds the given step after the current one and makes it the current step.
func (g *graphBuilder) nextStep(next Step) {
	if next.ID() == 0 {
		g.idgen++
		next.SetID(g.idgen)
	}

	if g.current != nil {
		if g.current.Next() != nil {
			panic(fmt.Sprintf("current %s already has a next %s, wanted %s\n", g.current, g.current.Next(), next))
		}
		if trace {
			fmt.Printf("fw.next: %d → %v\n", g.current.ID(), next)
		}
		g.current.SetNext(next)
	} else {
		if trace {
			fmt.Printf("fw.next:%v\n", next)
		}
		g.head = next
	}
	g.current = next
}
