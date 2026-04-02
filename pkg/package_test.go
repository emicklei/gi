package pkg

import (
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
	results, err := CallPackageFunction(pkg, "Hello", []any{"3i/Atlas"})
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

func TestV2Package(t *testing.T) {
	testMain(t, `package main
import (
	"math/rand/v2"
)
func main() {
	print(rand.IntN(1))
}`, "0")
}

func TestBlankImport(t *testing.T) {
	testMain(t, `package main
import (
	_ "math/rand/v2"
)
func main() {
	print(1)
}`, "1")
}

func TestDotImport(t *testing.T) {
	testMain(t, `package main
import (
	. "fmt"
)
func main() {
	Println("gi")
}`, "")
}

func TestDetector(t *testing.T) {
	defer debug(t)()
	source := `package main
import "slices"
import "fmt"

func Even[T int | float64](num T) bool {
	return num/2 == 0
}

func main() {
	nums := []int{1}
	for { if true { slices.Contains(nums,1)}}
	fmt.Println(nums)
	fmt.Println(Even(3))
}
`
	goPkg, _ := ParseSource(source)
	d := newGenericsDetector(goPkg)
	for _, stx := range goPkg.Syntax {
		for _, decl := range stx.Decls {
			d.Visit(decl)
		}
	}
}
