package internal

import "testing"

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
	t.Skip()
	trace = true
	source := `package main
import "fmt"
func Hello(name string) int {
	fmt.Println("Hello,", name)
	return 42
}
`
	pkg := buildPackage(t, source)
	results, err := CallPackageFunction(pkg, "Hello", []any{"3i/Atlas"}, nil)
	if err != nil {
		t.Fatalf("CallPackageFunction failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != 42 {
		t.Fatalf("expected result 42, got %v", results[0])
	}
}
