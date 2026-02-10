package pkg

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNilAny(t *testing.T) {
	// print(nil)
	println(untypedNil)
	fmt.Println(nil)
	fmt.Println(untypedNil)
}

func TestProgramTypeConvert(t *testing.T) {
	tests := []struct {
		typeName string
	}{
		{"int"},
		{"int8"},
		{"int16"},
		{"int32"},
		{"int64"},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.typeName, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`package main
			func main() {
				a := %s(1) + 2
				print(a)
			}`, tc.typeName)
			out := parseAndWalk(t, src)
			if got, want := out, "3"; got != want {
				t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		})
	}
}

func TestProgramTypeUnsignedConvert(t *testing.T) {
	tests := []struct {
		typeName string
	}{
		{"uint"},
		{"uint8"},
		{"uint16"},
		{"uint32"},
		{"uint64"},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.typeName, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`
			package main

			func main() {
				a := %s(1) + %s(2)
				print(a)
			}`, tc.typeName, tc.typeName)
			out := parseAndWalk(t, src)
			if got, want := out, "3"; got != want {
				t.Errorf("[step] got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		})
	}
}

func TestReflectCondition(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  reflect.Value
	}{
		{"true", true, reflectTrue},
		{"false", false, reflectFalse},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := reflectCondition(tc.input)
			if got.Kind() != reflect.Bool {
				t.Fatalf("reflectCondition(%v) kind = %v, want %v", tc.input, got.Kind(), reflect.Bool)
			}
			if got.Bool() != tc.want.Bool() {
				t.Fatalf("reflectCondition(%v) bool = %v, want %v", tc.input, got.Bool(), tc.want.Bool())
			}
			if got != tc.want {
				t.Fatalf("reflectCondition(%v) returned %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
