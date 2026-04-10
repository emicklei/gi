package pkg

import (
	"reflect"
	"testing"
)

func TestStructValueClone(t *testing.T) {
	fields := map[string]reflect.Value{
		"A": reflect.ValueOf(42),
		"B": reflect.ValueOf("hello"),
	}

	sv := StructValue{
		structType: new(StructType),
		fields:     &fields,
	}

	cloned := sv.clone()

	if cloned.structType != sv.structType {
		t.Errorf("expected structType %p, got %p", sv.structType, cloned.structType)
	}

	if cloned.fields == sv.fields {
		t.Errorf("expected fields map pointer to be different")
	}

	if len(*cloned.fields) != len(*sv.fields) {
		t.Errorf("expected %d fields, got %d", len(*sv.fields), len(*cloned.fields))
	}

	if (*cloned.fields)["A"].Int() != 42 {
		t.Errorf("expected field A to be 42, got %v", (*cloned.fields)["A"].Interface())
	}

	if (*cloned.fields)["B"].String() != "hello" {
		t.Errorf("expected field B to be 'hello', got %v", (*cloned.fields)["B"].Interface())
	}

	// Mutate original to ensure it's a deep copy of the map
	(*sv.fields)["A"] = reflect.ValueOf(100)

	if (*cloned.fields)["A"].Int() != 42 {
		t.Errorf("expected cloned field A to still be 42 after original was mutated, got %v", (*cloned.fields)["A"].Interface())
	}
}
