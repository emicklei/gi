package pkg

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// graphBuilder helps building a control flow graph by keeping track of the current step.
type graphBuilder struct {
	idgen            int
	goPkg            *packages.Package   // for type information
	previous         Step                // the previous step before current; or nil
	current          Step                // the current step to attach the next step to; or nil
	funcStack        stack[Func]         // to keep track of current function for branch statements
	breakStack       stack[*labeledStep] // to keep track of break targets
	continueStack    stack[*labeledStep] // to keep track of continue targets
	fallthroughStack stack[*labeledStep] // to keep track of fallthrough targets
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
		g.fatal("call to newStep without Evaluable")
	}
	g.idgen++
	es := new(evaluableStep)
	es.id = g.idgen
	es.Evaluable = e
	return es
}

// newLabeledStep creates a labeled step but does not add it to the current flow.
func (g *graphBuilder) newLabeledStep(label string, pos token.Pos) *labeledStep {
	return &labeledStep{label: label, pos: pos}
}

// nextStep adds the given step after the current one and makes it the current step.
func (g *graphBuilder) nextStep(next Step) {
	// ensure it has an ID
	if next.ID() == 0 {
		g.idgen++
		next.SetID(g.idgen)
	}
	if g.current != nil {
		if g.current.Next() != nil {
			g.fatalf("current %s already has a next %s, wanted %s\n", g.current, g.current.Next(), next)
		}
		if trace {
			fmt.Printf("fw: %d → %v", g.current.ID(), next)
			if g.goPkg != nil && g.goPkg.Fset != nil {
				f := g.goPkg.Fset.File(g.current.Pos())
				if f != nil {
					nodir := filepath.Base(f.Name())
					fmt.Print(" @ ", nodir, ":", f.Line(g.current.Pos()))
				} else {
					fmt.Print(" @ bad file info")
				}
			}
			fmt.Println()
		}
		g.current.SetNext(next)
	} else {
		if trace {
			fmt.Printf("fw: %v\n", next)
		}
	}
	g.previous = g.current
	g.current = next
}

// deprecated: use fatalf instead
func (g *graphBuilder) fatal(err any) {
	fmt.Fprintln(os.Stderr, "[gi] fatal graph error:", err)
	panic(err)
}
func (g *graphBuilder) fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[gi] fatal graph error: "+format+"\n", args...)
	panic(fmt.Sprintf(format, args...))
}

func (g *graphBuilder) stepBack() {
	g.idgen--
	g.current = g.previous
	if g.current != nil {
		g.current.SetNext(nil)
	}
	g.previous = nil // we don't track further back for now
}
