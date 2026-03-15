package pkgb

import "nestedpkgs/pkga"

var B = "B"

const BC = "BC"

func init() {
	println("init B")
	println("pkgb refs", pkga.A)
}
