package pkg

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
	d.vm.takeAllStartingAt(decl.graph)
	// pop results
	reflectErr := d.vm.popOperand()
	reflectN := d.vm.popOperand()
	// unreflect
	return int(reflectN.Int()), reflectErr.Interface().(error)
}
