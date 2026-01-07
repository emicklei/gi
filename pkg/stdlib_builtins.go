package pkg

import (
	"fmt"
	"reflect"
)

var (
	untypedNil        any
	reflectNil        = reflect.ValueOf(untypedNil)
	undeclaredNil     = struct{}{}
	reflectUndeclared = reflect.ValueOf(undeclaredNil)
	reflectTrue       = reflect.ValueOf(true)
	reflectFalse      = reflect.ValueOf(false)
)

// https://pkg.go.dev/builtin
var builtins = map[string]reflect.Value{
	"string": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[string](),
		prtZeroValue: reflect.ValueOf((*string)(nil)),
		convertFunc:  reflect.ValueOf(toString),
	}),
	"rune": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[rune](),
		prtZeroValue: reflect.ValueOf((*rune)(nil)),
	}),
	"bool": reflect.ValueOf(builtinType{
		typ:         reflect.TypeFor[bool](),
		convertFunc: reflect.ValueOf(toBool),
	}),
	"byte": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[byte](),
		prtZeroValue: reflect.ValueOf((*byte)(nil)),
		convertFunc:  reflect.ValueOf(toUint8),
	}),
	"complex128": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[complex128](),
		prtZeroValue: reflect.ValueOf((*complex128)(nil)),
		convertFunc:  reflect.ValueOf(toComplex128),
	}),
	"complex64": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[complex64](),
		prtZeroValue: reflect.ValueOf((*complex64)(nil)),
		convertFunc:  reflect.ValueOf(toComplex64),
	}),
	"float32": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[float32](),
		prtZeroValue: reflect.ValueOf((*float32)(nil)),
		convertFunc:  reflect.ValueOf(toFloat32),
	}),
	"float64": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[float64](),
		prtZeroValue: reflect.ValueOf((*float64)(nil)),
		convertFunc:  reflect.ValueOf(toFloat64),
	}),
	"int": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[int](),
		prtZeroValue: reflect.ValueOf((*int)(nil)),
		convertFunc:  reflect.ValueOf(toInt),
	}),
	"int8": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[int8](),
		prtZeroValue: reflect.ValueOf((*int8)(nil)),
		convertFunc:  reflect.ValueOf(toInt8),
	}),
	"int16": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[int16](),
		prtZeroValue: reflect.ValueOf((*int16)(nil)),
		convertFunc:  reflect.ValueOf(toInt16),
	}),
	"int32": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[int32](),
		prtZeroValue: reflect.ValueOf((*int32)(nil)),
		convertFunc:  reflect.ValueOf(toInt32),
	}),
	"int64": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[int64](),
		prtZeroValue: reflect.ValueOf((*int64)(nil)),
		convertFunc:  reflect.ValueOf(toInt64),
	}),
	"uint": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[uint](),
		prtZeroValue: reflect.ValueOf((*uint)(nil)),
		convertFunc:  reflect.ValueOf(toUint),
	}),
	"uint8": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[uint8](),
		prtZeroValue: reflect.ValueOf((*uint8)(nil)),
		convertFunc:  reflect.ValueOf(toUint8),
	}),
	"uint16": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[uint16](),
		prtZeroValue: reflect.ValueOf((*uint16)(nil)),
		convertFunc:  reflect.ValueOf(toUint16),
	}),
	"uint32": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[uint32](),
		prtZeroValue: reflect.ValueOf((*uint32)(nil)),
		convertFunc:  reflect.ValueOf(toUint32),
	}),
	"uint64": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[uint64](),
		prtZeroValue: reflect.ValueOf((*uint64)(nil)),
		convertFunc:  reflect.ValueOf(toUint64),
	}),
	"uintptr": reflect.ValueOf(builtinType{
		typ:          reflect.TypeFor[uintptr](),
		prtZeroValue: reflect.ValueOf((*uintptr)(nil)),
		convertFunc:  reflect.ValueOf(func(i uint) uintptr { return uintptr(i) }),
	}),
	"any": reflect.ValueOf(builtinType{
		typ:         reflect.TypeFor[any](),
		convertFunc: reflect.ValueOf(func(a any) any { return a }),
	}),
	"error": reflect.ValueOf(builtinType{
		typ:         reflect.TypeFor[error](), // reflect.TypeOf((*error)(nil)).Elem(),
		convertFunc: reflect.ValueOf(func(e any) error { return e.(error) }),
	}),

	// built-in functions implemented as builtinFunc
	"delete":  reflect.ValueOf(builtinFunc{name: "delete"}),
	"min":     reflect.ValueOf(builtinFunc{name: "min"}),
	"max":     reflect.ValueOf(builtinFunc{name: "max"}),
	"append":  reflect.ValueOf(builtinFunc{name: "append"}),
	"clear":   reflect.ValueOf(builtinFunc{name: "clear"}),
	"make":    reflect.ValueOf(builtinFunc{name: "make"}),
	"new":     reflect.ValueOf(builtinFunc{name: "new"}),
	"copy":    reflect.ValueOf(builtinFunc{name: "copy"}),
	"recover": reflect.ValueOf(builtinFunc{name: "recover"}),

	// built-in functions implemented as normal functions
	"imag":    reflect.ValueOf(func(c complex128) float64 { return imag(c) }),
	"real":    reflect.ValueOf(func(c complex128) float64 { return real(c) }),
	"len":     reflect.ValueOf(func(v any) int { return reflect.ValueOf(v).Len() }),
	"panic":   reflect.ValueOf(func(v any) { panic(v) }),
	"print":   reflect.ValueOf(func(args ...any) { fmt.Print(args...) }),
	"println": reflect.ValueOf(func(args ...any) { fmt.Println(args...) }),
	"cap":     reflect.ValueOf(func(v any) int { return reflect.ValueOf(v).Cap() }),

	// built-in values implemented as reflect.Value
	"true":  reflect.ValueOf(true), // not presented as Literal
	"nil":   reflectNil,
	"false": reflect.ValueOf(false), // not presented as Literal
}

type builtinType struct {
	// actual type in Go SDK
	typ reflect.Type
	// e.g. (*int64)(nil)
	prtZeroValue reflect.Value
	// need to store reflect.Value otherwise type info is lost
	convertFunc reflect.Value
	// TODO not sure if needed
	ptrConvertFunc reflect.Value
}

// reflectCondition converts a boolean to shared reflect.Value.
func reflectCondition(b bool) reflect.Value {
	if b {
		return reflectTrue
	}
	return reflectFalse
}
