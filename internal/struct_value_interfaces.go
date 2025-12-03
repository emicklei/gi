package internal

import (
	"fmt"
	"reflect"
)

type StructValueMethodDispatcher struct {
	vm  *VM
	val *StructValue
}

// Write implements io.Writer
func (d StructValueMethodDispatcher) Write(b []byte) (n int, err error) {
	console("StructValueMethodDispatcher.Write called")
	decl, ok := d.val.structType.methods["Write"]
	if !ok {
		return 0, fmt.Errorf("method Write not found")
	}
	d.vm.pushOperand(reflect.ValueOf(b))
	d.vm.takeAllStartingAt(decl.callGraph)
	reflectErr := d.vm.callStack.top().pop()
	reflectN := d.vm.callStack.top().pop()
	return int(reflectN.Int()), reflectErr.Interface().(error)
}
