package internal

import "testing"

func TestParseSource(t *testing.T) {
	source := `package main
func main() {
	print("gi")
}`
	pkg, err := ParseSource(source)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}
	if pkg.PkgPath != "main" {
		t.Errorf("expected package path 'main', got '%s'", pkg.PkgPath)
	}
}
