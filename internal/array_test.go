package internal

import (
	"reflect"
	"testing"
)

func TestReflectIntSlice(t *testing.T) {
	rt := reflect.TypeOf(0)
	st := reflect.SliceOf(rt)
	rs := reflect.MakeSlice(st, 0, 0)
	rs = reflect.Append(rs, reflect.ValueOf(1))
	t.Log(rs)

	// using Call
	// rf := reflect.ValueOf(func(params ...reflect.Value) []reflect.Value {
	// 	return params
	// })
	// args := []reflect.Value{rs, reflect.ValueOf(2)}
	// vals := rf.Call(args)
	// t.Log(vals)

	// Check the result
	// expected := []int{1, 2}
	// actual := vals.Interface().([]int)
	// if !reflect.DeepEqual(actual, expected) {
	// 	t.Errorf("expected %v, got %v", expected, actual)
	// }
}
