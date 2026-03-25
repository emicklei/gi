package main

import (
	"reflect"

	"github.com/emicklei/dot"
	"github.com/emicklei/gi"
)

func init() {
	symbols := map[string]reflect.Value{
		"NewGraph": reflect.ValueOf(dot.NewGraph),
	}
	print("register dot")
	gi.RegisterPackage("github.com/emicklei/dot", symbols)
}
