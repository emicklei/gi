package main

import "github.com/emicklei/gi"

func main() {
	pkg, _ := gi.ParseSource(`package main
import "fmt"
func Hello(name string) {
	fmt.Println("Hello,", name)
}
`)
	gi.Call(pkg, "Hello", "3i/Atlas")
}
