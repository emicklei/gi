package internal

import "reflect"

var incDecFuncs = map[string]IncDecFunc{
	// int
	"inc_int":   func(v reflect.Value) reflect.Value { return reflect.ValueOf(int(v.Int() + 1)) },
	"inc_int8":  func(v reflect.Value) reflect.Value { return reflect.ValueOf(int8(v.Int() + 1)) },
	"inc_int16": func(v reflect.Value) reflect.Value { return reflect.ValueOf(int16(v.Int() + 1)) },
	"inc_int32": func(v reflect.Value) reflect.Value { return reflect.ValueOf(int32(v.Int() + 1)) },
	"inc_int64": func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Int() + 1) },
	"dec_int":   func(v reflect.Value) reflect.Value { return reflect.ValueOf(int(v.Int() - 1)) },
	"dec_int8":  func(v reflect.Value) reflect.Value { return reflect.ValueOf(int8(v.Int() - 1)) },
	"dec_int16": func(v reflect.Value) reflect.Value { return reflect.ValueOf(int16(v.Int() - 1)) },
	"dec_int32": func(v reflect.Value) reflect.Value { return reflect.ValueOf(int32(v.Int() - 1)) },
	"dec_int64": func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Int() - 1) },
	// uint
	"inc_uint":   func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint(v.Uint() + 1)) },
	"inc_uint8":  func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint8(v.Uint() + 1)) },
	"inc_uint16": func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint16(v.Uint() + 1)) },
	"inc_uint32": func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint32(v.Uint() + 1)) },
	"inc_uint64": func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Uint() + 1) },
	"dec_uint":   func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint(v.Uint() - 1)) },
	"dec_uint8":  func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint8(v.Uint() - 1)) },
	"dec_uint16": func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint16(v.Uint() - 1)) },
	"dec_uint32": func(v reflect.Value) reflect.Value { return reflect.ValueOf(uint32(v.Uint() - 1)) },
	"dec_uint64": func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Uint() - 1) },
	// float
	"inc_float32": func(v reflect.Value) reflect.Value { return reflect.ValueOf(float32(v.Float() + 1)) },
	"inc_float64": func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Float() + 1) },
	"dec_float32": func(v reflect.Value) reflect.Value { return reflect.ValueOf(float32(v.Float() - 1)) },
	"dec_float64": func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Float() - 1) },
}
