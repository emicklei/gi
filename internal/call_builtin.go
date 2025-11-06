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

// https://pkg.go.dev/builtin#copy
func (CallExpr) evalCopy(vm *VM) {
	dest := vm.frameStack.top().pop()
	src := vm.frameStack.top().pop()
	n := reflect.Copy(dest, src)
	vm.pushOperand(reflect.ValueOf(n))
}

// https://pkg.go.dev/builtin#append
func (c CallExpr) evalAppend(vm *VM) {
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
	var result reflect.Value
	if len(c.Args) == 2 {
		right := vm.frameStack.top().pop() // first to last, see Flow
		left := vm.frameStack.top().pop()
		less := BinaryExprValue{op: token.LSS, left: left, right: right}.Eval()
		if less.Bool() {
			result = right
		} else {
			result = left
		}
	} else {
		// 3
		third := vm.frameStack.top().pop()
		second := vm.frameStack.top().pop()
		first := vm.frameStack.top().pop()
		less := BinaryExprValue{op: token.LSS, left: first, right: second}.Eval()
		if less.Bool() {
			result = second
		} else {
			result = first
		}
		less = BinaryExprValue{op: token.LSS, left: result, right: third}.Eval()
		if less.Bool() {
			result = third
		}
	}
	vm.pushOperand(result)
}

// https://go.dev/ref/spec#Making_slices_maps_and_channels
func (c CallExpr) evalMake(vm *VM) {
	// stack has 1,2, or 3 arguments, left to right
	typ := vm.frameStack.top().pop()
	length := 0
	if len(c.Args) > 1 {
		length = int(vm.frameStack.top().pop().Int())
	}
	// TODO
	// if len(c.Args) > 2 {
	// 	capacity = int(vm.frameStack.top().pop().Int())
	// }
	if ci, ok := typ.Interface().(CanInstantiate); ok {
		instance := ci.Instantiate(vm, length, nil)
		vm.pushOperand(instance)
		return
	}
	vm.fatal(fmt.Sprintf("make: expected a CanInstantiate value:%v", typ))
}

func (c CallExpr) evalNew(vm *VM) {
	valWithType := vm.frameStack.top().pop()
	typ := valWithType.Interface()
	if valWithType.Kind() == reflect.Struct {
		// typ is an instance of a standard or imported external type
		rtype := reflect.TypeOf(typ)
		rval := reflect.New(rtype)
		vm.pushOperand(rval)
		return
	}
	if ci, ok := typ.(CanInstantiate); ok {
		instance := ci.Instantiate(vm, 0, nil)
		vm.pushOperand(instance)
		return
	}
	vm.fatal(fmt.Sprintf("new: expected a CanInstantiate value:%v", typ))
}
