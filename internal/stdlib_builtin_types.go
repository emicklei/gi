package internal

import "reflect"

// https://pkg.go.dev/builtin
var builtinTypesMap = map[string]reflect.Type{}

func init() {
	{
		var v bool
		builtinTypesMap["bool"] = reflect.TypeOf(v)
	}
	{
		var v any
		builtinTypesMap["any"] = reflect.TypeOf(&v).Elem()
	}
	{
		var v byte
		builtinTypesMap["byte"] = reflect.TypeOf(v)
	}
	{
		var v complex128
		builtinTypesMap["complex128"] = reflect.TypeOf(v)
	}
	{
		var v complex64
		builtinTypesMap["complex64"] = reflect.TypeOf(v)
	}
	{
		var v error
		builtinTypesMap["error"] = reflect.TypeOf(&v).Elem()
	}
	{
		var v float32
		builtinTypesMap["float32"] = reflect.TypeOf(v)
	}
	{
		var v float64
		builtinTypesMap["float64"] = reflect.TypeOf(v)
	}
	{
		var v int
		builtinTypesMap["int"] = reflect.TypeOf(v)
	}
	{
		var v int8
		builtinTypesMap["int8"] = reflect.TypeOf(v)
	}
	{
		var v int16
		builtinTypesMap["int16"] = reflect.TypeOf(v)
	}
	{
		var v int32
		builtinTypesMap["int32"] = reflect.TypeOf(v)
	}
	{
		var v int64
		builtinTypesMap["int64"] = reflect.TypeOf(v)
	}
	{
		var v rune
		builtinTypesMap["rune"] = reflect.TypeOf(v)
	}
	{
		var v string
		builtinTypesMap["string"] = reflect.TypeOf(v)
	}
	{
		var v uint
		builtinTypesMap["uint"] = reflect.TypeOf(v)
	}
	{
		var v uint8
		builtinTypesMap["uint8"] = reflect.TypeOf(v)
	}
	{
		var v uint16
		builtinTypesMap["uint16"] = reflect.TypeOf(v)
	}
	{
		var v uint32
		builtinTypesMap["uint32"] = reflect.TypeOf(v)
	}
	{
		var v uint64
		builtinTypesMap["uint64"] = reflect.TypeOf(v)
	}
	{
		var v uintptr
		builtinTypesMap["uintptr"] = reflect.TypeOf(v)
	}
}
