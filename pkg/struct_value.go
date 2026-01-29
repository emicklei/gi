package pkg

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"maps"
	"reflect"
	"unicode"

	"github.com/fatih/structtag"
)

var structValueType = reflect.TypeOf(StructValue{})

var _ fmt.Formatter = StructValue{}
var _ CanCompose = StructValue{}
var _ FieldAssignable = StructValue{}

// StructValue represents an instance of an interpreted struct.
type StructValue struct {
	structType StructType
	fields     map[string]reflect.Value
}

// InstantiateStructValue creates a new StructValue of the given StructType.
func InstantiateStructValue(vm *VM, t StructType) StructValue {
	i := StructValue{structType: t,
		fields: map[string]reflect.Value{},
	}
	for _, field := range t.Fields.List {
		typ := vm.makeType(field.Type)
		for _, name := range field.Names {
			i.fields[name.Name] = reflect.New(typ).Elem()
		}
	}
	return i
}

func (i StructValue) toString() string {
	return fmt.Sprintf("StructValue(%v)", i.structType)
}

// TODO maybe return extra bool for ok?
func (i StructValue) selectFieldOrMethod(name string) reflect.Value {
	if v, ok := i.fields[name]; ok {
		return v
	}
	if method, ok := i.structType.methods[name]; ok {
		return reflect.ValueOf(method)
	}
	panic("no such field or method: " + name)
}

func (i StructValue) fieldAssign(fieldName string, val reflect.Value) {
	if _, ok := i.fields[fieldName]; ok {
		// override, TODO what if HeapPointer?
		i.fields[fieldName] = val
		return
	}
	panic("no such field: " + fieldName)
}

// composite is (a reflect on) an StructValue
func (i StructValue) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	if len(values) == 0 {
		return composite
	}
	// check first element to decide keyed or not
	if _, ok := values[0].Interface().(keyValue); ok {
		for _, each := range values {
			if kv, ok := each.Interface().(keyValue); ok {
				i.fields[mustString(kv.Key)] = kv.Value
			}
		}
	} else {
		// unkeyed
		var fieldNames []string
		for _, field := range i.structType.Fields.List {
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

func (i StructValue) MarshalJSON() ([]byte, error) {
	m := map[string]any{}
	for fieldName, val := range i.fields {
		tagName, ok := i.tagFieldName("json", fieldName, val)
		if ok {
			m[tagName] = val.Interface()
		}
	}
	return json.Marshal(m)
}

func (i StructValue) UnmarshalJSON(data []byte) error {
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

func (i StructValue) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = i.structType.Name
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	for fieldName, val := range i.fields {
		tagName, ok := i.tagFieldName("xml", fieldName, val)
		if ok {
			elem := xml.StartElement{Name: xml.Name{Local: tagName}}
			if err := enc.EncodeElement(val.Interface(), elem); err != nil {
				return err
			}
		}
	}
	return enc.EncodeToken(start.End())
}

// TODO cache tagFieldName results in StructType
// tagFieldName returns the name of the field as it should appear in JSON.
func (i StructValue) tagFieldName(key string, fieldName string, fieldValue reflect.Value) (string, bool) {
	if !unicode.IsUpper(rune(fieldName[0])) {
		// unexported field
		return "", false
	}
	lit := i.structType.tagForField(fieldName)
	if lit == nil {
		return fieldName, true
	}
	unquoted := *lit
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
	if jsonTag.HasOption("omitempty") || jsonTag.HasOption("omitzero") {
		if fieldValue.IsZero() {
			return "", false
		}
	}
	return jsonTag.Name, true
}

func (i StructValue) Format(f fmt.State, verb rune) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%s{", i.structType.Name)
	c := 0
	for fieldName, val := range i.fields {
		if unicode.IsUpper(rune(fieldName[0])) {
			if c > 0 {
				fmt.Fprint(&buf, ", ")
			}
			fmt.Fprint(&buf, fieldName, ":")
			formatFieldValue(&buf, verb, val.Interface())
			c++
		}
	}
	fmt.Fprint(&buf, "}")
	f.Write(buf.Bytes())
}

func (i StructValue) clone() StructValue {
	return StructValue{
		structType: i.structType,
		fields:     maps.Clone(i.fields),
	}
}

func formatFieldValue(w io.Writer, verb rune, val any) {
	if s, ok := val.(string); ok {
		fmt.Fprintf(w, "%q", s)
		return
	}
	format(w, verb, val)
}
