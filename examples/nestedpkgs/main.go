package main

import (
	"nestedpkgs/pkga"
	_ "nestedpkgs/pkgb"
	"nestedpkgs/pkgb/pkgc"
)

func main() {
	println("main refs", pkga.A)
	println("main refs", pkgc.C)
}
