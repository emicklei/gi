package internal

import (
	"iter"
	"reflect"
)

func init() {
	stdfuncs["maps"] = map[string]reflect.Value{
		"Keys": reflect.ValueOf(func(a any) iter.Seq[any] {
			ra := reflect.ValueOf(a)
			if ra.Kind() != reflect.Map {
				return func(yield func(any) bool) {}
			}
			return func(yield func(any) bool) {
				for _, k := range ra.MapKeys() {
					if !yield(k.Interface()) {
						return
					}
				}
			}
		}),
	}
}
