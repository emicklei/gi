package pkg

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"reflect"
	"time"

	"github.com/emicklei/dot"
	"github.com/spewerspew/spew"
	"golang.org/x/tools/go/packages"
)

type Package struct {
	*packages.Package // TODO look for actual data used here
	env               *PkgEnvironment
	initialized       bool
	callGraph         Step // the setup flow: declarations and calling inits
}

func (p *Package) selectFieldOrMethod(name string) reflect.Value {
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
	// resolve := newFuncStep(token.NoPos, fmt.Sprintf("%s.declare", p.Name), func(vm *VM) {
	// 	p.resolveDeclarations(vm)
	// })
	// if head == nil {
	// 	head = resolve
	// }

	// use for statement to eval all declarations until all are declared, then move to inits
	//
	// done := false
	// for !done {
	// 	done = true
	//  <declare all and update done>
	// }
	doneVar := Ident{name: internalVarName("done", g.idgen)}
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

	// reset done to true at the beginning of the loop; each declaration will set it to false if it is not yet declared
	trueLit := newBasicLit(token.NoPos, reflectTrue)
	resetDone := AssignStmt{
		tok:    token.ASSIGN,
		tokPos: token.NoPos,
		lhs:    []Expr{doneVar},
		rhs:    []Expr{trueLit},
	}
	body.list = append(body.list, resetDone)

	for _, decl := range p.env.declarations2 {
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

	//g.nextStep(resolve)
	for i, funcDecl := range p.env.inits {
		s := newFuncStep(funcDecl.pos(), fmt.Sprintf("call %s.init.%d", p.Name, i), func(vm *VM) {
			CallExpr{}.handleFuncDecl(vm, funcDecl)
		})
		g.nextStep(s)
	}
	g.nextStep(newFuncStep(token.NoPos, "pop frame", func(vm *VM) { vm.popFrame() }))
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

//
// for debugging
//

func (p *Package) writeAST(fileName string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SPEW: failed to write AST file", r)
		}
	}()
	buf := new(bytes.Buffer)
	spew.Config.DisableMethods = true
	spew.Config.MaxDepth = 8 // TODO see if this is enough
	done := make(chan bool)
	go func() {
		// only dump the actual values of each var/function in the environment
		for _, v := range p.env.Env.(*Environment).valueTable {
			// skip SDKPackage
			val := v.Interface()
			if _, ok := val.(SDKPackage); ok {
				continue
			}
			spew.Fdump(buf, val)
		}
		done <- true
	}()
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("AST writing took more than 2 seconds, aborting")
		close(done)
	case <-done:
	}
	os.WriteFile(fileName, buf.Bytes(), 0644)
}

func writeGoAST(fileName string, goPkg *packages.Package) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("SPEW: failed to write Go AST file", r)
		}
	}()
	buf := new(bytes.Buffer)
	// do not write fileset
	fs := goPkg.Fset
	goPkg.Fset = nil
	defer func() {
		goPkg.Fset = fs
	}()
	spew.Config.DisableMethods = true
	spew.Config.MaxDepth = 8 // TODO see if this is enough
	done := make(chan bool)
	go func() {
		spew.Fdump(buf, goPkg)
		done <- true
	}()
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("Go AST writing took more than 2 seconds, aborting")
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

func (p *Package) String() string {
	if p == nil || p.Package == nil {
		return "Package(<nil>)"
	}
	return fmt.Sprintf("Package(%s,%s)", p.Name, p.PkgPath)
}

type SDKPackage struct {
	name        string
	pkgPath     string
	symbolTable map[string]reflect.Value // const,var,func, not types
	typesTable  map[string]reflect.Value // not reflect.Type to make Select work uniformly
}

func (p SDKPackage) selectFieldOrMethod(name string) reflect.Value {
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
