package gi

import (
	"os"
	"reflect"

	"github.com/emicklei/gi/internal"
)

// Run loads, builds, and runs the Go package located at the specified file path.
// filePath is the file path to a folder that contains a main.go file
// or any Go source file with a main function.
func Run(filePath string) error {
	gopkg, err := internal.LoadPackage(filePath, nil)
	if err != nil {
		return err
	}
	ffpkg, err := internal.BuildPackage(gopkg, os.Getenv("GI_STEP") != "")
	if err != nil {
		return err
	}
	_, err = internal.RunPackageFunction(ffpkg, "main", nil, nil)
	return err
}

// ParseSource parses the provided Go source code string and returns a Package representation of it.
// The source must be a valid Go, e.g. main package with a main function.
// It cannot have external dependencies ; only standard library packages are allowed.
func ParseSource(source string) (*Package, error) {
	return internal.ParseSource(source)
}

// Package represents a Go package loaded and parsed by the gi library.
type Package = internal.Package

// Call calls a function named funcName in the given package pkg with the provided parameters values.
// It returns the results of the function call and an error if any occurred during the call.
func Call(pkg *Package, funcName string, params ...any) ([]any, error) {
	return internal.RunPackageFunction(pkg, funcName, params, nil)
}

// RegisterPackage registers an external package with its symbols for use within gi-executed code.
func RegisterPackage(pkgPath string, symbols map[string]reflect.Value) {
	if pkgPath == "" || symbols == nil {
		panic("pkgPath and symbols must be non-nil/empty")
	}
	internal.RegisterPackage(pkgPath, symbols)
}
