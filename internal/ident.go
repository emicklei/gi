package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = Ident{}
var _ CanAssign = Ident{}

type Ident struct {
	NamePos token.Pos
	Name    string
}

func (i Ident) Eval(vm *VM) {
	if i.Name == "iota" {
		val := vm.iotaValue()
		vm.pushOperand(reflect.ValueOf(val))
		return
	}
	vm.pushOperand(vm.localEnv().valueLookUp(i.Name))
}

func (i Ident) Assigneable(vm *VM) CanAssign { return i }

func (i Ident) Assign(vm *VM, value reflect.Value) {
	if i.Name == "_" {
		return
	}
	owner := vm.localEnv().valueOwnerOf(i.Name)
	owner.set(i.Name, value)
}
func (i Ident) Define(vm *VM, value reflect.Value) {
	vm.localEnv().set(i.Name, value)
}

func (i Ident) Flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}

func (i Ident) Pos() token.Pos {
	return i.NamePos
}

func (i Ident) String() string {
	return fmt.Sprintf("Ident(%v)", i.Name)
}
