package mod

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func ParsePackageContent(pkgImportPath string) (*PackageContent, error) {
	// Use build to get the package details, which respects build tags.
	ctxt := build.Default
	ctxt.CgoEnabled = false
	ctxt.GOOS = runtime.GOOS
	ctxt.GOARCH = runtime.GOARCH
	pkg, err := ctxt.Import(pkgImportPath, "", 0)
	if err != nil {
		return nil, fmt.Errorf("could not import package %s: %v", pkgImportPath, err)
	}

	fset := token.NewFileSet()
	funcSet := make(map[string]struct{})
	typeSet := make(map[string]struct{})
	var pkgName string
	for _, file := range pkg.GoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, file), nil, 0)
		if err != nil {
			return nil, fmt.Errorf("could not parse file %s: %v", file, err)
		}
		if pkgName == "" {
			pkgName = f.Name.Name
		}

		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				isGeneric := fn.Type.TypeParams != nil && fn.Type.TypeParams.NumFields() > 0
				// Only include functions, not methods.
				if fn.Recv == nil && fn.Name.IsExported() && !isGeneric &&
					!strings.HasPrefix(fn.Name.Name, "Test") &&
					!strings.HasPrefix(fn.Name.Name, "Example") &&
					!strings.HasPrefix(fn.Name.Name, "Benchmark") {
					funcSet[fn.Name.Name] = struct{}{}
				}
			}
			if gd, ok := decl.(*ast.GenDecl); ok {
				if gd.Tok == token.TYPE {
					for _, spec := range gd.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							if ts.Name.IsExported() {
								isGeneric := ts.TypeParams != nil && ts.TypeParams.NumFields() > 0
								if _, ok := ts.Type.(*ast.StructType); ok && !isGeneric {
									typeSet[ts.Name.Name] = struct{}{}
								} else {
									//fmt.Printf("%s %T\n", ts.Name.Name, ts.Type)
								}
								if _, ok := ts.Type.(*ast.Ident); ok && !isGeneric {
									typeSet[ts.Name.Name] = struct{}{}
								}
							}
						}
					}
				}
				if gd.Tok == token.VAR || gd.Tok == token.CONST {
					for _, spec := range gd.Specs {
						if vs, ok := spec.(*ast.ValueSpec); ok {
							isGeneric := false
							if vs.Type != nil {
								if ft, ok := vs.Type.(*ast.FuncType); ok {
									if ft.TypeParams != nil && ft.TypeParams.NumFields() > 0 {
										isGeneric = true
									}
								}
							}
							if !isGeneric && len(vs.Values) > 0 {
								if fl, ok := vs.Values[0].(*ast.FuncLit); ok {
									if fl.Type.TypeParams != nil && fl.Type.TypeParams.NumFields() > 0 {
										isGeneric = true
									}
								}
							}
							if !isGeneric {
								for _, name := range vs.Names {
									if name.IsExported() {
										funcSet[name.Name] = struct{}{}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	var funcs []string
	for f := range funcSet {
		funcs = append(funcs, f)
	}
	sort.Strings(funcs)
	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)
	return &PackageContent{PkgName: pkgName, ImportPath: pkgImportPath, Values: funcs, Types: types}, nil
}
