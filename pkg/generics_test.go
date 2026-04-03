package pkg

import "testing"

func TestDetector(t *testing.T) {
	defer debug(t)()
	source := `package main

import (
		"fmt"
		"slices"
)

func Even[T int | float64](num T) bool {
		return num/2 == 0
}

func main() {
		fmt.Println(slices.Contains([]int{1}, 1))
		fmt.Println(slices.Contains([]string{"gi"}, "gi"))
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

func TestGenericAsValue(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func Even[T int | float64](num T) bool {
		return num/2 == 0
}

func main() {
		ef := Even[float64]
		print(ef(2.0))
}`, "false")
}
