package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Stmt = ReturnStmt{}

type ReturnStmt struct {
	*ast.ReturnStmt
	Results []Expr
}

func (r ReturnStmt) stmtStep() Evaluable { return r }

func (r ReturnStmt) String() string {
	return fmt.Sprintf("return(len=%d)", len(r.Results))
}

func (r ReturnStmt) Eval(vm *VM) {
	// abort function body iteration
	// TEMPORARY use funcStack
	if !vm.isStepping {
		if len(vm.activeFuncStack) != 0 {
			vm.activeFuncStack.top().setDone()
		}
	}

	if len(r.Results) == 0 {
		return
	}
	results := make([]reflect.Value, len(r.Results))
	for i, each := range r.Results {
		var val reflect.Value
		if vm.isStepping {
			val = vm.callStack.top().pop()
		} else {
			val = vm.returnsEval(each)
		}
		results[i] = val
	}
	// set return values for the top frame
	top := vm.callStack.top()
	top.returnValues = results
	vm.callStack.replaceTop(top)
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
	g.nextStep(&returnStep{step: g.newStep(r)})
	// if nothing to return then returnStep is the head
	if head == nil {
		head = g.current
	}
	// no next step after return
	g.current = nil
	return
}
