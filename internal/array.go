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

// Eval creates and pushes an instance of the array or slice type onto the operand stack.
func (a ArrayType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(a))
}

func (a ArrayType) Instantiate(vm *VM, size int, constructorArgs []reflect.Value) reflect.Value {
	if a.Len != nil {
		len := vm.returnsEval(a.Len)
		/// override size from Len expression unless Ellipsis
		if len.Kind() == reflect.Int {
			size = int(len.Int())
		}
	}
	eltTypeName := mustIdentName(a.Elt)
	eltType := vm.localEnv().typeLookUp(eltTypeName)
	if a.ArrayType.Len == nil {
		// slice
		sliceType := reflect.SliceOf(eltType)
		return reflect.MakeSlice(sliceType, size, size)
	} else {
		// array
		arrayType := reflect.ArrayOf(size, eltType)
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
func (a ArrayType) LiteralCompose(composite reflect.Value, elementType reflect.Type, values []reflect.Value) reflect.Value {
	// TODO optimize this

	if a.ArrayType.Len == nil { // slice has the right length
		for i, v := range values {
			// TODO check if conversion is needed
			if v.CanConvert(elementType) {
				composite.Index(i).Set(v.Convert(elementType))
			} else {
				composite.Index(i).Set(v)
			}
		}
		return composite
	}
	// array
	for i, v := range values {
		if v.CanConvert(elementType) {
			composite.Index(i).Set(v.Convert(elementType))
		} else {
			composite.Index(i).Set(v)
		}
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
	var result reflect.Value
	x = vm.frameStack.top().pop()
	if low.IsValid() {
		if high.IsValid() {
			result = x.Slice(int(low.Int()), int(high.Int()))
		} else {
			result = x.Slice(int(low.Int()), x.Len())
		}
	}
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
