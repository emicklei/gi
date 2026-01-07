package pkg

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
		panic("int64 convert error")
	}
}
func toPtrInt64(a any) *int64 {
	if a == untypedNil {
		return (*int64)(nil)
	}
	switch v := a.(type) {
	case *int64:
		return v
	default:
		panic("*int64 convert error")
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
func toPtrBool(a any) *bool {
	switch v := a.(type) {
	case *bool:
		return v
	default:
		panic("*bool convert error")
	}
}
