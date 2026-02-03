package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = Ident{}
var _ CanAssign = Ident{}

type Ident struct {
	namePos token.Pos
	name    string
}

func (i Ident) Eval(vm *VM) {
	vm.pushOperand(vm.localEnv().valueLookUp(i.name))
}

func (i Ident) assign(vm *VM, value reflect.Value) {
	if i.name == "_" {
		return
	}
	owner := vm.localEnv().valueOwnerOf(i.name)
	owner.set(i.name, value)
}
func (i Ident) define(vm *VM, value reflect.Value) {
	vm.localEnv().set(i.name, value)
}

func (i Ident) flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}

func (i Ident) Pos() token.Pos {
	return i.namePos
}

func (i Ident) String() string {
	return fmt.Sprintf("Ident(%s)", i.name)
}
