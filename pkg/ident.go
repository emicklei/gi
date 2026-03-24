package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = Ident{}
var _ CanAssign = Ident{}

type Ident struct {
	name    string
	namePos token.Pos
}

func (i Ident) eval(vm *VM) {
	vm.pushOperand(vm.currentEnv().valueLookUp(i.name))
}

func (i Ident) assign(vm *VM, value reflect.Value) {
	if i.name == "_" {
		return
	}
	// if undeclared then set it
	owner, oldValue := vm.currentEnv().valueOwnerOf(i.name)
	if oldValue == reflectUndeclared {
		owner.valueSet(i.name, value)
		return
	}
	// if allocated on the heap then update the pointers value
	if hp, ok := oldValue.Interface().(*HeapPointer); ok {
		vm.heap.write(hp, value)
		return
	}
	owner.valueSet(i.name, value)
}
func (i Ident) define(vm *VM, value reflect.Value) {
	vm.currentEnv().valueSet(i.name, value)
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
