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
	if len(r.Results) == 0 {
		return
	}
	results := make([]reflect.Value, len(r.Results))
	for i := range r.Results {
		val := vm.frameStack.top().pop()
		results[i] = val
	}
	// bind result valutes to named results of the function if any
	fd := vm.frameStack.top().creator.(FuncDecl)
	ri := 0
	for _, fields := range fd.Type.Results.List {
		for _, name := range fields.Names {
			if name != nil && name.Name != "_" {
				vm.localEnv().set(name.Name, results[ri])
			}
			ri++
		}
	}

	// set return values for the top frame
	top := vm.frameStack.top()
	top.returnValues = results
	vm.frameStack.replaceTop(top)
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
