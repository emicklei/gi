package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

type valuesAndTypes struct {
	isGeneric map[string]bool // Set
	values    map[string]reflect.Value
	types     map[string]reflect.Value
}

var importedPkgs = make(map[string]valuesAndTypes)
var loadMode = packages.NeedName | packages.NeedSyntax | packages.NeedFiles | packages.NeedTypesInfo

func RegisterPackage(pkgPath string, values map[string]reflect.Value, types map[string]reflect.Type) {
	typesAsValues := map[string]reflect.Value{}
	// current design requires a registry of values for types
	for k, v := range types {
		typesAsValues[k] = reflect.ValueOf(v)
	}
	// TODO mutex
	importedPkgs[pkgPath] = valuesAndTypes{
		isGeneric: map[string]bool{},
		values:    values,
		types:     typesAsValues}
}

func RegisterFunction(pkgPath string, funcName string, fn reflect.Value) {
	// TODO mutex
	vant, ok := importedPkgs[pkgPath]
	if ok {
		// append/overwrite
		vant.values[funcName] = fn
	} else {
		vant = valuesAndTypes{
			isGeneric: map[string]bool{},
			values:    map[string]reflect.Value{funcName: fn},
			types:     map[string]reflect.Value{},
		}
	}
	// remember that it is a type parameterized function
	paramTypeIndex := strings.Index(funcName, "[")
	if paramTypeIndex != -1 {
		// key is without the type info
		vant.isGeneric[funcName[0:paramTypeIndex-1]] = true
	}
	importedPkgs[pkgPath] = vant
}

func LoadPackage(dir string, optionalConfig *packages.Config) (*packages.Package, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory must be specified")
	}
	var cfg *packages.Config
	if optionalConfig != nil {
		cfg = optionalConfig
	} else {
		cfg = &packages.Config{
			Mode: loadMode,
			Fset: token.NewFileSet(),
			Dir:  dir,
			// set the [parser.SkipObjectResolution] parser flag to disable syntactic object resolution
			ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
				return parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
			},
		}
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %v", err)
	}
	if count := packages.PrintErrors(pkgs); count > 0 {
		return nil, fmt.Errorf("errors during package loading: %d", count)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found")
	}
	return pkgs[0], nil
}

func BuildPackage(goPkg *packages.Package) (*Package, error) {
	if trace {
		now := time.Now()
		defer func() {
			fmt.Printf("pkg.build(%s) took %v\n", goPkg.PkgPath, time.Since(now))
		}()
	}
	b := newASTBuilder(goPkg)
	for _, stx := range goPkg.Syntax {
		for _, decl := range stx.Decls {
			b.Visit(decl)
			if b.buildErr != nil {
				// fail fast
				return nil, b.buildErr
			}
		}
	}
	pkg := &Package{Package: goPkg, env: b.env.(*PkgEnvironment)}

	// build and store package setup flow
	gb := newGraphBuilder(goPkg)
	pkg.callGraph = pkg.flow(gb)

	// reorganize methods so that they are associated with their interpreted types, rather than as standalone functions in the package environment
	if err := pkg.moveMethodsToInterpretedTypes(); err != nil {
		return pkg, err
	}

	// export if requested via env vars, for debugging
	if callGraphFilename := os.Getenv("GI_CALL"); callGraphFilename != "" {
		pkg.writeCallGraph(callGraphFilename)
	}
	return pkg, nil
}

func CallPackageFunction(pkg *Package, functionName string, args []any) ([]any, error) {
	return NewVM(pkg).callPackageFunction(functionName, args)
}

// ParseSource is a helper function that allows parsing and building a package directly from a source string, without needing to read from the filesystem.
// It creates a temporary directory, writes the source to a main.go file, and then uses the standard LoadPackage.
func ParseSource(source string) (*packages.Package, error) {

	// create a temp dir with a main.go file and go.mod
	dir, err := os.MkdirTemp("", "gi-temp-dir")
	if err != nil {
		return nil, err
	}
	mainFile := path.Join(dir, "main.go")
	err = os.WriteFile(mainFile, []byte(source), 0644)
	if err != nil {
		return nil, err
	}
	modeFile := path.Join(dir, "go.mod")
	err = os.WriteFile(modeFile, []byte("module tempmod\n go 1.25\n"), 0644)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir) // Clean up

	cfg := &packages.Config{
		Mode: loadMode,
		Fset: token.NewFileSet(),
		Dir:  dir,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			return parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
		},
	}
	return LoadPackage(cfg.Dir, cfg)
}
