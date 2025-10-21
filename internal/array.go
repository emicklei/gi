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
	typeName := mustIdentName(a.Elt)
	rt := builtinTypesMap[typeName]
	if a.ArrayType.Len == nil {
		st := reflect.SliceOf(rt)
		return reflect.MakeSlice(st, 0, 4)
	} else {
		size := vm.returnsEval(a.Len)
		st := reflect.ArrayOf(int(size.Int()), rt)
		pArray := reflect.New(st)
		return pArray.Elem()
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
	for i, v := range values {
		composite.Index(i).Set(v)
	}
	return composite
}
