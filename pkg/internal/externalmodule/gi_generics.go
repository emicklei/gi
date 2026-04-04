package main

import (
	"reflect"
	"slices"

	"github.com/emicklei/gi"
)

func init() {
	gi.RegisterFunction(
		"slices",
		"Contains([]int,int)",
		reflect.ValueOf(func(a0 []int, a1 int) bool {
			return slices.Contains(a0, a1)
		}))
}
