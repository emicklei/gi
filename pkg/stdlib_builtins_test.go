package pkg

import (
	"fmt"
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
	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`package main
			func main() {
				a := %s(1) + 2
				print(a)
			}`, tt.typeName)
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
	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`
			package main

			func main() {
				a := %s(1) + %s(2)
				print(a)
			}`, tt.typeName, tt.typeName)
			out := parseAndWalk(t, src)
			if got, want := out, "3"; got != want {
				t.Errorf("[step] got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		})
	}
}
