package internal

import "reflect"

var binFuncs = map[string]BinaryExprFunc{
	"int64_12_int64": func(x, y reflect.Value) reflect.Value { // token.ADD == 12
		return reflect.ValueOf(x.Int() + y.Int())
	},
	"int_12_int": func(x, y reflect.Value) reflect.Value { // token.ADD == 12
		return reflect.ValueOf(int(x.Int() + y.Int()))
	},
}
