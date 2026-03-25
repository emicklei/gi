package main

import (
	"os"

	"github.com/emicklei/gi"
)

func main() {
	content, _ := os.ReadFile("main.gi")
	pkg, _ := gi.Parse(string(content))
	_, _ = gi.Call(pkg, "main", nil)
}
