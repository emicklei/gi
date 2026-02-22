package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"time"

	"golang.org/x/tools/go/packages"
)

var importedPkgs = make(map[string]map[string]reflect.Value)
var loadMode = packages.NeedName | packages.NeedSyntax | packages.NeedFiles | packages.NeedTypesInfo

func RegisterPackage(pkgPath string, symbols map[string]reflect.Value) {
	// TODO check for override?
	importedPkgs[pkgPath] = symbols
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
		if astFilename := os.Getenv("GI_AST"); astFilename != "" {
			writeGoAST(astFilename+".pkg", goPkg)
		}
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
	if astFilename := os.Getenv("GI_AST"); astFilename != "" {
		pkg.writeAST(astFilename)
	}
	return pkg, nil
}

func CallPackageFunction(pkg *Package, functionName string, args []any) ([]any, error) {
	return callPackageFunction(functionName, args, NewVM(pkg))
}

// only works if stepping = false
func callPackageFunction(functionName string, args []any, vm *VM) ([]any, error) {
	vm.takeAllStartingAt(vm.pkg.callGraph)

	// TODO maybe let the call do the lookup?
	fun := vm.pkg.env.valueLookUp(functionName)
	funValue := fun.Interface().(Func)
	if !fun.IsValid() {
		return nil, fmt.Errorf("%s function definition not found", functionName)
	}
	vm.pushNewFrame(funValue)
	defer vm.popFrame()

	// add noop expressions as arguments; the values will be pushed on the operand stack
	callArgs := make([]Expr, len(args))
	for i := range len(args) {
		callArgs[i] = noExpr{}
	}
	// make a CallExpr and reuse its logic to set up the call
	call := CallExpr{
		fun:  Ident{name: functionName},
		args: callArgs,
	}
	// push arguments as parameters on the operand stack, in reverse order
	for i := len(args) - 1; i >= 0; i-- {
		vm.pushOperand(reflect.ValueOf(args[i]))
	}
	call.handleFuncDecl(vm, funValue.(*FuncDecl))

	// collect non-reflection return values
	top := vm.currentFrame
	vals := []any{}
	results := funValue.results()
	if results != nil {
		for range len(results.List) {
			val := top.pop()
			vals = append(vals, val.Interface())
		}
	}
	return vals, nil
}

// ParseSource is a helper function that allows parsing and building a package directly from a source string, without needing to read from the filesystem.
// It creates a temporary directory, writes the source to a main.go file, and then uses the standard LoadPackage and BuildPackage functions to create the Package struct.
func ParseSource(source string) (*Package, error) {

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
	gopkg, err := LoadPackage(cfg.Dir, cfg)
	if err != nil {
		return nil, err
	}
	return BuildPackage(gopkg)
}
