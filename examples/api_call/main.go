package main

import (
	"fmt"

	"github.com/emicklei/gi"
)

func main() {
	pkg, _ := gi.ParseSource(`package main

import "fmt"

func Hello(name string) int {
	fmt.Println("Hello,", name)
	return 42
}
`)
	val, err := gi.Call(pkg, "Hello", "3i/Atlas")
	fmt.Println("returned:", val, "error:", err)
}
