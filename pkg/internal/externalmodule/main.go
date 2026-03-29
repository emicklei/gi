package main

import (
	"log"

	"github.com/emicklei/gi"
)

func main() {
	if err := gi.Run("exmod"); err != nil {
		log.Fatal(err)
	}
}
