package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var unaryFuncs map[string]UnaryExprFunc

func init() {
	unaryFuncs = map[string]UnaryExprFunc{
		fmt.Sprintf("bool%d", token.NOT): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(!v.Bool())
		},
		fmt.Sprintf("int%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int(-v.Int()))
		},
		fmt.Sprintf("int%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int(v.Int()))
		},
		fmt.Sprintf("int%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int(^v.Int()))
		},
		fmt.Sprintf("int8%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int8(-v.Int()))
		},
		fmt.Sprintf("int8%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int8(v.Int()))
		},
		fmt.Sprintf("int8%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int8(^v.Int()))
		},
		fmt.Sprintf("int16%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int16(-v.Int()))
		},
		fmt.Sprintf("int16%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int16(v.Int()))
		},
		fmt.Sprintf("int16%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int16(^v.Int()))
		},
		fmt.Sprintf("int32%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int32(-v.Int()))
		},
		fmt.Sprintf("int32%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int32(v.Int()))
		},
		fmt.Sprintf("int32%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(int32(^v.Int()))
		},
		fmt.Sprintf("int64%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(-v.Int())
		},
		fmt.Sprintf("int64%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(v.Int())
		},
		fmt.Sprintf("int64%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(^v.Int())
		},
		fmt.Sprintf("uint%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint(-v.Uint()))
		},
		fmt.Sprintf("uint%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint(v.Uint()))
		},
		fmt.Sprintf("uint%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint(^v.Uint()))
		},
		fmt.Sprintf("uint8%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint8(-v.Uint()))
		},
		fmt.Sprintf("uint8%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint8(v.Uint()))
		},
		fmt.Sprintf("uint8%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint8(^v.Uint()))
		},
		fmt.Sprintf("uint16%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint16(-v.Uint()))
		},
		fmt.Sprintf("uint16%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint16(v.Uint()))
		},
		fmt.Sprintf("uint16%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint16(^v.Uint()))
		},
		fmt.Sprintf("uint32%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint32(-v.Uint()))
		},
		fmt.Sprintf("uint32%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint32(v.Uint()))
		},
		fmt.Sprintf("uint32%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint32(^v.Uint()))
		},
		fmt.Sprintf("uint64%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint64(-v.Uint()))
		},
		fmt.Sprintf("uint64%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint64(v.Uint()))
		},
		fmt.Sprintf("uint64%d", token.XOR): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(uint64(^v.Uint()))
		},
		fmt.Sprintf("float32%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(float32(-v.Float()))
		},
		fmt.Sprintf("float32%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(float32(v.Float()))
		},
		fmt.Sprintf("float64%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(-v.Float())
		},
		fmt.Sprintf("float64%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(v.Float())
		},
		// untyped float
		fmt.Sprintf("float%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(-v.Float())
		},
		// untyped float
		fmt.Sprintf("float%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(v.Float())
		},
		fmt.Sprintf("complex64%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(complex64(-v.Complex()))
		},
		fmt.Sprintf("complex64%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(complex64(v.Complex()))
		},
		fmt.Sprintf("complex128%d", token.SUB): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(-v.Complex())
		},
		fmt.Sprintf("complex128%d", token.ADD): func(v reflect.Value) reflect.Value {
			return reflect.ValueOf(v.Complex())
		},
		fmt.Sprintf("chan%d", token.ARROW): func(v reflect.Value) reflect.Value {
			val, _ := v.Recv()
			return val
		},
	}
}
