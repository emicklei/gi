package main

import (
	"nestedpkgs/pkga"
	_ "nestedpkgs/pkgb"
	"nestedpkgs/pkgb/pkgc"
)

/*
init A
init B
pkgb refs A
init C
main refs A
main refs C
Printed A
Printed C
Printed B
*/
func main() {
	println("main refs", pkga.A)
	println("main refs", pkgc.C)
	pkga.Print()
	pkgc.Print()
}
