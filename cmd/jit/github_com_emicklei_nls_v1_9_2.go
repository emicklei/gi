package main

import (
	"reflect"

	"github.com/emicklei/dot"
	"github.com/emicklei/gi"
)

func init() {
	gi.RegisterPackage("github.com/emicklei/dot", map[string]reflect.Value{
		"NewGraph": reflect.ValueOf(dot.NewGraph),
	})
}
