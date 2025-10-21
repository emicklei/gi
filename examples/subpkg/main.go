package main

import (
	"fmt"

	"subpkg/pkg"
)

func main() {
	fmt.Println(pkg.Name, pkg.IsWeekend("Sunday"))
}
