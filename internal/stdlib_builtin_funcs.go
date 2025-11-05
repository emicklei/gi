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
	"byte":       reflect.ValueOf(func(i int) byte { return byte(i) }), // alias for uint8
	"cap":        reflect.ValueOf(func(v any) int { return reflect.ValueOf(v).Cap() }),
	"complex128": reflect.ValueOf(func(c complex64) complex128 { return complex128(c) }),
	"complex64":  reflect.ValueOf(func(c complex128) complex64 { return complex64(c) }),
	"false":      reflect.ValueOf(false), // not presented as Literal
	"float32":    reflect.ValueOf(func(f float64) float32 { return float32(f) }),
	"float64":    reflect.ValueOf(func(f float32) float64 { return float64(f) }),
	"int":        reflect.ValueOf(func(i int) int { return int(i) }),
	"int16":      reflect.ValueOf(func(i int) int16 { return int16(i) }),
	"int32":      reflect.ValueOf(func(i int) int32 { return int32(i) }),
	"int64":      reflect.ValueOf(func(i int) int64 { return int64(i) }),
	"int8":       reflect.ValueOf(func(i int) int8 { return int8(i) }),
	"imag":       reflect.ValueOf(func(c complex128) float64 { return imag(c) }),
	"len":        reflect.ValueOf(func(v any) int { return reflect.ValueOf(v).Len() }),
	"nil":        reflect.ValueOf(untypedNil),
	"panic":      reflect.ValueOf(func(v any) { panic(v) }),
	"print":      reflect.ValueOf(func(args ...any) { fmt.Print(args...) }),
	"println":    reflect.ValueOf(func(args ...any) { fmt.Println(args...) }),
	"real":       reflect.ValueOf(func(c complex128) float64 { return real(c) }),
	"rune":       reflect.ValueOf(func(i int) rune { return rune(i) }), // alias for int32
	"string":     reflect.ValueOf(func(b []byte) string { return string(b) }),
	"true":       reflect.ValueOf(true), // not presented as Literal
	"uint":       reflect.ValueOf(func(i int) uint { return uint(i) }),
	"uint16":     reflect.ValueOf(func(i int) uint16 { return uint16(i) }),
	"uint32":     reflect.ValueOf(func(i int) uint32 { return uint32(i) }),
	"uint64":     reflect.ValueOf(func(i int) uint64 { return uint64(i) }),
	"uint8":      reflect.ValueOf(func(i int) uint8 { return uint8(i) }),
	"uintptr":    reflect.ValueOf(func(i uint) uintptr { return uintptr(i) }),

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
