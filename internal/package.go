package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/emicklei/dot"
	"github.com/spewerspew/spew"
	"golang.org/x/tools/go/packages"
)

var importedPkgs = make(map[string]map[string]reflect.Value)

func RegisterPackage(pkgPath string, symbols map[string]reflect.Value) {
	// TODO check for override?
	importedPkgs[pkgPath] = symbols
}

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
		panic(fmt.Sprintf("package not found: %s %s", p.PkgPath, name))
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
	*packages.Package // TODO look for actual data used here
	Env               *PkgEnvironment
	Initialized       bool
}

func (p *Package) Select(name string) reflect.Value {
	return p.Env.valueLookUp(name)
}

func (p *Package) Initialize(vm *VM) error {
	if p.Initialized {
		return nil
	}
	p.Initialized = true

	// move methhods to types
	for _, decl := range p.Env.methods {
		structType := vm.localEnv().valueLookUp(decl.Recv.List[0].Type.(Ident).Name) // ugly
		console(decl)
		console(structType)
	}

	// try declare all of them until none left
	// a declare may refer to other unseen declares.
	done := false
	for !done {
		done = true
		for i, decl := range p.Env.declarations {
			if decl != nil {
				if p.Env.declarations[i].Declare(vm) {
					p.Env.declarations[i] = nil
				} else {
					done = false
				}
			}
		}
	}
	// then run all inits
	for _, each := range p.Env.inits {
		// TODO clean up
		call := CallExpr{
			Fun:  Ident{Name: "init"},
			Args: []Expr{}, // TODO for now, main only
		}
		call.handleFuncDecl(vm, each)
	}
	return nil
}

func (p *Package) writeAST(fileName string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SPEW: failed to write AST file", r)
		}
	}()
	buf := new(bytes.Buffer)
	spew.Config.DisableMethods = true
	spew.Config.MaxDepth = 4 // TODO see if this is enough
	done := make(chan struct{})
	go func() {
		// only dump the actual values of each var/function in the environment
		for _, v := range p.Env.Env.(*Environment).valueTable {
			spew.Fdump(buf, v.Interface())
		}
		done <- struct{}{}
	}()
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("AST writing took more than 2 seconds, aborting")
		close(done)
	case <-done:
	}
	os.WriteFile(fileName, buf.Bytes(), 0644)
}

func (p *Package) writeCallGraph(fileName string) {
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
	if p == nil || p.Package == nil {
		return "Package(<nil>)"
	}
	return fmt.Sprintf("Package(%s,%s)", p.Name, p.PkgPath)
}

var loadMode = packages.NeedName | packages.NeedSyntax | packages.NeedFiles | packages.NeedTypesInfo

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
	pkg := &Package{Package: goPkg, Env: b.env.(*PkgEnvironment)}
	if callGraphFilename := os.Getenv("GI_CALL"); callGraphFilename != "" {
		pkg.writeCallGraph(callGraphFilename)
	}
	if astFilename := os.Getenv("GI_AST"); astFilename != "" {
		pkg.writeAST(astFilename)
	}
	return pkg, nil
}

func CallPackageFunction(pkg *Package, functionName string, args []any, optionalVM *VM) ([]any, error) {
	var vm *VM
	if optionalVM != nil {
		vm = optionalVM
	} else {
		vm = NewVM(pkg.Env)
		vm.setFileSet(pkg.Fset)
	}
	for _, subpkg := range pkg.Env.packageTable {
		subvm := NewVM(subpkg.Env)
		subvm.setFileSet(pkg.Fset)
		if err := subpkg.Initialize(subvm); err != nil {
			return nil, fmt.Errorf("failed to initialize package %s: %v", subpkg.PkgPath, err)
		}
	}
	if err := pkg.Initialize(vm); err != nil {
		return nil, fmt.Errorf("failed to initialize package %s: %v", pkg.PkgPath, err)
	}
	// TODO maybe let the call do the lookup?
	fun := pkg.Env.valueLookUp(functionName)
	if !fun.IsValid() {
		return nil, fmt.Errorf("%s function definition not found", functionName)
	}
	// add noop expressions as arguments; the values will be pushed on the operand stack
	callArgs := make([]Expr, len(args))
	for i := range len(args) {
		callArgs[i] = noExpr{}
	}
	// make a CallExpr and reuse its logic to set up the call
	call := CallExpr{
		Fun:  Ident{Name: functionName},
		Args: callArgs,
	}
	// push arguments as parameters on the operand stack, in reverse order
	for i := len(args) - 1; i >= 0; i-- {
		vm.pushOperand(reflect.ValueOf(args[i]))
	}
	call.handleFuncDecl(vm, fun.Interface().(FuncDecl))

	// collect non-reflection return values
	top := vm.callStack.top()
	vals := []any{}
	results := fun.Interface().(FuncDecl).Type.Results
	if results != nil {
		for _, field := range results.List {
			if field.Names != nil {
				for range len(field.Names) {
					vals = append(vals, top.pop().Interface())
				}
			} else { // unnamed results
				vals = append(vals, top.pop().Interface())
			}
		}
	}
	return vals, nil
}

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
