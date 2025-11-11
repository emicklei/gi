package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	"github.com/fatih/structtag"
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
		typ := vm.returnsType(field.Type)
		for _, name := range field.Names {
			i.fields[name.Name] = reflect.New(typ).Elem()
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

func (i Instance) MarshalJSON() ([]byte, error) {
	m := map[string]any{}
	for fieldName, val := range i.fields {
		tagName, ok := i.tagFieldName("json", fieldName, val)
		if ok {
			m[tagName] = val.Interface()
		}
	}
	return json.Marshal(m)
}

func (i Instance) UnmarshalJSON(data []byte) error {
	m := map[string]any{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	for fieldName, val := range i.fields {
		tagName, ok := i.tagFieldName("json", fieldName, val)
		if ok {
			if val, ok := m[tagName]; ok {
				i.fields[fieldName] = reflect.ValueOf(val)
			}
		}
	}
	return nil
}

func (i Instance) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	// TODO
	return nil
}

// tagFieldName returns the name of the field as it should appear in JSON.
func (i Instance) tagFieldName(key string, fieldName string, fieldValue reflect.Value) (string, bool) {
	if !unicode.IsUpper(rune(fieldName[0])) {
		// unexported field
		return "", false
	}
	lit := i.Type.tagForField(fieldName)
	if lit == nil {
		return fieldName, true
	}
	unquoted, err := strconv.Unquote(lit.Value)
	if err != nil {
		return fieldName, false
	}
	tags, err := structtag.Parse(unquoted)
	if err != nil {
		return fieldName, false
	}
	jsonTag, err := tags.Get(key)
	if err != nil {
		return fieldName, false
	}
	if jsonTag.Name == "-" {
		return "", false
	}
	if jsonTag.HasOption("omitempty") {
		if fieldValue.IsZero() {
			return "", false
		}
	}
	return jsonTag.Name, true
}
