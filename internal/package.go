package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"time"

	"golang.org/x/tools/go/packages"
)

type StandardPackage struct {
	Name        string
	PkgPath     string
	symbolTable map[string]reflect.Value
}

func (p StandardPackage) Select(name string) reflect.Value {
	v, ok := p.symbolTable[name]
	if !ok {
		return reflect.Value{}
	}
	return v
}
func (p StandardPackage) String() string {
	return fmt.Sprintf("StandardPackage(%s,%s)", p.Name, p.PkgPath)
}

type ExternalPackage struct {
	StandardPackage
}

func (p ExternalPackage) String() string {
	return fmt.Sprintf("ExternalPackage(%s,%s)", p.Name, p.PkgPath)
}

// TODO rename to LocalPackage?
type Package struct {
	*packages.Package
	Env         *PkgEnvironment
	Initialized bool
}

func (p *Package) Select(name string) reflect.Value {
	return p.Env.valueLookUp(name)
}

func (p *Package) Initialize(vm *VM) error {
	if p.Initialized {
		return nil
	}
	p.Initialized = true
	// first run const and vars
	// try declare all of them until none left
	// a declare may refer to other unseen declares.
	done := false
	for !done {
		done = true
		for i, each := range p.Env.declarations {
			if each != nil {
				if each.Declare(vm) {
					p.Env.declarations[i] = nil
					done = false
				}
			}
		}
	}
	// then run all inits
	for _, each := range p.Env.inits {
		vm.pushNewFrame(each)
		if trace {
			vm.traceEval(each)
		} else {
			each.Eval(vm)
		}
		vm.popFrame()
	}
	return nil
}

func (p *Package) String() string {
	return fmt.Sprintf("Package(%s,%s)", p.Name, p.PkgPath)
}

func LoadPackage(dir string, optionalConfig *packages.Config) (*packages.Package, error) {
	if trace {
		now := time.Now()
		defer func() {
			fmt.Printf("LoadPackage(%s) took %v\n", dir, time.Since(now))
		}()
	}
	var cfg *packages.Config
	if optionalConfig != nil {
		cfg = optionalConfig
	} else {
		cfg = &packages.Config{
			Mode: packages.NeedName | packages.NeedSyntax | packages.NeedFiles,
			Fset: token.NewFileSet(),
			Dir:  dir,
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

func BuildPackageFromAST(ast *ast.File, isStepping bool) (*Package, error) {
	if trace {
		now := time.Now()
		defer func() {
			fmt.Printf("BuildPackageFromAST(%s) took %v\n", ast.Name.Name, time.Since(now))
		}()
	}
	b := newStepBuilder()
	b.opts = buildOptions{callGraph: isStepping}
	for _, imp := range ast.Imports {
		b.Visit(imp)
	}
	for _, decl := range ast.Decls {
		b.Visit(decl)
	}
	return &Package{Package: &packages.Package{
		ID: "main", Name: ast.Name.Name, PkgPath: "main",
	}, Env: b.env.(*PkgEnvironment)}, nil
}

// TODO build options
func BuildPackage(pkg *packages.Package, dotFilename string, isStepping bool) (*Package, error) {
	if trace {
		now := time.Now()
		defer func() {
			fmt.Printf("BuildPackage(%s) took %v\n", pkg.PkgPath, time.Since(now))
		}()
	}
	b := newStepBuilder()
	b.opts = buildOptions{callGraph: isStepping, dotFilename: dotFilename}
	for _, stx := range pkg.Syntax {
		for _, decl := range stx.Decls {
			b.Visit(decl)
		}
	}
	return &Package{Package: pkg, Env: b.env.(*PkgEnvironment)}, nil
}

func RunPackageFunction(pkg *Package, functionName string, optionalVM *VM) error {
	var vm *VM
	if optionalVM != nil {
		vm = optionalVM
	} else {
		vm = newVM(pkg.Env)
	}
	for _, subpkg := range pkg.Env.packageTable {
		subvm := newVM(subpkg.Env)
		if err := subpkg.Initialize(subvm); err != nil {
			return fmt.Errorf("failed to initialize package %s: %v", subpkg.PkgPath, err)
		}
	}
	if err := pkg.Initialize(vm); err != nil {
		return fmt.Errorf("failed to initialize package %s: %v", pkg.PkgPath, err)
	}
	fun := pkg.Env.valueLookUp(functionName)
	if !fun.IsValid() {
		return fmt.Errorf("%s function definition not found", functionName)
	}
	// TODO

	fundecl := fun.Interface().(FuncDecl)
	vm.pushNewFrame(fundecl)
	fundecl.Eval(vm)
	vm.popFrame()
	return nil
}

func WalkPackageFunction(pkg *Package, functionName string, optionalVM *VM) error {

	// TODO code duplication
	var vm *VM
	if optionalVM != nil {
		vm = optionalVM
	} else {
		vm = newVM(pkg.Env)
	}
	for _, subpkg := range pkg.Env.packageTable {
		subvm := newVM(subpkg.Env)
		if err := subpkg.Initialize(subvm); err != nil {
			return fmt.Errorf("failed to initialize package %s: %v", subpkg.PkgPath, err)
		}
	}
	if err := pkg.Initialize(vm); err != nil {
		return fmt.Errorf("failed to initialize package %s: %v", pkg.PkgPath, err)
	}
	fun := pkg.Env.valueLookUp(functionName)
	if !fun.IsValid() {
		return fmt.Errorf("%s function definition not found", functionName)
	}
	decl := fun.Interface().(FuncDecl)

	// run it step by step
	vm.takeAll(decl.callGraph)
	return nil
}

func ParseSource(source string) (*Package, error) {
	ast, err := parser.ParseFile(token.NewFileSet(), "main.go", source, 0)
	if err != nil {
		return nil, err
	}
	ffpkg, err := BuildPackageFromAST(ast, true)
	if err != nil {
		return nil, err
	}
	return ffpkg, nil
}
