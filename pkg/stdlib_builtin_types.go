package pkg

import "reflect"

// https://pkg.go.dev/builtin
var builtinTypesMap = map[string]reflect.Type{}

func init() {
	builtinTypesMap["bool"] = reflect.TypeFor[bool]()
	builtinTypesMap["byte"] = reflect.TypeFor[byte]()
	builtinTypesMap["complex128"] = reflect.TypeFor[complex128]()
	builtinTypesMap["complex64"] = reflect.TypeFor[complex64]()
	builtinTypesMap["float32"] = reflect.TypeFor[float32]()
	builtinTypesMap["float64"] = reflect.TypeFor[float64]()
	builtinTypesMap["int"] = reflect.TypeFor[int]()
	builtinTypesMap["int8"] = reflect.TypeFor[int8]()
	builtinTypesMap["int16"] = reflect.TypeFor[int16]()
	builtinTypesMap["int32"] = reflect.TypeFor[int32]()
	builtinTypesMap["int64"] = reflect.TypeFor[int64]()
	builtinTypesMap["rune"] = reflect.TypeFor[rune]()
	builtinTypesMap["string"] = reflect.TypeFor[string]()
	builtinTypesMap["uint"] = reflect.TypeFor[uint]()
	builtinTypesMap["uint8"] = reflect.TypeFor[uint8]()
	builtinTypesMap["uint16"] = reflect.TypeFor[uint16]()
	builtinTypesMap["uint32"] = reflect.TypeFor[uint32]()
	builtinTypesMap["uint64"] = reflect.TypeFor[uint64]()
	builtinTypesMap["uintptr"] = reflect.TypeFor[uintptr]()

	var v any
	builtinTypesMap["any"] = reflect.TypeOf(&v).Elem()
	var e error
	builtinTypesMap["error"] = reflect.TypeOf(&e).Elem()
}
