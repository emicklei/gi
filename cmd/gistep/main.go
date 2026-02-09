package main

import (
	"flag"
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
	vm := pkg.NewVM(ipkg)
	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		print(string(b))
	}
}
