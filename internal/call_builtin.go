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
	// Pop arguments from stack. They are in reverse order of declaration.
	args := make([]reflect.Value, len(c.Args))
	for i := range c.Args {
		args[i] = vm.frameStack.top().pop()
	}
	slice := args[0]
	elements := args[1:]

	// Special case: append to []byte from string or byte
	if slice.Type().Elem().Kind() == reflect.Uint8 {
		if sliceBytes, ok := slice.Interface().([]byte); ok {
			var canHandle = true
			var grow int
			for _, el := range elements {
				switch v := el.Interface().(type) {
				case string:
					grow += len(v)
				case byte:
					grow += 1
				default:
					canHandle = false
				}
				if !canHandle {
					break
				}
			}

			if canHandle {
				// Optimized append for []byte
				newSlice := append(sliceBytes, make([]byte, grow)...)
				offset := len(sliceBytes)
				for _, el := range elements {
					switch v := el.Interface().(type) {
					case string:
						offset += copy(newSlice[offset:], v)
					case byte:
						newSlice[offset] = v
						offset++
					}
				}
				vm.pushOperand(reflect.ValueOf(newSlice))
				return
			}
		}
	}

	// Fallback to generic reflect.Append
	result := reflect.Append(slice, elements...)
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
