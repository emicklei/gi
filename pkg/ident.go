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

func (i Ident) eval(vm *VM) {
	vm.pushOperand(vm.currentFrame.env.valueLookUp(i.name))
}

func (i Ident) assign(vm *VM, value reflect.Value) {
	if i.name == "_" {
		return
	}
	owner := vm.currentFrame.env.valueOwnerOf(i.name)
	owner.valueSet(i.name, value)
}
func (i Ident) define(vm *VM, value reflect.Value) {
	vm.currentFrame.env.valueSet(i.name, value)
}

func (i Ident) flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}

func (i Ident) pos() token.Pos {
	return i.namePos
}

func (i Ident) String() string {
	return fmt.Sprintf("Ident(%s)", i.name)
}
