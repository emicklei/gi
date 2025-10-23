package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = Ident{}
var _ CanAssign = Ident{}

type Ident struct {
	*ast.Ident
}

func (i Ident) CanEval(vm *VM) bool {
	return vm.localEnv().valueLookUp(i.Name).IsValid()
}

func (i Ident) Eval(vm *VM) {
	vm.pushOperand(vm.localEnv().valueLookUp(i.Name))
}
func (i Ident) Assign(vm *VM, value reflect.Value) {
	owner := vm.localEnv().valueOwnerOf(i.Name)
	if owner == nil {
		vm.fatal("undefined identifier: " + i.Name)
	}
	owner.set(i.Name, value)
}
func (i Ident) Define(vm *VM, value reflect.Value) {
	vm.localEnv().set(i.Name, value)
}

// ZeroValue returns the zero value iff the Ident represents a standard type.
func (i Ident) ZeroValue(env Env) reflect.Value {
	// TODO handle interpreted types.
	rt := env.typeLookUp(i.Name)
	if rt == nil { // invalid
		return reflect.Value{}
	}
	return reflect.Zero(rt)
}

func (i Ident) String() string {
	if i.Obj == nil {
		return fmt.Sprintf("Ident(%v)", i.Name)
	}
	return fmt.Sprintf("Ident(%v)", i.Obj.Name)
}

func (i Ident) Flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}
