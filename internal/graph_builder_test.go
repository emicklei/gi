package internal

import (
	"testing"
)

func TestGrapherFor(t *testing.T) {
	t.Skip()
	source := `
package main

func main() {
	j := 1
	for i := 0; i < 3; i++ {
		j = i
		print(i)
	}
	print(j)
}`

	//trace = true)
	out := parseAndWalk(t, "testgraphs/TestGrapherFor.dot", source)
	expected := `0122`
	if out != expected {
		t.Fatalf("expected %q got %q", expected, out)
	}
}
