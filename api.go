package gi

import "github.com/emicklei/gi/internal"

func Run(absolutePath string) error {
	gopkg, err := internal.LoadPackage(absolutePath, nil)
	if err != nil {
		return err
	}
	ffpkg, err := internal.BuildPackage(gopkg, false)
	if err != nil {
		return err
	}
	return internal.RunPackageFunction(ffpkg, "main", nil)
}

func ParseSource(source string) (*internal.Package, error) {
	return internal.ParseSource(source)
}

// type Package = internal.Package

// func LoadPackage(absolutePath string) (*Package, error) {
// 	return nil, nil
// }

func Call(pkg *internal.Package, funcName string, args []any) ([]any, error) {
	internal.RunPackageFunction(pkg, funcName, nil)
	return nil, nil
}
