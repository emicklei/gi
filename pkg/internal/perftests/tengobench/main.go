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
	count      = flag.Int("count", 5, "number of iterations")
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

func fib(n int) int {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else {
		return fib(n-1) + fib(n-2)
	}
}

func main() {
	fib(35)
}
`
	p, _ := gi.ParseSource(src)
	for range *count {
		_, _ = pkg.CallPackageFunction(p, "main", nil, nil)
	}
}
