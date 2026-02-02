package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Stmt = ReturnStmt{}

type ReturnStmt struct {
	returnPos token.Pos
	results   []Expr
}

func (r ReturnStmt) Eval(vm *VM) {
	if len(r.results) == 0 {
		return
	}
	results := make([]reflect.Value, len(r.results))
	for i := range r.results {
		val := vm.popOperand()
		results[i] = val
	}
	// bind result values to named results of the function if any
	fn := vm.currentFrame.creator
	if fn != nil {
		i := 0
		for _, fields := range fn.results().List {
			for _, name := range fields.Names {
				owner := vm.currentFrame.env.valueOwnerOf(name.Name)
				owner.set(name.Name, results[i])
				i++
			}
		}
	}
}

func (r ReturnStmt) flow(g *graphBuilder) (head Step) {
	// reverse order to keep Eval correct
	for i := len(r.results) - 1; i >= 0; i-- {
		each := r.results[i]
		if i == len(r.results)-1 {
			head = each.flow(g)
			continue
		}
		each.flow(g)
	}
	ret := new(returnStep)
	ret.Evaluable = r
	g.nextStep(ret)
	// if nothing to return then returnStep is the head
	if head == nil {
		head = g.current
	}
	// no next step after return
	g.current = nil
	return
}

func (r ReturnStmt) Pos() token.Pos {
	return r.returnPos
}

func (r ReturnStmt) stmtStep() Evaluable { return r }

func (r ReturnStmt) String() string {
	return fmt.Sprintf("return(len=%d)", len(r.results))
}
