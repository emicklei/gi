package main

import (
	"fmt"

	"github.com/emicklei/dot"
)

func main() {
	g := dot.NewGraph()
	fmt.Println(g.String())
}
