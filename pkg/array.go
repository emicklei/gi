package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = ArrayType{}

type ArrayType struct {
	lbrackPos token.Pos // position of "["
	len       Expr
	elt       Expr
}

// Eval creates and pushes an instance of the array or slice type onto the operand stack.
func (a ArrayType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(a))
}

func (a ArrayType) makeValue(vm *VM, size int, elements []reflect.Value) reflect.Value {
	if a.len != nil {
		len := vm.returnsEval(a.len)
		/// override size from Len expression unless Ellipsis
		if len.Kind() == reflect.Int {
			size = int(len.Int())
		}
	}
	eltType := vm.makeType(a.elt)
	if a.len == nil {
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

func (a ArrayType) flow(g *graphBuilder) (head Step) {
	g.next(a)
	return g.current
}

// composite is (a reflect on) a Go array or slice
func (a ArrayType) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	if len(values) == 0 {
		return composite
	}
	// TODO optimize this
	elementType := composite.Type().Elem()

	for i, v := range values {

		if elementType.Kind() == reflect.Array {
			composingElem := a.elt.(CanCompose)
			elemValues := v.Interface().([]reflect.Value)
			composingElem.literalCompose(vm, composite.Index(i), elemValues)
			continue
		}

		rv := reflect.TypeOf(v)
		needConversion := elementType != rv
		if needConversion {
			if v.CanConvert(elementType) {
				composite.Index(i).Set(v.Convert(elementType))
			}
		} else {
			composite.Index(i).Set(v)
		}
	}
	return composite

}

func (a ArrayType) Pos() token.Pos { return a.lbrackPos }

func (a ArrayType) String() string {
	return fmt.Sprintf("ArrayType(%v,slice=%v)", a.elt, a.len == nil)
}

var _ Expr = SliceExpr{}

// http://golang.org/ref/spec#Slice_expressions
type SliceExpr struct {
	x         Expr      // expression
	lbrackPos token.Pos // position of "["
	low       Expr      // begin of slice range; or nil
	high      Expr      // end of slice range; or nil
	max       Expr      // maximum capacity of slice; or nil
	// TODO handle this
	slice3 bool // true if 3-index slice (2 colons present)
}

func (s SliceExpr) Eval(vm *VM) {
	// stack has max, high, low, x
	var high, low, x reflect.Value
	if s.max != nil {
		// ignore max
		_ = vm.popOperand()
	}
	if s.high != nil {
		high = vm.popOperand()
	}
	if s.low != nil {
		low = vm.popOperand()
	}
	var result reflect.Value
	x = vm.popOperand()
	if low.IsValid() {
		if high.IsValid() {
			result = x.Slice(int(low.Int()), int(high.Int()))
		} else {
			result = x.Slice(int(low.Int()), x.Len())
		}
	}
	vm.pushOperand(result)
}

func (s SliceExpr) flow(g *graphBuilder) (head Step) {
	head = s.x.flow(g)
	if s.low != nil {
		s.low.flow(g)
	}
	if s.high != nil {
		s.high.flow(g)
	}
	if s.max != nil {
		s.max.flow(g)
	}
	g.next(s)
	return
}

func (s SliceExpr) Pos() token.Pos { return s.lbrackPos }

func (s SliceExpr) String() string {
	return fmt.Sprintf("SliceExpr(%v,%v:%v:%v)", s.x, s.low, s.high, s.max)
}
