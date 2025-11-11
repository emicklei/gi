package internal

import (
	"reflect"
	"testing"
)

func TestParseSource(t *testing.T) {
	source := `package main
func main() {
	print("gi")
}`
	_, err := ParseSource(source)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}
}

func TestCallPackage(t *testing.T) {
	source := `package hello
import "fmt"
func Hello(name string) (int, string) {
	fmt.Println("Hello,", name)
	return 42, "World"
}
`
	pkg := buildPackage(t, source)
	results, err := CallPackageFunction(pkg, "Hello", []any{"3i/Atlas"}, nil)
	if err != nil {
		t.Fatalf("CallPackageFunction failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != 42 {
		t.Fatalf("expected result 42, got %v", results[0])
	}
	if results[1] != "World" {
		t.Fatalf("expected result 'World', got %v", results[1])
	}
}

func TestWriteAST(t *testing.T) {
	p := Package{Env: newPkgEnvironment(nil)}
	p.Env.set("a", reflect.ValueOf(1))
	p.writeAST("testgraphs/test.ast")
	// file must have: (int) 1
}
