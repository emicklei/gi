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

	//
}

var builtinTypesMap2 = map[string]reflect.Value{
	"bool": reflect.ValueOf(builtinType{
		typ:            reflect.TypeFor[bool](),
		convertFunc:    reflect.ValueOf(toBool),
		ptrConvertFunc: reflect.ValueOf(toPtrBool),
	}),
	"int64": reflect.ValueOf(builtinType{
		typ:            reflect.TypeFor[int64](),
		prtZeroValue:   reflect.ValueOf((*int64)(nil)),
		convertFunc:    reflect.ValueOf(toInt64),
		ptrConvertFunc: reflect.ValueOf(toPtrInt64),
	}),
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
