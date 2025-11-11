package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

/*
*
var _ Decl = ValueSpec{}

	type ValueSpec struct {
		CVs []ConstOrVar
	}

func (v ValueSpec) Eval(vm *VM) {} // noop

	func (v ValueSpec) Flow(g *graphBuilder) (head Step) {
		for i, cv := range v.CVs {
			cvFlow := cv.Flow(g)
			if i == 0 {
				head = cvFlow
			}
		}
		return head
	}

	func (v ValueSpec) Declare(vm *VM) bool {
		for _, cv := range v.CVs {
			if !cv.Declare(vm) {
				return false
			}
		}
		return true
	}

func (v ValueSpec) declStep() CanDeclare { return v }

	func (v ValueSpec) Pos() token.Pos {
		return token.NoPos // TODO
	}

	func (v ValueSpec) String() string {
		return fmt.Sprintf("ValueSpec(len=%d)", len(v.CVs))
	}

*
*/
var _ CanAssign = ConstOrVar{}

type ConstOrVar struct {
	*ast.ValueSpec
	// for each Name in ValueSpec there is a ConstOrVar
	NamePos token.Pos
	Name    *Ident
	Type    Expr
	Value   Expr
	// value flow
	callGraph Step
}

func (v ConstOrVar) ValueFlow() Step {
	return v.callGraph
}

func (v ConstOrVar) Assign(vm *VM, value reflect.Value) {
	vm.localEnv().valueOwnerOf(v.Name.Name).set(v.Name.Name, value)
}
func (v ConstOrVar) Define(vm *VM, value reflect.Value) {
	vm.localEnv().set(v.Name.Name, value)
}
func (v ConstOrVar) Declare(vm *VM) bool {
	if v.Value != nil {
		val := vm.frameStack.top().pop()
		if !val.IsValid() {
			return false
		}
		if val.Interface() == untypedNil {
			typ := vm.returnsType(v.Type)
			zv := reflect.Zero(typ)
			vm.localEnv().set(v.Name.Name, zv)
			return true
		}
		vm.localEnv().set(v.Name.Name, val)
		return true
	}
	// if nil then zero
	if z, ok := v.Type.(CanInstantiate); ok {
		zv := z.Instantiate(vm, 0, nil)
		vm.localEnv().set(v.Name.Name, zv)
		return true
	}
	typ := vm.returnsType(v.Type)
	zv := reflect.Zero(typ)
	vm.localEnv().set(v.Name.Name, zv)
	return true
}

func (v ConstOrVar) Eval(vm *VM) {
	vv := vm.localEnv().valueLookUp(v.Name.Name)
	vm.pushOperand(vv)
}

func (v ConstOrVar) Flow(g *graphBuilder) (head Step) {
	// only follow the value; the const or var itself is not a step
	if v.Value == nil {
		return g.current
	}
	head = v.Value.Flow(g)
	return
}

func (v ConstOrVar) declStep() CanDeclare { return v }

func (v ConstOrVar) Pos() token.Pos { return v.NamePos }

func (v ConstOrVar) String() string {
	return fmt.Sprintf("ConstOrVar(%v)", v.Name.Name)
}
