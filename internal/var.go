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

func (v ValueSpec) CallGraph() Step {
	return v.callGraph
}

func (v ValueSpec) Declare(vm *VM) bool {
	if v.Type == nil {
		for _, idn := range v.Names {
			val := vm.callStack.top().pop()
			if val == reflectUndeclared {
				// this happens when the value expression is referencing an undeclared variable
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
			val := vm.callStack.top().pop()
			if !val.IsValid() { // TODO check undeclared?
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
			val := vm.callStack.top().pop()
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

var _ Expr = new(iotaExpr)

// represents successive untyped integer constants
type iotaExpr struct {
	pos        token.Pos
	count      int
	fixed      bool // set to true when iota value is overridden with explicit value
	fixedValue int
}

func (i *iotaExpr) value() int {
	if i.fixed {
		return i.fixedValue
	}
	return i.count
}

func (i *iotaExpr) next() int {
	if i.fixed {
		return i.fixedValue
	}
	i.count++
	return i.count
}
func (i *iotaExpr) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(i.value()))
}
func (i *iotaExpr) Flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}
func (i *iotaExpr) Pos() token.Pos {
	return i.pos
}
func (i *iotaExpr) String() string {
	return fmt.Sprintf("iota(%d)", i.value())
}
