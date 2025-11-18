package main

import (
	"os"

	"github.com/emicklei/gi"
)

func main() {
	if err := gi.Run("."); err != nil {
		print(err.Error())
		os.Exit(1)
	}
}
