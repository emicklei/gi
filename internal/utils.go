package internal

import (
	"fmt"
	"reflect"
)

func expected(value any, expectation string) reflect.Value {
	panic(fmt.Sprintf("expected %s : %v (%T)", expectation, value, value))
}

func mustString(v reflect.Value) string {
	if !v.IsValid() {
		panic("value not valid as string")
	}
	if !v.CanInterface() {
		panic("cannot get interface for string")
	}
	s, ok := v.Interface().(string)
	if !ok {
		panic(fmt.Sprintf("expected string but got %T", v.Interface()))
	}
	return s
}

func mustIdentName(e Expr) string {
	if id, ok := e.(Ident); ok {
		return id.Name
	}
	if id, ok := e.(*Ident); ok {
		return id.Name
	}
	panic(fmt.Sprintf("expected Ident but got %T", e))
}

func pushCallResults(vm *VM, vals []reflect.Value) {
	// Push return values onto the operand stack in reverse order,
	// so the first return value ends up on top of the stack.
	for i := len(vals) - 1; i >= 0; i-- {
		vm.pushOperand(vals[i])
	}
}

// makeReflect is a helper function used in generated code to create reflect.Value of a type T.
func makeReflect[T any]() reflect.Value {
	var t T
	return reflect.ValueOf(t)
}
