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
