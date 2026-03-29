package main

import (
	"fmt"
	"slices"

	"github.com/emicklei/dot"
)

func main() {
	g := dot.NewGraph()
	fmt.Println(g.String())
	fmt.Println(slices.Contains([]int{42}, 42))
}
