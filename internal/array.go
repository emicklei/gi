package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = ArrayType{}

type ArrayType struct {
	Lbrack token.Pos // position of "["
	Len    Expr
	Elt    Expr
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
	eltType := vm.returnsType(a.Elt)
	if a.Len == nil {
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

func (a ArrayType) Pos() token.Pos { return a.Lbrack }

func (a ArrayType) String() string {
	return fmt.Sprintf("ArrayType(%v,slice=%v)", a.Elt, a.Len == nil)
}

// composite is (a reflect on) a Go array or slice
func (a ArrayType) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	if len(values) == 0 {
		return composite
	}
	fmt.Printf("composite: %#v %T\n", composite.Interface(), composite.Interface())
	// TODO optimize this
	elementType := composite.Type().Elem()
	fmt.Printf("element type: %#v\n", elementType)

	for i, v := range values {
		fmt.Println("value[i]", v.Interface())
		needConversion := elementType != reflect.TypeOf(v)
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

var _ Expr = SliceExpr{}

// http://golang.org/ref/spec#Slice_expressions
type SliceExpr struct {
	X      Expr      // expression
	Lbrack token.Pos // position of "["
	Low    Expr      // begin of slice range; or nil
	High   Expr      // end of slice range; or nil
	Max    Expr      // maximum capacity of slice; or nil
	// TODO handle this
	Slice3 bool // true if 3-index slice (2 colons present)
}

func (s SliceExpr) Eval(vm *VM) {
	// stack has max, high, low, x
	var high, low, x reflect.Value
	if s.Max != nil {
		// ignore max
		_ = vm.callStack.top().pop()
	}
	if s.High != nil {
		high = vm.callStack.top().pop()
	}
	if s.Low != nil {
		low = vm.callStack.top().pop()
	}
	var result reflect.Value
	x = vm.callStack.top().pop()
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

func (s SliceExpr) Pos() token.Pos { return s.Lbrack }

func (s SliceExpr) String() string {
	return fmt.Sprintf("SliceExpr(%v,%v:%v:%v)", s.X, s.Low, s.High, s.Max)
}
