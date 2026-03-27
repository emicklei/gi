package gi

import (
	"reflect"

	"github.com/emicklei/gi/pkg"
)

// Loads and builds the Go package located at the specified file path.
// filePath is the file path to a folder that contains Go files.
func LoadPackage(filePath string) (*pkg.Package, error) {
	gopkg, err := pkg.LoadPackage(filePath, nil)
	if err != nil {
		return nil, err
	}
	return pkg.BuildPackage(gopkg)
}

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
	_, err = pkg.CallPackageFunction(p, "main", nil)
	return err
}

// Parse parses the provided Go source code string and returns a Package representation of it.
// The source must be valid Go, e.g. main package with a main function.
// It cannot have external dependencies ; only standard library packages are allowed.
func Parse(source string) (*pkg.Package, error) {
	return pkg.ParseSource(source)
}

// Call calls a function named funcName in the given package pkg with the provided parameters values.
// It returns the results of the function call and an error if any occurred during the call.
func Call(p *pkg.Package, funcName string, params ...any) ([]any, error) {
	return pkg.CallPackageFunction(p, funcName, params)
}

// RegisterPackage registers an external package with its values and types for use within gi-executed code.
// This function exist to support generated code and is not meant to be used beyond that.
func RegisterPackage(pkgPath string, values map[string]reflect.Value, types map[string]reflect.Type) {
	if pkgPath == "" || values == nil || types == nil {
		panic("pkgPath,values and types must be non-nil/empty")
	}
	pkg.RegisterPackage(pkgPath, values, types)
}
