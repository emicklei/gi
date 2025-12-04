package internal

import (
	"fmt"
	"reflect"
)

type StructValueWrapper struct {
	vm  *VM
	val *StructValue
}

// Write implements io.Writer
func (d StructValueWrapper) Write(b []byte) (n int, err error) {
	decl, ok := d.val.structType.methods["Write"]
	if !ok {
		return 0, fmt.Errorf("method Write not found")
	}
	// push arguments
	d.vm.pushOperand(reflect.ValueOf(b))
	d.vm.takeAllStartingAt(decl.callGraph)
	// pop results
	reflectErr := d.vm.callStack.top().pop()
	reflectN := d.vm.callStack.top().pop()
	// unreflect
	return int(reflectN.Int()), reflectErr.Interface().(error)
}
