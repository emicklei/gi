package pkg

import (
	"fmt"
	"go/token"
	"os"
	"reflect"

	"github.com/emicklei/dot"
	"golang.org/x/tools/go/packages"
)

var _ CanSelect = (*Package)(nil)

type Package struct {
	*packages.Package // TODO look for actual data used here
	env               *PkgEnvironment
	initialized       bool
	callGraph         Step // the setup flow: declarations and calling inits
}

func (p *Package) selectByName(name string) reflect.Value {
	return p.env.valueLookUp(name)
}

func (p *Package) flow(g *graphBuilder) (head Step) {
	// first all subpackages, recursively, then resolve declarations, then inits
	for _, subpkg := range p.env.packageTable {
		subFlow := subpkg.flow(g)
		if head == nil {
			head = subFlow
		}
		g.nextStep(subFlow)
	}
	if len(p.env.declarations) > 0 {
		// use for statement to eval all declarations until all are declared, then move to inits
		doneVarName := internalVarName("done", g.idgen)
		doneVar := Ident{name: doneVarName}
		falseLit := newBasicLit(token.NoPos, reflectFalse)
		initDone := AssignStmt{
			tok:    token.DEFINE,
			tokPos: token.NoPos,
			lhs:    []Expr{doneVar},
			rhs:    []Expr{falseLit},
		}
		cond := UnaryExpr{
			x:         doneVar,
			unaryFunc: unaryFuncs["bool43"], // hack, TODO
			op:        token.NOT,
		}
		body := BlockStmt{}

		// reset done to true at the beginning of the loop; each declaration will set it to false if it is not yet resolved
		trueLit := newBasicLit(token.NoPos, reflectTrue)
		resetDone := AssignStmt{
			tok:    token.ASSIGN,
			tokPos: token.NoPos,
			lhs:    []Expr{doneVar},
			rhs:    []Expr{trueLit},
		}
		body.list = append(body.list, resetDone)

		for _, decl := range p.env.declarations {
			body.list = append(body.list, decl)
			// each decl will push the result (true,false)
			updateDone := AssignStmt{
				tok:    token.ASSIGN,
				tokPos: token.NoPos,
				lhs:    []Expr{doneVar},
				rhs:    []Expr{noExpr{}}, // boolean result is already on stack
			}
			body.list = append(body.list, updateDone)
		}
		forStmt := ForStmt{
			init: initDone,
			cond: cond,
			body: &body,
		}
		loop := forStmt.flowWithOptions(g, true) // do not create a new environment for the loop
		if head == nil {
			head = loop
		}
		// remove done variable after resolving declarations
		g.nextStep(newFuncStep(token.NoPos, "unset done", func(vm *VM) {
			vm.currentEnv().valueUnset(doneVarName)
		}))
	}
	for i, funcDecl := range p.env.inits {
		s := newFuncStep(funcDecl.pos(), fmt.Sprintf("call %s.init.%d", p.Name, i), func(vm *VM) {
			CallExpr{}.handleFuncDecl(vm, funcDecl)
		})
		if head == nil {
			head = s
		}
		g.nextStep(s)
	}
	popFrameStep := newFuncStep(token.NoPos, "pop frame", func(vm *VM) { vm.popFrame() })
	if head == nil {
		head = popFrameStep
	}
	g.nextStep(popFrameStep)
	return
}

func (p *Package) moveMethodsToInterpretedTypes() error {
	for _, decl := range p.env.methods {
		recvType := decl.recv.List[0].typ
		var typeName string
		switch rt := recvType.(type) {
		case StarExpr:
			if ident, ok := rt.x.(Ident); ok {
				typeName = ident.name
			}
		case Ident:
			typeName = rt.name
		default:
			return fmt.Errorf("unsupported receiver type in method declaration: %T", recvType)
		}
		methodHolder := p.env.valueLookUp(typeName).Interface()
		switch holder := methodHolder.(type) {
		case StructType:
			holder.addMethod(decl)
		case ExtendedType:
			holder.addMethod(decl)
		default:
			return fmt.Errorf("unknown type holding methods: %T", holder)
		}
	}
	clear(p.env.methods)
	return nil
}

func (p *Package) String() string {
	if p == nil || p.Package == nil {
		return "Package(<nil>)"
	}
	return fmt.Sprintf("Package(%s,%s)", p.Name, p.PkgPath)
}

func (p *Package) writeCallGraph(fileName string) {
	g := dot.NewGraph(dot.Directed)
	g.NodeInitializer(func(n dot.Node) {
		n.Box()
		n.Attr("fillcolor", "#EBFAFF") // https://htmlcolorcodes.com/
		n.Attr("style", "rounded,filled")
	})
	// setup flow
	if p.callGraph != nil {
		sub := g.Subgraph("pkg."+p.Name, dot.ClusterOption{})
		p.callGraph.traverse(sub, p.Fset)
	} else {
		fmt.Println("WARN: no call graph to write for package", p.Name)
	}

	// for each function in the package create a subgraph
	values := p.env.Env.(*Environment).valueTable
	for k, v := range values {
		if funDecl, ok := v.Interface().(*FuncDecl); ok {
			if funDecl.graph == nil {
				continue
			}
			sub := g.Subgraph(k, dot.ClusterOption{})
			funDecl.graph.traverse(sub, p.Fset)
		}
	}
	os.WriteFile(fileName, []byte(g.String()), 0644)
}

var _ CanSelect = (*SDKPackage)(nil)

type SDKPackage struct {
	name        string
	pkgPath     string
	symbolTable map[string]reflect.Value // const,var,func, not types
	typesTable  map[string]reflect.Value // not reflect.Type to make Select work uniformly
}

func (p SDKPackage) selectByName(name string) reflect.Value {
	v, ok := p.symbolTable[name]
	if !ok {
		t, ok := p.typesTable[name]
		if ok {
			return t
		}
		panic(fmt.Sprintf("package not found: %s %s", p.pkgPath, name))
	}
	return v
}
func (p SDKPackage) String() string {
	return fmt.Sprintf("SDKPackage(%s,%s)", p.name, p.pkgPath)
}

type ExternalPackage struct {
	SDKPackage
}

func (p ExternalPackage) String() string {
	return fmt.Sprintf("ExternalPackage(%s,%s)", p.name, p.pkgPath)
}

var _ CanSelect = (*DotPackages)(nil)

// DotPackages represents the collection of packages that are dot-imported into the current package.
// It is used to resolve names that are not found in the current package's environment, but may be found in one of the dot-imported packages.
type DotPackages struct {
	packages []CanSelect
}

func (d DotPackages) selectByName(name string) reflect.Value {
	for _, pkg := range d.packages {
		v := pkg.selectByName(name)
		if v.IsValid() {
			return v
		}
	}
	return reflectUndeclared
}
