package internal

import (
	"fmt"
	"go/types"
	"reflect"
)

// Model: Gemini 3
// Prompt: in golang , how to get the reflect type from a given types.Type from the " go/types " package

// ToReflectType converts a go/types.Type to a reflect.Type.
// Note: Named structs will be converted to anonymous structs with the same layout.
func ToReflectType(t types.Type, valuesOrNil []reflect.Value) (reflect.Type, error) {
	switch t := t.(type) {
	case *types.Basic:
		return getBasicType(t)
	case *types.Pointer:
		elem, err := ToReflectType(t.Elem(), valuesOrNil)
		if err != nil {
			return nil, err
		}
		return reflect.PointerTo(elem), nil
	case *types.Slice:
		elem, err := ToReflectType(t.Elem(), valuesOrNil)
		if err != nil {
			return nil, err
		}
		return reflect.SliceOf(elem), nil
	case *types.Array:
		elemValues := valuesOrNil[0]
		elem, err := ToReflectType(t.Elem(), elemValues.Interface().([]reflect.Value))
		if err != nil {
			return nil, err
		}
		at := reflect.ArrayOf(int(t.Len()), elem)

		// new array as value
		nat := reflect.New(at).Elem()
		fmt.Println(nat.Interface())
		for i, v := range valuesOrNil {
			fmt.Println(v, v.Interface())
			nat.Index(i).Set(v)
		}

		return at, nil

	case *types.Map:
		key, err := ToReflectType(t.Key(), valuesOrNil)
		if err != nil {
			return nil, err
		}
		val, err := ToReflectType(t.Elem(), valuesOrNil)
		if err != nil {
			return nil, err
		}
		return reflect.MapOf(key, val), nil
	case *types.Struct:
		return createStructType(t)
	case *types.Named:
		// For named types, we generally want the underlying structure
		// because we cannot recreate the actual "Name" at runtime
		// unless that type is compiled into this binary.
		return ToReflectType(t.Underlying(), valuesOrNil)
	case *types.Interface:
		// Simply return the empty interface type for generic usage,
		// or reflect.TypeOf((*error)(nil)).Elem() etc.
		return reflect.TypeOf((*interface{})(nil)).Elem(), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", t)
	}
}

func createStructType(t *types.Struct) (reflect.Type, error) {
	fields := make([]reflect.StructField, t.NumFields())
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)
		typ, err := ToReflectType(f.Type(), nil)
		if err != nil {
			return nil, err
		}

		fields[i] = reflect.StructField{
			Name:    f.Name(),
			Type:    typ,
			Tag:     reflect.StructTag(t.Tag(i)),
			PkgPath: f.Pkg().Path(), // Important for unexported fields
		}

		// Handle embedded fields
		if f.Anonymous() {
			fields[i].Anonymous = true
		}
	}
	return reflect.StructOf(fields), nil
}

func getBasicType(b *types.Basic) (reflect.Type, error) {
	switch b.Kind() {
	case types.Bool:
		return reflect.TypeOf(false), nil
	case types.Int:
		return reflect.TypeOf(int(0)), nil
	case types.Int8:
		return reflect.TypeOf(int8(0)), nil
	case types.Int16:
		return reflect.TypeOf(int16(0)), nil
	case types.Int32:
		return reflect.TypeOf(int32(0)), nil
	case types.Int64:
		return reflect.TypeOf(int64(0)), nil
	case types.Uint:
		return reflect.TypeOf(uint(0)), nil
	case types.Uint8:
		return reflect.TypeOf(uint8(0)), nil
	case types.Uint16:
		return reflect.TypeOf(uint16(0)), nil
	case types.Uint32:
		return reflect.TypeOf(uint32(0)), nil
	case types.Uint64:
		return reflect.TypeOf(uint64(0)), nil
	case types.Float32:
		return reflect.TypeOf(float32(0)), nil
	case types.Float64:
		return reflect.TypeOf(float64(0)), nil
	case types.Complex64:
		return reflect.TypeOf(complex64(0)), nil
	case types.Complex128:
		return reflect.TypeOf(complex128(0)), nil
	case types.String:
		return reflect.TypeOf(""), nil
	case types.Uintptr:
		return reflect.TypeOf(uintptr(0)), nil
	case types.UnsafePointer:
		return reflect.TypeOf((*interface{})(nil)), nil // Approximate
	default:
		return nil, fmt.Errorf("unsupported basic kind: %v", b.Kind())
	}
}
