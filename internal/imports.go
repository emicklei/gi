package internal

import "reflect"

var importedPkgs = make(map[string]map[string]reflect.Value)

func RegisterPackage(pkgPath string, symbols map[string]reflect.Value) {
	// TODO check for override?
	importedPkgs[pkgPath] = symbols
}
