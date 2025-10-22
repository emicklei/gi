package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ CanAssign = ConstOrVar{}

type ConstOrVar struct {
	*ast.ValueSpec
	// for each Name in ValueSpec there is a ConstOrVar
	Name  *Ident
	Type  Expr
	Value Expr
}

func (v ConstOrVar) Assign(vm *VM, value reflect.Value) {
	vm.localEnv().valueOwnerOf(v.Name.Name).set(v.Name.Name, value)
}
func (v ConstOrVar) Define(vm *VM, value reflect.Value) {
	vm.localEnv().set(v.Names[0].Name, value)
}
func (v ConstOrVar) Declare(vm *VM) bool {
	if v.Value != nil {
		var val reflect.Value
		if vm.isStepping {
			val = vm.callStack.top().pop()
		} else {
			val = vm.returnsEval(v.Value)
		}
		if !val.IsValid() {
			return false
		}
		if val.Interface() == untypedNil {
			// if nil then zero
			if z, ok := v.Type.(HasZeroValue); ok {
				zv := z.ZeroValue(vm.localEnv())
				vm.localEnv().set(v.Name.Name, zv)
				return true
			}
		}
		vm.localEnv().set(v.Name.Name, val)
		return true
	}
	// if nil then zero
	if z, ok := v.Type.(HasZeroValue); ok {
		zv := z.ZeroValue(vm.localEnv())
		vm.localEnv().set(v.Name.Name, zv)
	}
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

func (v ConstOrVar) String() string {
	return fmt.Sprintf("ConstOrVar(%v)", v.Name.Name)
}
