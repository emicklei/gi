package internal

import (
	"reflect"
	"testing"
)

func TestStackframe(t *testing.T) {
	var sf stackFrame
	sf.push(reflect.ValueOf(42))
	sf.push(reflect.ValueOf("hello"))

	v1 := sf.pop()
	if v1.Interface() != "hello" {
		t.Errorf("expected 'hello', got %v", v1.Interface())
	}
	v2 := sf.pop()
	if v2.Interface() != 42 {
		t.Errorf("expected 42, got %v", v2.Interface())
	}
}
