package internal

import (
	"fmt"
	"reflect"
)

// first for struct
type Instance struct {
	Type   StructType
	fields map[string]reflect.Value
}

func NewInstance(vm *VM, t StructType) Instance {
	i := Instance{Type: t,
		fields: map[string]reflect.Value{},
	}
	for _, field := range t.Fields.List {
		for _, name := range field.Names {
			i.fields[name.Name] = reflect.Value{} // field.Type.ZeroValue(vm.localEnv())
		}
	}
	return i
}
func (i Instance) String() string {
	return fmt.Sprintf("Instance(%v)", i.Type)
}

func (i Instance) Select(name string) reflect.Value {
	if v, ok := i.fields[name]; ok {
		return v
	}
	panic("no such field: " + name)
}

// composite is (a reflect on) an Instance
func (i Instance) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	// fmt.Printf("%v (%T)", composite, composite)
	for _, each := range values {
		// fmt.Printf("%v (%T)\n", each, each)
		if kv, ok := each.Interface().(KeyValue); ok {
			i.fields[mustString(kv.Key)] = kv.Value
		}
	}
	return composite
}
