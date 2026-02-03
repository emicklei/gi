package pkg

import (
	"fmt"
	"go/token"
)

var _ CanDeclare = ConstDecl{}
var _ Decl = ConstDecl{}

type ConstDecl struct {
	specs    []ValueSpec
	iotaExpr *iotaExpr // each const block has its independent iota counter
	graph    Step
}

func (c ConstDecl) declare(vm *VM) bool {
	done := true
	if c.iotaExpr != nil {
		// reset iota for this const declaration
		c.iotaExpr.reset()
	}
	for _, spec := range c.specs {
		vm.takeAllStartingAt(spec.callGraph())
		if !spec.declare(vm) {
			done = false
			// continue trying others; we come back later
		}
		if c.iotaExpr != nil {
			// if iota was used, advance it
			c.iotaExpr.next()
		}
	}
	return done
}
func (c ConstDecl) Eval(vm *VM) {} // noop

func (c ConstDecl) flow(g *graphBuilder) (head Step) {
	// process in order of declaration because of iota
	for i, spec := range c.specs {
		s := spec.flow(g)
		if i == 0 {
			head = s
		}
	}
	// empty specs?
	if head == nil {
		head = g.current
	}
	return
}
func (c ConstDecl) callGraph() Step {
	return c.graph
}
func (c ConstDecl) declStep() CanDeclare { return c }
func (c ConstDecl) Pos() token.Pos {
	if len(c.specs) == 0 {
		return token.NoPos
	}
	return c.specs[0].Pos()
}
func (c ConstDecl) String() string {
	return fmt.Sprintf("ConstDecl(len=%d)", len(c.specs))
}
