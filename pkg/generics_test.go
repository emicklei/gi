package pkg

import (
	"reflect"
	"slices"
	"testing"
)

func TestDetector(t *testing.T) {
	t.Skip()
	defer debug(t)()

	RegisterFunction(
		"slices",
		"Contains([]int,int)",
		reflect.ValueOf(func(a0 []int, a1 int) bool {
			return slices.Contains(a0, a1)
		}))
	RegisterFunction(
		"slices",
		"Contains([]string,string)",
		reflect.ValueOf(func(a0 []string, a1 string) bool {
			return slices.Contains(a0, a1)
		}))

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
	testMain(t, source, `
true
true
false`)
}

func TestGenericAsValue(t *testing.T) {
	testMain(t, `package main

func Even[T int | float64](num T) bool {
		return num/2 == 0
}

func main() {
		ef := Even[float64]
		print(ef(2.0))
}`, "false")
}
