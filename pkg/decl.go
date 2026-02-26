package pkg

import (
	"fmt"
	"go/token"
)

var _ CanDeclare = ConstVarDecl{}
var _ Decl = ConstVarDecl{}

type ConstVarDecl struct {
	specs    []ValueSpec
	iotaExpr *iotaExpr // each const block has its independent iota counter
	graph    Step
}

func (c ConstVarDecl) stmtStep() Evaluable { return c } // needed? TODO

func (c ConstVarDecl) declare(vm *VM) bool {
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
func (c ConstVarDecl) eval(vm *VM) {} // noop

// when done with this flow, the stack will have a single boolean value indicating whether all declarations were successful
func (c ConstVarDecl) flow(g *graphBuilder) (head Step) {
	// empty specs? TODO

	// process in order of declaration because of iota
	declared := newFuncStep(c.pos(), "set declared", func(vm *VM) {
		vm.pushOperand(reflectTrue)
	})
	head = declared
	g.nextStep(declared)

	resetIota := newFuncStep(c.pos(), "reset iota", func(vm *VM) {
		if c.iotaExpr != nil {
			c.iotaExpr.reset()
		}
	})
	g.nextStep(resetIota)

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

		consumeItoa := newFuncStep(spec.pos(), "consume iota", func(vm *VM) {
			if c.iotaExpr != nil {
				c.iotaExpr.next()
			}
		})
		g.nextStep(consumeItoa)
	}
	return
}
func (c ConstVarDecl) callGraph() Step {
	return c.graph
}
func (c ConstVarDecl) declStep() CanDeclare { return c }
func (c ConstVarDecl) pos() token.Pos {
	if len(c.specs) == 0 {
		return token.NoPos
	}
	return c.specs[0].pos()
}
func (c ConstVarDecl) String() string {
	return fmt.Sprintf("ConstDecl(len=%d)", len(c.specs))
}
