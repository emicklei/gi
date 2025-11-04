package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

type builtinFunc struct{ name string }

// https://pkg.go.dev/builtin#delete
func (c CallExpr) evalDelete(vm *VM) {
	target := vm.frameStack.top().pop()
	key := vm.frameStack.top().pop()
	target.SetMapIndex(key, reflect.Value{}) // delete
}

// https://pkg.go.dev/builtin#append
func (c CallExpr) evalAppend(vm *VM) {
	args := make([]reflect.Value, len(c.Args))
	for i := range c.Args {
		v := vm.frameStack.top().pop()
		args[i] = v
	}
	// special case: append to []byte from string or byte
	if len(c.Args) > 0 {
		target := args[0].Interface()
		if bytes, ok := target.([]byte); ok {
			// allow both []byte and string to be appended
			for i := 1; i < len(args); i++ {
				if str, ok := args[i].Interface().(string); ok {
					bytes = append(bytes, str...)
				} else if b, ok := args[i].Interface().(byte); ok {
					bytes = append(bytes, b)
				} else {
					vm.fatal(fmt.Sprintf("append: cannot append %T to []byte", args[i].Interface()))
				}
			}
			vm.pushOperand(reflect.ValueOf(bytes))
			return
		}
	}
	result := reflect.Append(args[0], args[1:]...)
	vm.pushOperand(result)
}

// https://pkg.go.dev/builtin#clear
// It returns the cleared map or slice.
func (c CallExpr) evalClear(vm *VM) reflect.Value {
	mapOrSlice := vm.frameStack.top().pop()
	mapOrSlice.Clear()
	return mapOrSlice
}

func (c CallExpr) evalMin(vm *VM) {
	right := vm.frameStack.top().pop() // first to last, see Flow
	left := vm.frameStack.top().pop()
	result := BinaryExprValue{op: token.LSS, left: left, right: right}.Eval()
	if result.Bool() {
		result = left
	} else {
		result = right
	}
	vm.pushOperand(result)
}

func (c CallExpr) evalMax(vm *VM) {
	right := vm.frameStack.top().pop() // first to last, see Flow
	left := vm.frameStack.top().pop()
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

func (c CallExpr) evalNew(vm *VM) {
	valWithType := vm.frameStack.top().pop()
	typ := valWithType.Interface()
	if ci, ok := typ.(CanInstantiate); ok {
		instance := ci.Instantiate(vm)
		vm.pushOperand(instance)
		return
	}
	if valWithType.Kind() == reflect.Struct {
		// typ is an instance of a standard or imported external type
		rtype := reflect.TypeOf(typ)
		rval := reflect.New(rtype)
		vm.pushOperand(rval)
		return
	}
	vm.fatal(fmt.Sprintf("new: expected a CanInstantiate value:%v", typ))
}
