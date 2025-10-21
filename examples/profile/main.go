package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/emicklei/gi"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	src := `package main

func foo(a, b int) int {
	return a + b
}

func main() {
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if i == j {
				foo(i, j)
			} else if i < j {
				foo(j, i)
			} else {
				foo(i, j)
			}
		}
	}
}`
	pkg, _ := gi.ParseSource(src)
	for i := 0; i < 1000; i++ {
		_, _ = gi.Call(pkg, "main", nil)
	}
}
