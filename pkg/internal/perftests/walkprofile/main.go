package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/emicklei/gi"
	"github.com/emicklei/gi/pkg"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	count      = flag.Int("count", 1000, "number of iterations")
)

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
	p, _ := gi.ParseSource(src)
	for range *count {
		_, _ = pkg.CallPackageFunction(p, "main", nil, nil)
	}
}
