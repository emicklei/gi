package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

type builtinFunc struct{ name string }

// https://pkg.go.dev/builtin#delete
func (c CallExpr) evalDelete(vm *VM) {
	target := vm.returnsEval(c.Args[0])
	key := vm.returnsEval(c.Args[1])
	target.SetMapIndex(key, reflect.Value{}) // delete
}

// https://pkg.go.dev/builtin#append
func (c CallExpr) evalAppend(vm *VM) {
	args := make([]reflect.Value, len(c.Args))
	for i, arg := range c.Args {
		args[i] = vm.returnsEval(arg)
	}
	result := reflect.Append(args[0], args[1:]...)
	vm.pushOperand(result)
}

// https://pkg.go.dev/builtin#clear
// It returns the cleared map or slice.
func (c CallExpr) evalClear(vm *VM) reflect.Value {
	var mapOrSlice reflect.Value
	if vm.isStepping {
		mapOrSlice = vm.callStack.top().pop()
	} else {
		mapOrSlice = vm.returnsEval(c.Args[0])
	}
	mapOrSlice.Clear()
	return mapOrSlice
}

func (c CallExpr) evalMin(vm *VM) {
	var left, right reflect.Value
	if vm.isStepping {
		right = vm.callStack.top().pop() // first to last, see Flow
		left = vm.callStack.top().pop()
	} else {
		left = vm.returnsEval(c.Args[0])
		right = vm.returnsEval(c.Args[1])
	}
	result := BinaryExprValue{op: token.LSS, left: left, right: right}.Eval()
	if result.Bool() {
		result = left
	} else {
		result = right
	}
	vm.pushOperand(result)
}

func (c CallExpr) evalMax(vm *VM) {
	var left, right reflect.Value
	if vm.isStepping {
		right = vm.callStack.top().pop() // first to last, see Flow
		left = vm.callStack.top().pop()
	} else {
		left = vm.returnsEval(c.Args[0])
		right = vm.returnsEval(c.Args[1])
	}
	result := BinaryExprValue{op: token.LSS, left: left, right: right}.Eval()
	if result.Bool() {
		result = right
	} else {
		result = left
	}
	vm.pushOperand(result)
}

func (c CallExpr) evalMake(vm *VM) {
	typ := vm.returnsEval(c.Args[0])
	if ci, ok := typ.Interface().(CanInstantiate); ok {
		instance := ci.Instantiate(vm)
		vm.pushOperand(instance)
		return
	}
	vm.fatal(fmt.Sprintf("make: expected a CanInstantiate value:%v", typ))
}
