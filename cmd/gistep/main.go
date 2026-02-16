package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emicklei/gi/pkg"
)

var dir = flag.String("dir", "/Users/emicklei/Projects/gi/examples/nestedloop", "directory to run gistep on")

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
	runner := pkg.NewVM(ipkg)
	runner.Setup(ipkg, "main", nil)
	var b []byte = make([]byte, 1)
	fmt.Println("Press any key to step, 'q' to quit")
	for {
		os.Stdin.Read(b)
		if b[0] == 'q' {
			os.Exit(0)
		}
		if err := runner.Step(); err != nil {
			if err == io.EOF {
				return
			}
			log.Fatal(err)
		}
		fmt.Println(runner.Location())
	}
}
