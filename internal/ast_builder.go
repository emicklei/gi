package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

var _ ast.Visitor = (*ASTBuilder)(nil)

// TODO used?
type buildOptions struct {
}

type funcDeclPair struct {
	decl    FuncDecl      // the mirror node
	astDecl *ast.FuncDecl // need this to find index of each block stmt
}

type ASTBuilder struct {
	stack     []Evaluable
	env       Env
	opts      buildOptions
	goPkg     *packages.Package
	funcStack stack[funcDeclPair]
	buildErr  error // capture any error during building
}

func newASTBuilder(goPkg *packages.Package) ASTBuilder {
	builtins := newBuiltinsEnvironment(nil)
	pkgenv := newPkgEnvironment(builtins)
	return ASTBuilder{goPkg: goPkg, env: pkgenv, opts: buildOptions{}}
}

func (b *ASTBuilder) Err() error { return b.buildErr }

func (b *ASTBuilder) pushEnv() {
	b.env = b.env.newChild()
}

func (b *ASTBuilder) popEnv() {
	b.env = b.env.getParent()
}

func (b *ASTBuilder) push(s Evaluable) {
	b.stack = append(b.stack, s)
}

func (b *ASTBuilder) pop() Evaluable {
	if len(b.stack) == 0 {
		panic("builder.stack is empty")
	}
	top := b.stack[len(b.stack)-1]
	b.stack = b.stack[0 : len(b.stack)-1]
	return top
}

func (b *ASTBuilder) envSet(name string, value reflect.Value) {
	b.env.set(name, value)
}

func (b *ASTBuilder) pushFuncDecl(f FuncDecl, a *ast.FuncDecl) {
	b.funcStack.push(funcDeclPair{decl: f, astDecl: a})
}
func (b *ASTBuilder) popFuncDecl() {
	b.funcStack.pop()
}

// Visit implements the ast.Visitor interface
func (b *ASTBuilder) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {

	case *ast.TypeAssertExpr:
		s := TypeAssertExpr{Lparen: n.Lparen}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		if n.Type != nil {
			b.Visit(n.Type)
			e := b.pop()
			s.Type = e.(Expr)
		}
		b.push(s)

	case *ast.ParenExpr:
		s := ParenExpr{LParen: n.Lparen}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.push(s)

	case *ast.SendStmt:
		s := SendStmt{Arrow: n.Arrow}
		b.Visit(n.Chan)
		e := b.pop()
		s.Chan = e.(Expr)
		b.Visit(n.Value)
		e = b.pop()
		s.Value = e.(Expr)
		b.push(s)

	case *ast.ChanType:
		s := ChanType{Begin: n.Begin, Arrow: n.Arrow, Dir: n.Dir}
		b.Visit(n.Value)
		e := b.pop()
		s.Value = e.(Expr)
		b.push(s)

	case *ast.SliceExpr:
		s := SliceExpr{Lbrack: n.Lbrack, Slice3: n.Slice3}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		if n.Low != nil {
			b.Visit(n.Low)
			e = b.pop()
			s.Low = e.(Expr)
		}
		if n.High != nil {
			b.Visit(n.High)
			e = b.pop()
			s.High = e.(Expr)
		}
		if n.Max != nil {
			b.Visit(n.Max)
			e = b.pop()
			s.Max = e.(Expr)
		}
		b.push(s)

	case *ast.TypeSwitchStmt:
		s := TypeSwitchStmt{SwitchPos: n.Switch}
		if n.Init != nil {
			b.Visit(n.Init)
			e := b.pop()
			s.Init = e.(Stmt)
		}
		if n.Assign != nil {
			b.Visit(n.Assign)
			e := b.pop()
			s.Assign = e.(Stmt)
		}
		if n.Body != nil {
			b.Visit(n.Body)
			blk := b.pop().(BlockStmt)
			s.Body = &blk
		}
		b.push(s)
	case *ast.Ellipsis:
		s := Ellipsis{Ellipsis: n}
		if n.Elt != nil {
			b.Visit(n.Elt)
			e := b.pop()
			s.Elt = e.(Expr)
		}
		b.push(s)
	case *ast.DeferStmt:
		s := DeferStmt{DeferPos: n.Defer}
		if n.Call != nil {
			b.Visit(n.Call)
			e := b.pop()
			s.Call = e.(Expr)
			// store call graph in the DeferStmt
			g := newGraphBuilder(b.goPkg)
			s.callGraph = s.Call.Flow(g)
		}
		b.push(s)
	case *ast.FuncLit:
		b.pushEnv()
		defer b.popEnv()
		s := FuncLit{FuncLit: n}
		if n.Type != nil {
			b.Visit(n.Type)
			e := b.pop().(FuncType)
			s.Type = &e
		}
		if n.Body != nil {
			b.Visit(n.Body)
			e := b.pop().(BlockStmt)
			s.Body = &e
		}
		// store call graph in the FuncLit
		g := newGraphBuilder(b.goPkg)
		s.callGraph = s.Body.Flow(g)

		b.push(s)
	case *ast.SwitchStmt:
		s := SwitchStmt{SwitchStmt: n}
		if n.Init != nil {
			b.Visit(n.Init)
			s.Init = b.pop().(Stmt)
		}
		if n.Tag != nil {
			b.Visit(n.Tag)
			e := b.pop()
			s.Tag = e.(Expr)
		}
		if n.Body != nil {
			b.Visit(n.Body)
			blk := b.pop().(BlockStmt)
			s.Body = blk
		}
		b.push(s)
	case *ast.CaseClause:
		s := CaseClause{CaseClause: n}
		for _, expr := range n.List {
			b.Visit(expr)
			e := b.pop()
			s.List = append(s.List, e.(Expr))
		}
		for _, stmt := range n.Body {
			b.Visit(stmt)
			e := b.pop()
			s.Body = append(s.Body, e.(Stmt))
		}
		b.push(s)
	case *ast.MapType:
		s := MapType{MapPos: n.Map}
		b.Visit(n.Key)
		e := b.pop()
		s.Key = e.(Expr)
		b.Visit(n.Value)
		e = b.pop()
		s.Value = e.(Expr)
		b.push(s)
	case *ast.IncDecStmt:
		s := &IncDecStmt{Tok: n.Tok, TokPos: n.TokPos}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.push(s)
	case *ast.ForStmt:
		s := ForStmt{ForStmt: n}
		if n.Init != nil {
			b.Visit(n.Init)
			e := b.pop()
			s.Init = e.(Stmt)
		}
		if n.Cond != nil {
			b.Visit(n.Cond)
			e := b.pop()
			s.Cond = e.(Expr)
		}
		if n.Post != nil {
			b.Visit(n.Post)
			e := b.pop()
			s.Post = e.(Stmt)
		}
		b.Visit(n.Body)
		e := b.pop()
		blk := e.(BlockStmt)
		s.Body = &blk
		b.push(s)
	case *ast.UnaryExpr:
		s := UnaryExpr{Op: n.Op, OpPos: n.OpPos}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.push(s)
	case *ast.ValueSpec:
		s := ValueSpec{}
		for _, each := range n.Names {
			b.Visit(each)
			e := b.pop()
			i := e.(Ident)
			s.Names = append(s.Names, &i)
		}
		if n.Type != nil {
			b.Visit(n.Type)
			e := b.pop()
			s.Type = e.(Expr)
		}
		if n.Values != nil {
			for _, val := range n.Values {
				b.Visit(val)
				e := b.pop()
				s.Values = append(s.Values, e.(Expr))
			}
		}
		b.push(s)

	case *ast.ExprStmt:
		s := ExprStmt{ExprStmt: n}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.push(s)
	case *ast.Ident:
		s := Ident{Name: n.Name}
		s.NamePos = n.NamePos
		b.push(s)
	case *ast.BlockStmt:
		s := BlockStmt{LbracePos: n.Lbrace}
		for _, stmt := range n.List {
			b.Visit(stmt)
			e := b.pop()
			s.List = append(s.List, e.(Stmt))
		}
		b.push(s)
	case *ast.AssignStmt:
		s := AssignStmt{Tok: n.Tok, TokPos: n.TokPos}
		for _, l := range n.Lhs {
			b.Visit(l)
			e := b.pop()
			s.Lhs = append(s.Lhs, e.(Expr))
		}
		for _, r := range n.Rhs {
			b.Visit(r)
			e := b.pop()
			s.Rhs = append(s.Rhs, e.(Expr))
		}
		b.push(s)
	case *ast.ImportSpec:
		unq, _ := strconv.Unquote(n.Path.Value)
		pkgName := path.Base(unq)
		if n.Name != nil {
			pkgName = n.Name.Name
		}
		// HACK TODO
		if strings.HasSuffix(unq, "v2") {
			if trace {
				fmt.Println("WARNING: need package name from import fix")
			}
			pkgName = path.Base(filepath.Join(unq, ".."))
		}
		// check for standard package
		if symbolTable := stdfuncs[unq]; symbolTable != nil {
			p := StandardPackage{
				Name:        pkgName,
				PkgPath:     unq,
				symbolTable: symbolTable,
			}
			// check for types
			typesTable, ok := stdtypes[unq]
			if ok {
				p.typesTable = typesTable
			}
			b.envSet(pkgName, reflect.ValueOf(p))
			break
		}
		// check for imported external package
		if symbols := importedPkgs[unq]; symbols != nil {
			p := ExternalPackage{StandardPackage: StandardPackage{
				Name:        pkgName,
				PkgPath:     unq,
				symbolTable: symbols,
			}}
			b.envSet(pkgName, reflect.ValueOf(p))
			break
		}

		// handle local file system package
		root := b.env.rootPackageEnv()
		ffpkg := root.packageTable[unq]
		if ffpkg == nil {
			// strip module prefix
			loc, err := filepath.Abs(filepath.Join("..", unq))
			if err != nil {
				b.buildErr = fmt.Errorf("failed to locate imported package %s: %v", unq, err)
				break
			}
			gopkg, err := LoadPackage(loc, nil)
			if err != nil {
				b.buildErr = fmt.Errorf("failed to load imported package %s: %v", unq, err)
				break
			}
			pkg, err := BuildPackage(gopkg)
			if err != nil {
				b.buildErr = fmt.Errorf("failed to build imported package %s: %v", unq, err)
				break
			}
			root.packageTable[unq] = pkg
			ffpkg = pkg
		}
		b.envSet(ffpkg.Name, reflect.ValueOf(ffpkg))
	case *ast.BasicLit:
		b.push(BasicLit{BasicLit: n})
	case *ast.BinaryExpr:
		// TODO experiment
		xt := b.goPkg.TypesInfo.TypeOf(n.X)
		yt := b.goPkg.TypesInfo.TypeOf(n.Y)
		binFuncKey := fmt.Sprintf("%s%d%s", xt.Underlying().String(), n.Op, yt.Underlying().String())
		binFunc, ok := binFuncs[binFuncKey]
		if ok {
			s := BinaryExpr2{}
			s.binFunc = binFunc
			s.OpPos = n.OpPos
			s.Op = n.Op
			b.Visit(n.X)
			e := b.pop()
			s.X = e.(Expr)
			b.Visit(n.Y)
			e = b.pop()
			s.Y = e.(Expr)
			b.push(s)
			return b
		}

		s := BinaryExpr{OpPos: n.OpPos, Op: n.Op}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.Visit(n.Y)
		e = b.pop()
		s.Y = e.(Expr)
		b.push(s)
	case *ast.CallExpr:
		s := CallExpr{Lparen: n.Lparen}
		b.Visit(n.Fun)
		e := b.pop()
		s.Fun = e.(Expr)
		for _, arg := range n.Args {
			b.Visit(arg)
			e := b.pop()
			s.Args = append(s.Args, e.(Expr))
		}
		b.push(s)
	case *ast.SelectorExpr:
		s := SelectorExpr{SelectorExpr: n}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.push(s)
	case *ast.StarExpr:
		s := StarExpr{StarPos: n.Star}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.push(s)
	case *ast.IfStmt:
		s := IfStmt{IfPos: n.If}
		if n.Init != nil {
			b.Visit(n.Init)
			e := b.pop()
			s.Init = e.(Stmt)
		}
		b.Visit(n.Cond)
		e := b.pop()
		s.Cond = e.(Expr)
		b.Visit(n.Body)
		e = b.pop()
		blk := e.(BlockStmt)
		s.Body = &blk
		if n.Else != nil {
			b.Visit(n.Else)
			e = b.pop()
			s.Else = e.(Stmt)
		}
		b.push(s)
	case *ast.ReturnStmt:
		s := ReturnStmt{ReturnStmt: n}
		for _, r := range n.Results {
			b.Visit(r)
			e := b.pop()
			s.Results = append(s.Results, e.(Expr))
		}
		b.push(s)
	case *ast.FuncDecl:
		// any declarations inside the function scope
		b.pushEnv()
		s := FuncDecl{
			labelToStmt: make(map[string]statementReference),
			fileSet:     b.goPkg.Fset}
		b.pushFuncDecl(s, n)
		defer b.popFuncDecl()
		if n.Recv != nil {
			b.Visit(n.Recv)
			e := b.pop()
			f := e.(FieldList)
			s.Recv = &f
		}
		b.Visit(n.Name)
		e := b.pop()
		i := e.(Ident)
		s.Name = &i

		b.Visit(n.Type)
		e = b.pop()
		f := e.(FuncType)
		s.Type = &f

		b.Visit(n.Body)
		e = b.pop()
		blk := e.(BlockStmt)
		s.Body = &blk
		b.push(s) // TODO ??

		// store call graph in the FuncDecl
		g := newGraphBuilder(b.goPkg)
		s.callGraph = s.Flow(g)

		// leave the function scope
		b.popEnv()

		if pe, ok := b.env.(*PkgEnvironment); ok {
			if n.Name.Name == "init" {
				pe.addInit(s)
			}
		}

		// register in current env
		b.envSet(n.Name.Name, reflect.ValueOf(s))

	case *ast.FuncType:
		s := FuncType{FuncPos: n.Func}
		if n.TypeParams != nil {
			b.Visit(n.TypeParams)
			e := b.pop().(FieldList)
			s.TypeParams = &e
		}
		if n.Params != nil {
			b.Visit(n.Params)
			e := b.pop()
			f := e.(FieldList)
			s.Params = &f
		}
		if n.Results != nil {
			b.Visit(n.Results)
			e := b.pop()
			f := e.(FieldList)
			s.Results = &f
		}
		b.push(s)
	case *ast.FieldList:
		s := FieldList{FieldList: n}
		for _, field := range n.List {
			b.Visit(field)
			e := b.pop()
			f := e.(Field)
			s.List = append(s.List, &f)
		}
		b.push(s)
	case *ast.Field:
		s := Field{Field: n}
		for _, name := range n.Names {
			b.Visit(name)
			e := b.pop()
			i := e.(Ident)
			s.Names = append(s.Names, &i)
		}
		b.Visit(n.Type)
		e := b.pop()
		s.Type = e.(Expr)
		// TODO tag, comment
		b.push(s)
	case *ast.GenDecl:
		// IMPORT, CONST, TYPE, or VAR
		switch n.Tok {
		case token.CONST, token.VAR:
			for _, each := range n.Specs {
				b.Visit(each)
				// let the environment know
				e := b.pop()
				c := e.(ValueSpec)
				g := newGraphBuilder(b.goPkg)
				c.callGraph = c.Flow(g)
				b.env.addConstOrVar(c)
				// add to stack as normal
				b.push(c)
			}
		case token.IMPORT:
			for _, each := range n.Specs {
				b.Visit(each)
			}
		case token.TYPE:
			for _, each := range n.Specs {
				b.Visit(each)
			}
		}
	case *ast.DeclStmt:
		s := DeclStmt{DeclStmt: n}
		b.Visit(n.Decl)
		e := b.pop()
		s.Decl = e.(Decl)
		b.push(s)
	case *ast.CompositeLit:
		s := CompositeLit{Lbrace: n.Lbrace}
		if n.Type != nil {
			b.Visit(n.Type)
			e := b.pop()
			s.Type = e.(Expr)
		}
		if n.Elts != nil {
			for _, elt := range n.Elts {
				b.Visit(elt)
				e := b.pop()
				s.Elts = append(s.Elts, e.(Expr))
			}
		}
		b.push(s)
	case *ast.ArrayType:
		s := ArrayType{Lbrack: n.Lbrack}
		if n.Len != nil {
			b.Visit(n.Len)
			e := b.pop()
			s.Len = e.(Expr)
		}
		b.Visit(n.Elt)
		e := b.pop()
		s.Elt = e.(Expr)
		b.push(s)
	case *ast.KeyValueExpr:
		s := KeyValueExpr{Colon: n.Colon}
		b.Visit(n.Key)
		e := b.pop()
		s.Key = e.(Expr)
		b.Visit(n.Value)
		e = b.pop()
		s.Value = e.(Expr)
		b.push(s)
	case *ast.TypeSpec:
		s := TypeSpec{AssignPos: n.Assign}
		if n.Name != nil {
			b.Visit(n.Name)
			e := b.pop().(Ident)
			s.Name = &e
		}
		if n.TypeParams != nil {
			b.Visit(n.TypeParams)
			e := b.pop().(FieldList)
			s.TypeParams = &e
		}
		b.Visit(n.Type)
		e := b.pop().(Expr)
		s.Type = e
		b.push(s)
		if s.Name != nil {
			b.envSet(s.Name.Name, reflect.ValueOf(s))
		} else {
			// what if nil?
		}
	case *ast.StructType:
		s := makeStructType(n)
		if n.Fields != nil {
			b.Visit(n.Fields)
			e := b.pop().(FieldList)
			s.Fields = &e
		}
		b.push(s)
	case *ast.RangeStmt:
		s := RangeStmt{ForPos: n.For}
		s.XType = b.goPkg.TypesInfo.TypeOf(n.X)
		if n.Key != nil {
			b.Visit(n.Key)
			e := b.pop()
			s.Key = e.(Expr)
		}
		if n.Value != nil {
			b.Visit(n.Value)
			e := b.pop()
			s.Value = e.(Expr)
		}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.Visit(n.Body)
		bs := b.pop().(BlockStmt)
		s.Body = &bs
		b.push(s)
	case *ast.IndexExpr:
		s := IndexExpr{Lbrack: n.Lbrack}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.Visit(n.Index)
		e = b.pop()
		s.Index = e.(Expr)
		b.push(s)
	case *ast.LabeledStmt:
		s := LabeledStmt{LabeledStmt: n}
		if n.Label != nil {
			b.Visit(n.Label)
			e := b.pop().(Ident)
			s.Label = &e
		}
		b.Visit(n.Stmt)
		e := b.pop()
		s.Stmt = e.(Stmt)
		b.push(s)

		// TODO
		// here we are created a step that actually should happen
		// when building the flow. So perhaps we need to store the statementReference
		// elsewhere and not set a labeledStep now.

		// add label -> statement by index mapping in current function
		index := slices.Index(b.funcStack.top().astDecl.Body.List, ast.Stmt(n))
		refStep := new(labeledStep)
		refStep.label = s.Label.Name
		refStep.pos = s.Pos()
		ref := statementReference{index: index, step: refStep} // has no ID
		b.funcStack.top().decl.labelToStmt[s.Label.Name] = ref

	case *ast.BranchStmt:
		s := BranchStmt{BranchStmt: n}
		if n.Label != nil {
			b.Visit(n.Label)
			e := b.pop().(Ident)
			s.Label = &e
		}
		b.push(s)
	case nil:
		// end of a branch
	default:
		fmt.Fprintf(os.Stderr, "unvisited %T\n", n)
	}
	return b
}
