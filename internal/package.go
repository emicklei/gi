package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"time"

	"github.com/emicklei/dot"
	"golang.org/x/tools/go/packages"
)

type StandardPackage struct {
	Name    string
	PkgPath string
	// TODO currently separate tables for types and other symbols
	symbolTable map[string]reflect.Value // const,var,func, not types
	typesTable  map[string]reflect.Value // not reflect.Type to make Select work uniformly
}

func (p StandardPackage) Select(name string) reflect.Value {
	v, ok := p.symbolTable[name]
	if !ok {
		t, ok := p.typesTable[name]
		if ok {
			return t
		}
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
		// for i, each := range p.Env.declarations {
		// 	if each != nil {
		// 		if each.Declare(vm) {
		// 			p.Env.declarations[i] = nil
		// 			done = false
		// 		}
		// 	}
		// }
		for i, decl := range p.Env.declarations {
			if decl != nil {
				vm.takeAll(decl.ValueFlow())
				// val := vm.frameStack.top().peek(0)
				// if val.IsValid() {
				// 	p.Env.declarations[i].(ConstOrVar).Declare(vm) // TODO CanDeclare -> CanDefine?
				// 	p.Env.declarationFlows[i] = nil
				// 	done = false
				// }
				if p.Env.declarations[i].Declare(vm) {
					p.Env.declarations[i] = nil
					done = false
				}
			}
		}
	}
	// then run all inits
	for _, each := range p.Env.inits {
		// TODO clean up
		call := CallExpr{
			Fun:  Ident{Ident: &ast.Ident{Name: "init"}},
			Args: []Expr{}, // TODO for now, main only
		}
		call.handleFuncDecl(vm, each)
	}
	return nil
}

func (p *Package) writeDotGraph(fileName string) {
	g := dot.NewGraph(dot.Directed)
	g.NodeInitializer(func(n dot.Node) {
		n.Box()
		n.Attr("fillcolor", "#EBFAFF") // https://htmlcolorcodes.com/
		n.Attr("style", "filled")
	})
	// for each function in the package create a subgraph
	values := p.Env.Env.(*Environment).valueTable
	for k, v := range values {
		if funDecl, ok := v.Interface().(FuncDecl); ok {
			if funDecl.callGraph == nil {
				continue
			}
			sub := g.Subgraph(k, dot.ClusterOption{})
			visited := map[int]dot.Node{}
			funDecl.callGraph.Traverse(sub, visited)
		}
	}
	os.WriteFile(fileName, []byte(g.String()), 0644)
}

func (p *Package) String() string {
	return fmt.Sprintf("Package(%s,%s)", p.Name, p.PkgPath)
}

func LoadPackage(dir string, optionalConfig *packages.Config) (*packages.Package, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory must be specified")
	}
	if trace {
		now := time.Now()
		defer func() {
			fmt.Printf("pkg.load(%s) took %v\n", dir, time.Since(now))
		}()
	}
	var cfg *packages.Config
	if optionalConfig != nil {
		cfg = optionalConfig
	} else {
		cfg = &packages.Config{
			Mode: packages.NeedName | packages.NeedSyntax | packages.NeedFiles | packages.NeedTypesInfo,
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

func BuildPackageFromAST(ast *ast.File, isStepping bool) (*Package, error) {
	if trace {
		now := time.Now()
		defer func() {
			fmt.Printf("pkg.buildFromAST(%s) took %v\n", ast.Name.Name, time.Since(now))
		}()
	}
	goPkg := &packages.Package{
		ID: "main", Name: ast.Name.Name, PkgPath: "main",
	}
	b := newASTBuilder(goPkg)
	b.opts = buildOptions{callGraph: isStepping}
	for _, imp := range ast.Imports {
		b.Visit(imp)
	}
	for _, decl := range ast.Decls {
		b.Visit(decl)
	}
	return &Package{Package: goPkg, Env: b.env.(*PkgEnvironment)}, nil
}

// TODO build options
func BuildPackage(goPkg *packages.Package, isStepping bool) (*Package, error) {
	if trace {
		now := time.Now()
		defer func() {
			fmt.Printf("pkg.build(%s) took %v\n", goPkg.PkgPath, time.Since(now))
		}()
	}
	b := newASTBuilder(goPkg)
	b.opts = buildOptions{callGraph: isStepping}
	for _, stx := range goPkg.Syntax {
		for _, decl := range stx.Decls {
			b.Visit(decl)
		}
	}
	pkg := &Package{Package: goPkg, Env: b.env.(*PkgEnvironment)}
	if dotFilename := os.Getenv("GI_DOT"); dotFilename != "" {
		pkg.writeDotGraph(dotFilename)
	}
	return pkg, nil
}

func RunPackageFunction(pkg *Package, functionName string, args []any, optionalVM *VM) ([]any, error) {
	var vm *VM
	if optionalVM != nil {
		vm = optionalVM
	} else {
		vm = newVM(pkg.Env)
	}
	for _, subpkg := range pkg.Env.packageTable {
		subvm := newVM(subpkg.Env)
		if err := subpkg.Initialize(subvm); err != nil {
			return nil, fmt.Errorf("failed to initialize package %s: %v", subpkg.PkgPath, err)
		}
	}
	if err := pkg.Initialize(vm); err != nil {
		return nil, fmt.Errorf("failed to initialize package %s: %v", pkg.PkgPath, err)
	}
	fun := pkg.Env.valueLookUp(functionName)
	if !fun.IsValid() {
		return nil, fmt.Errorf("%s function definition not found", functionName)
	}

	// collect parameter values
	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		params[i] = reflect.ValueOf(arg)
	}

	// create stack frame with parameters values
	fundecl := fun.Interface().(FuncDecl)
	vm.pushNewFrame(fundecl)
	frame := vm.frameStack.top()
	setParametersToFrame(fundecl.Type, params, vm, frame)

	// compose a CallExpr to leverage existing call handling
	call := CallExpr{
		Fun:  Ident{Ident: &ast.Ident{Name: functionName}},
		Args: []Expr{}, // TODO for now, main only
	}
	if trace {
		vm.traceEval(call)
	} else {
		call.Eval(vm)
	}

	// collect non-reflection return values
	top := vm.frameStack.top()
	results := make([]any, len(top.returnValues))
	for i, rv := range top.returnValues {
		if rv.CanInterface() {
			results[i] = rv.Interface()
		} else {
			results[i] = nil
		}
	}
	vm.popFrame()
	return results, nil
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

	// compose a CallExpr to leverage existing call handling
	newFunction(pkg, functionName, nil, vm)

	return nil
}

// TODO collect return values
func newFunction(pkg *Package, functionName string, args []any, vm *VM) {
	call := CallExpr{
		Fun:  Ident{Ident: &ast.Ident{Name: functionName}},
		Args: []Expr{}, // TODO for now, main only
	}
	g := newGraphBuilder(pkg.Package)
	vm.takeAll(call.Flow(g))
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
