package main

import (
	"fmt"

	"github.com/emicklei/gi/examples/subpkg/pkg"
)

func main() {
	fmt.Println(pkg.Name, pkg.IsWeekend("Sunday"))
}
