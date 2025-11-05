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

func (a ArrayType) Instantiate(vm *VM, constructorArgs []reflect.Value) reflect.Value {
	eltTypeName := mustIdentName(a.Elt)
	eltType := vm.localEnv().typeLookUp(eltTypeName)
	if a.ArrayType.Len == nil {
		// slice
		sliceType := reflect.SliceOf(eltType)
		// optionally, the new slice can have a length
		size := reflect.ValueOf(0)
		if len(constructorArgs) > 0 {
			size = constructorArgs[0]
		}
		return reflect.MakeSlice(sliceType, int(size.Int()), int(size.Int()))
	} else {
		// array
		len := vm.returnsEval(a.Len)
		arrayType := reflect.ArrayOf(int(len.Int()), eltType)
		ptrArray := reflect.New(arrayType)
		return ptrArray.Elem()
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
	elem := composite
	for i, v := range values {
		elem.Index(i).Set(v)
	}
	return composite
}

var _ Expr = SliceExpr{}

// http://golang.org/ref/spec#Slice_expressions
type SliceExpr struct {
	*ast.SliceExpr
	X    Expr
	Low  Expr // may be nil
	High Expr // may be nil
	Max  Expr // may be nil
}

func (s SliceExpr) String() string {
	return fmt.Sprintf("SliceExpr(%v,%v:%v:%v)", s.X, s.Low, s.High, s.Max)
}

func (s SliceExpr) Eval(vm *VM) {
	// stack has max, high, low, x
	var high, low, x reflect.Value
	if s.Max != nil {
		// ignore max
		_ = vm.frameStack.top().pop()
	}
	if s.High != nil {
		high = vm.frameStack.top().pop()
	}
	if s.Low != nil {
		low = vm.frameStack.top().pop()
	}
	x = vm.frameStack.top().pop()
	result := x.Slice(int(low.Int()), int(high.Int()))
	vm.pushOperand(result)
}

func (s SliceExpr) Flow(g *graphBuilder) (head Step) {
	head = s.X.Flow(g)
	if s.Low != nil {
		s.Low.Flow(g)
	}
	if s.High != nil {
		s.High.Flow(g)
	}
	if s.Max != nil {
		s.Max.Flow(g)
	}
	g.next(s)
	return
}
