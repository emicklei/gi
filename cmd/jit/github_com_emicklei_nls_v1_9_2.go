package main

import (
	"reflect"

	"github.com/emicklei/dot"
	"github.com/emicklei/gi"
)

// makeReflect is a helper function used in generated code to create reflect.Value of a type T.
func makeReflect[T any]() reflect.Value {
	var t T
	return reflect.ValueOf(t)
}

func init() {
	gi.RegisterPackage("github.com/emicklei/dot", map[string]reflect.Value{
		"NewGraph": reflect.ValueOf(dot.NewGraph),
		"Edge":     makeReflect[dot.Edge](),
	})
}
