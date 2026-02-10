package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/emicklei/gi/pkg"
)

var dir = flag.String("dir", ".", "directory to run gistep on")

func main() {
	flag.Parse()

	gopkg, err := pkg.LoadPackage(*dir, nil)
	if err != nil {
		log.Fatal(err)
	}
	ipkg, err := pkg.BuildPackage(gopkg)
	if err != nil {
		log.Fatal(err)
	}
	runner := pkg.NewRunner(ipkg)
	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if b[0] == 'q' {
			os.Exit(0)
		}
		if err := runner.Step(); err != nil {
			log.Fatal(err)
		}
		fmt.Println(runner.Location())
	}
}
