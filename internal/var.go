package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Decl = ValueSpec{}
var _ CanDeclare = ValueSpec{}

// Const or Var declaration
type ValueSpec struct {
	NamePos   token.Pos
	Names     []*Ident
	Type      Expr
	Values    []Expr
	callGraph Step
}

func (v ValueSpec) declStep() CanDeclare { return v }

func (v ValueSpec) ValueFlow() Step {
	return v.callGraph
}

func (v ValueSpec) Declare(vm *VM) bool {
	if v.Type == nil {
		for _, idn := range v.Names {
			val := vm.frameStack.top().pop()
			if !val.IsValid() {
				// this happens when the value expression is referencing a undeclared variable
				return false
			}
			vm.localEnv().set(idn.Name, val)
		}
		return true
	}
	typ := vm.returnsType(v.Type)
	// left to right, see Flow
	for _, idn := range v.Names {
		if v.Values != nil {
			val := vm.frameStack.top().pop()
			if !val.IsValid() {
				return false
			}
			if val.Interface() == untypedNil {
				zv := reflect.Zero(typ)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			vm.localEnv().set(idn.Name, val)
		} else {
			// if nil then zero
			if z, ok := v.Type.(CanInstantiate); ok {
				zv := z.Instantiate(vm, 0, nil)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			// zero primitive
			zv := reflect.Zero(typ)
			vm.localEnv().set(idn.Name, zv)
		}
	}
	return true
}

func (v ValueSpec) Eval(vm *VM) {
	typ := vm.returnsType(v.Type)
	// left to right, see Flow
	for _, idn := range v.Names {
		if v.Values != nil {
			val := vm.frameStack.top().pop()
			if !val.IsValid() {
				return
			}
			if val.Interface() == untypedNil {
				zv := reflect.Zero(typ)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			vm.localEnv().set(idn.Name, val)
		} else {
			// if nil then zero
			if z, ok := v.Type.(CanInstantiate); ok {
				zv := z.Instantiate(vm, 0, nil)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			// zero primitive
			typ := vm.returnsType(v.Type)
			zv := reflect.Zero(typ)
			vm.localEnv().set(idn.Name, zv)
		}
	}
}

func (v ValueSpec) Flow(g *graphBuilder) (head Step) {
	if v.Values != nil {
		// reverse the order to have first value on top of stack
		for i := len(v.Values) - 1; i >= 0; i-- {
			valFlow := v.Values[i].Flow(g)
			if i == len(v.Values)-1 {
				head = valFlow
			}
		}
	}
	if head == nil {
		head = g.current
	}
	return
}

func (v ValueSpec) Pos() token.Pos {
	return v.NamePos
}

func (v ValueSpec) String() string {
	return fmt.Sprintf("ValueSpec(len=%d)", len(v.Names))
}
