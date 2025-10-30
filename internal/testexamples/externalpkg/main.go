package main

import (
	"fmt"

	"github.com/emicklei/dot"
)

func main() {
	g := dot.NewGraph()
	g.Node("gi")
	fmt.Println(g.String())
}
