package pkg

import (
	"reflect"
	"testing"
)

func Generic[T any](arg T) (*T, error) { return &arg, nil }

// instantiations of Generic
func Generic_string(arg string) (*string, error) { return &arg, nil }

func TestCallGenericByReflect(t *testing.T) {
	rv := reflect.ValueOf(Generic_string)
	arg := "hello"
	rvArg := reflect.ValueOf(arg)
	results := rv.Call([]reflect.Value{rvArg})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].IsNil() {
		t.Fatal("result pointer should not be nil")
	}
	sPtr := results[0].Interface().(*string)
	if *sPtr != arg {
		t.Errorf("expected result to be %q, got %q", arg, *sPtr)
	}
	if !results[1].IsNil() {
		t.Errorf("expected error to be nil, got %v", results[1].Interface())
	}
}

func TestConvertStringPointer(t *testing.T) {
	var c *string
	ct := reflect.TypeOf(c)
	if ct.Kind() == reflect.Pointer {
		t.Log("pointer to", ct.Elem().Kind())
	}

	var s any = "hello"
	sPtr := &s
	rv := reflect.ValueOf(sPtr)
	if ct.Kind() == reflect.Pointer && rv.Elem().Kind() == reflect.Interface {
		innerValue := rv.Elem().Elem()

		if innerValue.CanAddr() {
			ptrToString := innerValue.Addr()
			t.Log(ptrToString)
		} else {
			// If the string inside wasn't addressable, make a new one
			newPtr := reflect.New(innerValue.Type())
			newPtr.Elem().Set(innerValue)
			t.Log("Converted to *string:", newPtr.Interface())
		}
	}
}
