package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = ArrayType{}

type ArrayType struct {
	*ast.ArrayType
	Len Expr
	Elt Expr
}

// used?
func (a ArrayType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(a))
}

func (a ArrayType) Instantiate(vm *VM) reflect.Value {
	eltTypeName := mustIdentName(a.Elt)
	eltType := vm.localEnv().typeLookUp(eltTypeName)
	if a.ArrayType.Len == nil {
		// slice
		sliceType := reflect.SliceOf(eltType)
		return reflect.MakeSlice(sliceType, 0, 4)
	} else {
		// array
		len := vm.returnsEval(a.Len)
		arrayType := reflect.ArrayOf(int(len.Int()), eltType)
		return reflect.New(arrayType)
	}
}

func (a ArrayType) Flow(g *graphBuilder) (head Step) {
	g.next(a)
	return g.current
}

func (a ArrayType) String() string {
	return fmt.Sprintf("ArrayType(%v,slice=%v)", a.Elt, a.ArrayType.Len == nil)
}

// composite is (a reflect on) a Go array or slice
func (a ArrayType) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	if a.ArrayType.Len == nil { // slice
		for _, v := range values {
			composite = reflect.Append(composite, v)
		}
		return composite
	}
	// array
	elem := composite.Elem()
	for i, v := range values {
		elem.Index(i).Set(v)
	}
	return composite
}
