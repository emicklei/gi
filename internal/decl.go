package internal

import (
	"fmt"
	"go/token"
)

var _ CanDeclare = ConstDecl{}
var _ Decl = ConstDecl{}

type ConstDecl struct {
	Specs     []ValueSpec
	iotaExpr  *iotaExpr // each const block has its independent iota counter
	callGraph Step
}

func (c ConstDecl) CallGraph() Step {
	return c.callGraph
}
func (c ConstDecl) declStep() CanDeclare { return c }
func (c ConstDecl) Declare(vm *VM) bool {
	done := true
	if c.iotaExpr != nil {
		c.iotaExpr.reset()
	}
	for _, spec := range c.Specs {
		vm.takeAllStartingAt(spec.CallGraph())
		if !spec.Declare(vm) {
			done = false
		}
		if c.iotaExpr != nil {
			c.iotaExpr.next()
		}
	}
	return done
}
func (c ConstDecl) Eval(vm *VM) {}
func (c ConstDecl) Flow(g *graphBuilder) (head Step) {
	// process in order of declaration because of iota
	for i, spec := range c.Specs {
		s := spec.Flow(g)
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
func (c ConstDecl) Pos() token.Pos {
	if len(c.Specs) == 0 {
		return token.NoPos
	}
	return c.Specs[0].Pos()
}
func (c ConstDecl) String() string {
	return fmt.Sprintf("ConstDecl(len=%d)", len(c.Specs))
}
