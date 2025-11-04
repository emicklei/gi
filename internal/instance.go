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
	panic("no such field or method: " + name)
}

func (i Instance) Assign(fieldName string, val reflect.Value) {
	if _, ok := i.fields[fieldName]; ok {
		// override, TODO what if HeapPointer?
		i.fields[fieldName] = val
		return
	}
	panic("no such field: " + fieldName)
}

// composite is (a reflect on) an Instance
func (i Instance) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	if len(values) == 0 {
		return composite
	}
	// check first element to decide keyed or not
	if _, ok := values[0].Interface().(KeyValue); ok {
		for _, each := range values {
			if kv, ok := each.Interface().(KeyValue); ok {
				i.fields[mustString(kv.Key)] = kv.Value
			}
		}
	} else {
		// unkeyed
		var fieldNames []string
		for _, field := range i.Type.Fields.List {
			for _, name := range field.Names {
				fieldNames = append(fieldNames, name.Name)
			}
		}
		for valueIndex, each := range values {
			if valueIndex < len(fieldNames) {
				i.fields[fieldNames[valueIndex]] = each
			}
		}
	}
	return composite
}
