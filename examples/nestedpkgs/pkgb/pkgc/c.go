package pkgc

import "nestedpkgs/pkgb"

var C = "C"

func init() {
	println("init C")
}

func Print() {
	println("Printed", C)
	pkgb.Print()
}
