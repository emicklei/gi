package main

import (
	"flag"

	"github.com/emicklei/gi"
	"github.com/pkg/profile"
)

var (
	memprofile = flag.String("memprofile", ".", "write memory profile to file")
	count      = flag.Int("count", 1000, "number of iterations")
)

func main() {
	flag.Parse()
	if *memprofile != "" {
		p := profile.Start(profile.MemProfile, profile.ProfilePath(*memprofile))
		defer func() {
			p.Stop()
		}()
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
	for range *count {
		gi.Call(pkg, "main", nil)
	}
}
