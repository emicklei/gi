package internal

func toInt64(a any) int64 {
	switch v := a.(type) {
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		panic("int64 convert error")
	}
}
