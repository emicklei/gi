package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

type builtinFunc struct{ name string }

// https://pkg.go.dev/builtin#delete
func (c CallExpr) evalDelete(vm *VM) {
	target := vm.popOperand()
	key := vm.popOperand()
	target.SetMapIndex(key, reflect.Value{}) // delete
}

// https://pkg.go.dev/builtin#copy
func (CallExpr) evalCopy(vm *VM) {
	dest := vm.popOperand()
	src := vm.popOperand()
	n := reflect.Copy(dest, src)
	vm.pushOperand(reflect.ValueOf(n))
}

// https://pkg.go.dev/builtin#append
func (c CallExpr) evalAppend(vm *VM) {
	args := make([]reflect.Value, len(c.Args))
	for i := range c.Args {
		args[i] = vm.popOperand()
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
	mapOrSlice := vm.popOperand()
	mapOrSlice.Clear()
	return mapOrSlice
}

func (c CallExpr) evalMin(vm *VM) {
	var result reflect.Value
	if len(c.Args) == 2 {
		right := vm.popOperand() // first to last, see Flow
		left := vm.popOperand()
		less := BinaryExprValue{op: token.LSS, left: left, right: right}.Eval()
		if less.Bool() {
			result = left
		} else {
			result = right
		}
	} else {
		// 3
		third := vm.popOperand()
		second := vm.popOperand()
		first := vm.popOperand()
		less := BinaryExprValue{op: token.LSS, left: first, right: second}.Eval()
		if less.Bool() {
			result = first
		} else {
			result = second
		}
		less = BinaryExprValue{op: token.LSS, left: result, right: third}.Eval()
		if !less.Bool() {
			result = third
		}
	}
	vm.pushOperand(result)
}

func (c CallExpr) evalMax(vm *VM) {
	var result reflect.Value
	if len(c.Args) == 2 {
		right := vm.popOperand() // first to last, see Flow
		left := vm.popOperand()
		less := BinaryExprValue{op: token.LSS, left: left, right: right}.Eval()
		if less.Bool() {
			result = right
		} else {
			result = left
		}
	} else {
		// 3
		third := vm.popOperand()
		second := vm.popOperand()
		first := vm.popOperand()
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
	typ := vm.popOperand()
	length := 0
	if len(c.Args) > 1 {
		length = int(vm.popOperand().Int())
	}
	// TODO
	// if len(c.Args) > 2 {
	// 	capacity = int(vm.frameStack.top().pop().Int())
	// }
	if ci, ok := typ.Interface().(CanMake); ok {
		structVal := ci.Make(vm, length, nil)
		vm.pushOperand(structVal)
		return
	}
	vm.fatal(fmt.Sprintf("make: expected a CanInstantiate value:%v", typ))
}

func (c CallExpr) evalNew(vm *VM) {
	valWithType := vm.popOperand()
	typ := valWithType.Interface()
	if valWithType.Kind() == reflect.Struct {
		if ts, ok := typ.(TypeSpec); ok {
			structVal := ts.Make(vm, 0, nil)
			vm.pushOperand(structVal)
			return
		}
		// typ is an instance of a standard or imported external type
		rtype := reflect.TypeOf(typ)
		rval := reflect.New(rtype)
		vm.pushOperand(rval)
		return
	}
	vm.fatal(fmt.Sprintf("new: expected a CanMake value:%v", typ))
}

func (c CallExpr) evalRecover(vm *VM) {
	recoverVarName := internalVarName("recover", 0) // TODO make constant
	env := vm.localEnv().valueOwnerOf(recoverVarName)
	if env == nil {
		// nothing to recover from
		vm.pushOperand(reflectNil)
		return
	}
	val := env.valueLookUp(recoverVarName)
	env.unset(recoverVarName)
	vm.pushOperand(val)
}
