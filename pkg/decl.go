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

func (c ConstDecl) stmtStep() Evaluable { return c } // needed? TODO

func (c ConstDecl) declare(vm *VM) bool {
	done := true
	if c.iotaExpr != nil {
		// reset iota for this const declaration
		c.iotaExpr.reset()
	}
	for _, spec := range c.specs {
		vm.stepThrough(spec.callGraph())
		if !spec.declare(vm) {
			done = false
			// continue trying others; we will come back later
		}
		if c.iotaExpr != nil {
			// if iota was used, advance it
			c.iotaExpr.next()
		}
	}
	return done
}
func (c ConstDecl) eval(vm *VM) {} // noop

// when done with this flow, the stack will have a single boolean value indicating whether all declarations were successful
func (c ConstDecl) flow(g *graphBuilder) (head Step) {
	// empty specs? TODO

	// process in order of declaration because of iota
	declared := newFuncStep(c.pos(), "set declared", func(vm *VM) {
		vm.pushOperand(reflectTrue)
	})
	head = declared
	g.nextStep(declared)

	for _, spec := range c.specs {
		spec.flow(g)
		// each spec pushes the result of its declaration on the stack; we pop it and push true if declared, false otherwise
		update := newFuncStep(spec.pos(), "update declared", func(vm *VM) {
			result := vm.popOperand()
			// take overall result; if any declaration failed, the overall result is false
			previousResult := vm.popOperand()
			if result == reflectFalse {
				// new declared value
				vm.pushOperand(reflectFalse)
			} else {
				// keep previous result of declared
				vm.pushOperand(previousResult)
			}
		})
		g.nextStep(update)
	}
	return
}
func (c ConstDecl) callGraph() Step {
	return c.graph
}
func (c ConstDecl) declStep() CanDeclare { return c }
func (c ConstDecl) pos() token.Pos {
	if len(c.specs) == 0 {
		return token.NoPos
	}
	return c.specs[0].pos()
}
func (c ConstDecl) String() string {
	return fmt.Sprintf("ConstDecl(len=%d)", len(c.specs))
}
