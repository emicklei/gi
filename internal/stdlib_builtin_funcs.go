package internal

import (
	"fmt"
	"reflect"
)

var (
	untypedNil   = Nil{}
	reflectTrue  = reflect.ValueOf(true)
	reflectFalse = reflect.ValueOf(false)
)

// reflectCondition converts a boolean to shared reflect.Value.
func reflectCondition(b bool) reflect.Value {
	if b {
		return reflectTrue
	}
	return reflectFalse
}

type Nil struct{}

func (Nil) String() string { return "untyped nil" }

// https://pkg.go.dev/builtin
var builtinsMap = map[string]reflect.Value{
	"byte": reflect.ValueOf(func(i int) byte { return byte(i) }), // alias for uint8
	"*byte": reflect.ValueOf(func(a any) *byte {
		if a == untypedNil {
			return (*byte)(nil)
		}
		return a.(*byte)
	}),
	"cap":        reflect.ValueOf(func(v any) int { return reflect.ValueOf(v).Cap() }),
	"complex128": reflect.ValueOf(func(c complex64) complex128 { return complex128(c) }),
	"*complex128": reflect.ValueOf(func(a any) *complex128 {
		if a == untypedNil {
			return (*complex128)(nil)
		}
		return a.(*complex128)
	}),
	"complex64": reflect.ValueOf(func(c complex128) complex64 { return complex64(c) }),
	"*complex64": reflect.ValueOf(func(a any) *complex64 {
		if a == untypedNil {
			return (*complex64)(nil)
		}
		return a.(*complex64)
	}),
	"false":   reflect.ValueOf(false), // not presented as Literal
	"float32": reflect.ValueOf(func(f float64) float32 { return float32(f) }),
	"*float32": reflect.ValueOf(func(a any) *float32 {
		if a == untypedNil {
			return (*float32)(nil)
		}
		return a.(*float32)
	}),
	"float64": reflect.ValueOf(func(f float32) float64 { return float64(f) }),
	"*float64": reflect.ValueOf(func(a any) *float64 {
		if a == untypedNil {
			return (*float64)(nil)
		}
		return a.(*float64)
	}),
	"int": reflect.ValueOf(func(i int) int { return int(i) }),
	"*int": reflect.ValueOf(func(a any) *int {
		if a == untypedNil {
			return (*int)(nil)
		}
		return a.(*int)
	}),
	"int16": reflect.ValueOf(func(i int) int16 { return int16(i) }),
	"*int16": reflect.ValueOf(func(a any) *int16 {
		if a == untypedNil {
			return (*int16)(nil)
		}
		return a.(*int16)
	}),
	"int32": reflect.ValueOf(func(i int) int32 { return int32(i) }),
	"*int32": reflect.ValueOf(func(a any) *int32 {
		if a == untypedNil {
			return (*int32)(nil)
		}
		return a.(*int32)
	}),
	"int64": reflect.ValueOf(toInt64),
	"*int64": reflect.ValueOf(func(a any) *int64 {
		if a == untypedNil {
			return (*int64)(nil)
		}
		return a.(*int64)
	}),
	"int8": reflect.ValueOf(func(i int) int8 { return int8(i) }),
	"*int8": reflect.ValueOf(func(a any) *int8 {
		if a == untypedNil {
			return (*int8)(nil)
		}
		return a.(*int8)
	}),
	"imag":    reflect.ValueOf(func(c complex128) float64 { return imag(c) }),
	"len":     reflect.ValueOf(func(v any) int { return reflect.ValueOf(v).Len() }),
	"nil":     reflect.ValueOf(untypedNil),
	"panic":   reflect.ValueOf(func(v any) { panic(v) }),
	"print":   reflect.ValueOf(func(args ...any) { fmt.Print(args...) }),
	"println": reflect.ValueOf(func(args ...any) { fmt.Println(args...) }),
	"real":    reflect.ValueOf(func(c complex128) float64 { return real(c) }),
	"rune":    reflect.ValueOf(func(i int) rune { return rune(i) }), // alias for int32
	"*rune": reflect.ValueOf(func(a any) *rune {
		if a == untypedNil {
			return (*rune)(nil)
		}
		return a.(*rune)
	}),
	"string": reflect.ValueOf(func(b []byte) string { return string(b) }),
	"*string": reflect.ValueOf(func(a any) *string {
		if a == untypedNil {
			return (*string)(nil)
		}
		return a.(*string)
	}),
	"true": reflect.ValueOf(true), // not presented as Literal
	"uint": reflect.ValueOf(func(i int) uint { return uint(i) }),
	"*uint": reflect.ValueOf(func(a any) *uint {
		if a == untypedNil {
			return (*uint)(nil)
		}
		return a.(*uint)
	}),
	"uint16": reflect.ValueOf(func(i int) uint16 { return uint16(i) }),
	"*uint16": reflect.ValueOf(func(a any) *uint16 {
		if a == untypedNil {
			return (*uint16)(nil)
		}
		return a.(*uint16)
	}),
	"uint32": reflect.ValueOf(func(i int) uint32 { return uint32(i) }),
	"*uint32": reflect.ValueOf(func(a any) *uint32 {
		if a == untypedNil {
			return (*uint32)(nil)
		}
		return a.(*uint32)
	}),
	"uint64": reflect.ValueOf(func(i int) uint64 { return uint64(i) }),
	"*uint64": reflect.ValueOf(func(a any) *uint64 {
		if a == untypedNil {
			return (*uint64)(nil)
		}
		return a.(*uint64)
	}),
	"uint8": reflect.ValueOf(func(i int) uint8 { return uint8(i) }),
	"*uint8": reflect.ValueOf(func(a any) *uint8 {
		if a == untypedNil {
			return (*uint8)(nil)
		}
		return a.(*uint8)
	}),
	"uintptr": reflect.ValueOf(func(i uint) uintptr { return uintptr(i) }),

	// built-in functions implemented as builtinFunc
	"delete": reflect.ValueOf(builtinFunc{name: "delete"}),
	"min":    reflect.ValueOf(builtinFunc{name: "min"}),
	"max":    reflect.ValueOf(builtinFunc{name: "max"}),
	"append": reflect.ValueOf(builtinFunc{name: "append"}),
	"clear":  reflect.ValueOf(builtinFunc{name: "clear"}),
	"make":   reflect.ValueOf(builtinFunc{name: "make"}),
	"new":    reflect.ValueOf(builtinFunc{name: "new"}),
	"copy":   reflect.ValueOf(builtinFunc{name: "copy"}),
}
