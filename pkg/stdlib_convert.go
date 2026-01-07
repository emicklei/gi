package pkg

import "fmt"

func toString(a any) string {
	switch v := a.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		panic(fmt.Sprintf("string convert undefined for %T", a))
	}
}

func toFloat32(a any) float32 {
	switch v := a.(type) {
	case float32:
		return v
	case float64:
		return float32(v)
	default:
		panic(fmt.Sprintf("float32 convert undefined for %T", a))
	}
}
func toFloat64(a any) float64 {
	switch v := a.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		panic(fmt.Sprintf("float64 convert undefined for %T", a))
	}
}

func toComplex64(a any) complex64 {
	switch v := a.(type) {
	case complex64:
		return v
	case complex128:
		return complex64(v)
	default:
		panic(fmt.Sprintf("complex64 convert undefined for %T", a))
	}
}
func toComplex128(a any) complex128 {
	switch v := a.(type) {
	case complex64:
		return complex128(v)
	case complex128:
		return v
	default:
		panic(fmt.Sprintf("complex128 convert undefined for %T", a))
	}
}

func toInt(a any) int {
	switch v := a.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		panic(fmt.Sprintf("int convert undefined for %T", a))
	}
}
func toInt8(a any) int8 {
	switch v := a.(type) {
	case int:
		return int8(v)
	case int8:
		return v
	case int16:
		return int8(v)
	case int32:
		return int8(v)
	case int64:
		return int8(v)
	case float64:
		return int8(v)
	default:
		panic(fmt.Sprintf("int8 convert undefined for %T", a))
	}
}
func toInt16(a any) int16 {
	switch v := a.(type) {
	case int:
		return int16(v)
	case int8:
		return int16(v)
	case int16:
		return v
	case int32:
		return int16(v)
	case int64:
		return int16(v)
	case float64:
		return int16(v)
	default:
		panic(fmt.Sprintf("int16 convert undefined for %T", a))
	}
}
func toInt32(a any) int32 {
	switch v := a.(type) {
	case int:
		return int32(v)
	case int8:
		return int32(v)
	case int16:
		return int32(v)
	case int32:
		return v
	case int64:
		return int32(v)
	case float64:
		return int32(v)
	default:
		panic(fmt.Sprintf("int32 convert undefined for %T", a))
	}
}
func toInt64(a any) int64 {
	switch v := a.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return int64(v)
	case float64:
		return int64(v)
	default:
		panic(fmt.Sprintf("int64 convert undefined for %T", a))
	}
}
func toUint(a any) uint {
	switch v := a.(type) {
	case int:
		return uint(v)
	case uint:
		return v
	case uint8:
		return uint(v)
	case uint16:
		return uint(v)
	case uint32:
		return uint(v)
	case uint64:
		return uint(v)
	case float64:
		return uint(v)
	default:
		panic(fmt.Sprintf("uint convert undefined for %T", a))
	}
}
func toUint8(a any) uint8 {
	switch v := a.(type) {
	case int:
		return uint8(v)
	case uint:
		return uint8(v)
	case uint8:
		return v
	case uint16:
		return uint8(v)
	case uint32:
		return uint8(v)
	case uint64:
		return uint8(v)
	case float64:
		return uint8(v)
	default:
		panic(fmt.Sprintf("uint8 convert undefined for %T", a))
	}
}

func toUint16(a any) uint16 {
	switch v := a.(type) {
	case int:
		return uint16(v)
	case uint:
		return uint16(v)
	case uint8:
		return uint16(v)
	case uint16:
		return v
	case uint32:
		return uint16(v)
	case uint64:
		return uint16(v)
	case float64:
		return uint16(v)
	default:
		panic(fmt.Sprintf("uint16 convert undefined for %T", a))
	}
}
func toUint32(a any) uint32 {
	switch v := a.(type) {
	case int:
		return uint32(v)
	case uint:
		return uint32(v)
	case uint8:
		return uint32(v)
	case uint16:
		return uint32(v)
	case uint32:
		return v
	case uint64:
		return uint32(v)
	case float64:
		return uint32(v)
	default:
		panic(fmt.Sprintf("uint32 convert undefined for %T", a))
	}
}
func toUint64(a any) uint64 {
	switch v := a.(type) {
	case int:
		return uint64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case float64:
		return uint64(v)
	default:
		panic(fmt.Sprintf("uint64 convert undefined for %T", a))
	}
}

func toBool(a any) bool {
	switch v := a.(type) {
	case bool:
		return v
	default:
		panic("bool convert error")
	}
}
