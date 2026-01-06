package gi

import (
	"reflect"

	"github.com/emicklei/gi/pkg"
)

// Run loads, builds, and runs the Go package located at the specified file path.
// filePath is the file path to a folder that contains a main.go file
// or any Go source file with a main function.
func Run(filePath string) error {
	gopkg, err := pkg.LoadPackage(filePath, nil)
	if err != nil {
		return err
	}
	p, err := pkg.BuildPackage(gopkg)
	if err != nil {
		return err
	}
	_, err = pkg.CallPackageFunction(p, "main", nil, nil)
	return err
}

// ParseSource parses the provided Go source code string and returns a Package representation of it.
// The source must be valid Go, e.g. main package with a main function.
// It cannot have external dependencies ; only standard library packages are allowed.
func ParseSource(source string) (*pkg.Package, error) {
	return pkg.ParseSource(source)
}

// Call calls a function named funcName in the given package pkg with the provided parameters values.
// It returns the results of the function call and an error if any occurred during the call.
func Call(p *pkg.Package, funcName string, params ...any) ([]any, error) {
	return pkg.CallPackageFunction(p, funcName, params, nil)
}

// RegisterPackage registers an external package with its symbols for use within gi-executed code.
func RegisterPackage(pkgPath string, symbols map[string]reflect.Value) {
	if pkgPath == "" || symbols == nil {
		panic("pkgPath and symbols must be non-nil/empty")
	}
	pkg.RegisterPackage(pkgPath, symbols)
}
