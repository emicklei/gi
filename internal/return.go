package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Stmt = ReturnStmt{}

type ReturnStmt struct {
	ReturnPos token.Pos
	Results   []Expr
}

func (r ReturnStmt) Eval(vm *VM) {
	if len(r.Results) == 0 {
		return
	}
	results := make([]reflect.Value, len(r.Results))
	for i := range r.Results {
		val := vm.popOperand()
		results[i] = val
	}
	// bind result values to named results of the function if any
	fd, ok := vm.currentFrame.creator.(Func)
	if ok {
		i := 0
		for _, fields := range fd.Results().List {
			for _, name := range fields.Names {
				vm.localEnv().set(name.Name, results[i])
				i++
			}
		}
	} else {
		vm.fatal("creator not set")
	}
}

func (r ReturnStmt) Flow(g *graphBuilder) (head Step) {
	// reverse order to keep Eval correct
	for i := len(r.Results) - 1; i >= 0; i-- {
		each := r.Results[i]
		if i == len(r.Results)-1 {
			head = each.Flow(g)
			continue
		}
		each.Flow(g)
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
	return r.ReturnPos
}

func (r ReturnStmt) stmtStep() Evaluable { return r }

func (r ReturnStmt) String() string {
	return fmt.Sprintf("return(len=%d)", len(r.Results))
}
