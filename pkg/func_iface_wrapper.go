package pkg

import "reflect"

type sdkInterfaceWrapper struct {
	vm       *VM
	receiver reflect.Value
	typ      HasMethods
}
