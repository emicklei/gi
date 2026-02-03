package pkg

import (
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

var _ ast.Visitor = (*astBuilder)(nil)

type astBuilder struct {
	stack     []Evaluable
	env       Env
	goPkg     *packages.Package
	funcStack stack[funcDeclPair]
	buildErr  error      // capture any error during building
	constDecl *ConstDecl // current const decl for iota tracking
}

func newASTBuilder(goPkg *packages.Package) astBuilder {
	builtinEnv := newBuiltinsEnvironment(nil)
	pkgenv := newPkgEnvironment(builtinEnv)
	return astBuilder{goPkg: goPkg, env: pkgenv}
}

func (b *astBuilder) Err() error { return b.buildErr }

func (b *astBuilder) pushEnv() {
	b.env = b.env.newChild()
}

func (b *astBuilder) popEnv() {
	b.env = b.env.getParent()
}

func (b *astBuilder) push(s Evaluable) {
	if trace {
		fmt.Printf("ast.push: %s\n", stringOf(s))
	}
	b.stack = append(b.stack, s)
}

func (b *astBuilder) pop() Evaluable {
	if len(b.stack) == 0 {
		panic("builder.stack is empty")
	}
	top := b.stack[len(b.stack)-1]
	b.stack = b.stack[0 : len(b.stack)-1]
	return top
}

func (b *astBuilder) envSet(name string, value reflect.Value) {
	b.env.set(name, value)
}

func (b *astBuilder) pushFunc(fn Func, stmtList []ast.Stmt) {
	b.funcStack.push(funcDeclPair{fn: fn, bodyList: stmtList})
}
func (b *astBuilder) popFunc() {
	b.funcStack.pop()
}

// Visit implements the ast.Visitor interface
func (b *astBuilder) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {

	case *ast.TypeAssertExpr:
		s := TypeAssertExpr{lparenPos: n.Lparen}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)
		if n.Type != nil {
			b.Visit(n.Type)
			e := b.pop()
			s.typ = e.(Expr)
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
		s := ChanType{beginPos: n.Begin, dir: n.Dir}
		b.Visit(n.Value)
		e := b.pop()
		s.valueType = e.(Expr)
		b.push(s)

	case *ast.SliceExpr:
		s := SliceExpr{lbrackPos: n.Lbrack, slice3: n.Slice3}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)
		if n.Low != nil {
			b.Visit(n.Low)
			e = b.pop()
			s.low = e.(Expr)
		}
		if n.High != nil {
			b.Visit(n.High)
			e = b.pop()
			s.high = e.(Expr)
		}
		if n.Max != nil {
			b.Visit(n.Max)
			e = b.pop()
			s.max = e.(Expr)
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
			if ce, ok := s.Call.(CallExpr); ok {
				s.callGraph = ce.deferFlow(g)
			} else {
				slog.Warn("defer statement call is not a CallExpr")
				s.callGraph = s.Call.flow(g)
			}
		}
		b.push(s)
	case *ast.FuncLit:
		b.pushEnv()
		defer b.popEnv()
		// create pointer to FuncLit to allow modification later at buildtime
		s := new(FuncLit)

		b.pushFunc(s, n.Body.List)
		defer b.popFunc()

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
		g.funcStack.push(s)
		s.callGraph = s.Body.flow(g)
		g.funcStack.pop()

		b.push(s)
	case *ast.SwitchStmt:
		s := SwitchStmt{switchPos: n.Switch}
		if n.Init != nil {
			b.Visit(n.Init)
			s.init = b.pop().(Stmt)
		}
		if n.Tag != nil {
			b.Visit(n.Tag)
			e := b.pop()
			s.tag = e.(Expr)
		}
		if n.Body != nil {
			b.Visit(n.Body)
			blk := b.pop().(BlockStmt)
			s.body = blk
		}
		b.push(s)
	case *ast.CaseClause:
		s := CaseClause{CasePos: n.Case}
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
		s := &IncDecStmt{tok: n.Tok, tokPos: n.TokPos}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)
		b.push(s)
	case *ast.ForStmt:
		s := ForStmt{forPos: n.Pos()}
		if n.Init != nil {
			b.Visit(n.Init)
			e := b.pop()
			s.init = e.(Stmt)
		}
		if n.Cond != nil {
			b.Visit(n.Cond)
			e := b.pop()
			s.cond = e.(Expr)
		}
		if n.Post != nil {
			b.Visit(n.Post)
			e := b.pop()
			s.post = e.(Stmt)
		}
		b.Visit(n.Body)
		e := b.pop()
		blk := e.(BlockStmt)
		s.body = &blk
		b.push(s)
	case *ast.UnaryExpr:
		s := UnaryExpr{op: n.Op, opPos: n.OpPos}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)

		// check type and operator combination for immediate function evaluation
		xt := b.goPkg.TypesInfo.TypeOf(n.X)
		xs := xt.Underlying().String()
		unaryFuncKey := fmt.Sprintf("%s%d", xs, n.Op)
		unaryFunc, ok := unaryFuncs[unaryFuncKey]
		if !ok {
			// check for untyped
			xs = strings.TrimPrefix(xs, "untyped ")
			unaryFuncKey = fmt.Sprintf("%s%d", xs, n.Op)
			unaryFunc, ok = unaryFuncs[unaryFuncKey]
		}
		if ok {
			s.unaryFunc = unaryFunc
		} // else use Eval switch
		b.push(s)
	case *ast.ValueSpec:
		s := ValueSpec{}
		for _, each := range n.Names {
			b.Visit(each)
			e := b.pop()
			i := e.(Ident)
			s.Names = append(s.Names, i)
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
		s := ExprStmt{}
		b.Visit(n.X)
		e := b.pop()
		s.X = e.(Expr)
		b.push(s)
	case *ast.Ident:
		// special case for iota
		if n.Name == "iota" {
			// ensure there is an iotaExpr in the current ConstDecl
			ie := b.constDecl.iotaExpr
			if ie == nil {
				ie = &iotaExpr{pos: n.NamePos}
				b.constDecl.iotaExpr = ie
			}
			b.push(ie)
			break
		}
		s := Ident{name: n.Name, namePos: n.NamePos}
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
		s := AssignStmt{tok: n.Tok, tokPos: n.TokPos}
		for _, l := range n.Lhs {
			b.Visit(l)
			e := b.pop()
			s.lhs = append(s.lhs, e.(Expr))
		}
		for _, r := range n.Rhs {
			b.Visit(r)
			e := b.pop()
			s.rhs = append(s.rhs, e.(Expr))
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
			p := SDKPackage{
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
			p := ExternalPackage{SDKPackage: SDKPackage{
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
		b.push(newBasicLit(n.ValuePos, basicLitValue(n)))
	case *ast.BinaryExpr:
		s := BinaryExpr{opPos: n.OpPos, op: n.Op}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)
		b.Visit(n.Y)
		e = b.pop()
		s.y = e.(Expr)

		// check type and operator combination for immediate function evaluation
		xt, yt := b.goPkg.TypesInfo.TypeOf(n.X), b.goPkg.TypesInfo.TypeOf(n.Y)
		xs, ys := xt.Underlying().String(), yt.Underlying().String()
		binFuncKey := fmt.Sprintf("%s%d%s", xs, n.Op, ys)
		binFunc, ok := binFuncs[binFuncKey]
		if !ok {
			// check for untyped
			xs = strings.TrimPrefix(xs, "untyped ")
			ys = strings.TrimPrefix(ys, "untyped ")
			binFuncKey = fmt.Sprintf("%s%d%s", xs, n.Op, ys)
			binFunc, ok = binFuncs[binFuncKey]
		}
		if ok {
			s.binaryFunc = binFunc
		}
		b.push(s)
	case *ast.CallExpr:
		s := CallExpr{lparenPos: n.Lparen}
		b.Visit(n.Fun)
		e := b.pop()
		s.fun = e.(Expr)
		if isRecoverCall(s.fun) {
			// mark enclosing function as having a recover call
			b.funcStack.underTop().fn.setHasRecoverCall(true)
		}
		for _, arg := range n.Args {
			b.Visit(arg)
			e := b.pop()
			s.args = append(s.args, e.(Expr))
		}
		b.push(s)
	case *ast.SelectorExpr:
		s := SelectorExpr{selector: &Ident{name: n.Sel.Name, namePos: n.Sel.NamePos}}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)
		b.push(s)
	case *ast.StarExpr:
		s := StarExpr{starPos: n.Star}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)
		b.push(s)
	case *ast.IfStmt:
		s := IfStmt{ifPos: n.If}
		if n.Init != nil {
			b.Visit(n.Init)
			e := b.pop()
			s.init = e.(Stmt)
		}
		b.Visit(n.Cond)
		e := b.pop()
		s.cond = e.(Expr)
		b.Visit(n.Body)
		e = b.pop()
		blk := e.(BlockStmt)
		s.body = &blk
		if n.Else != nil {
			b.Visit(n.Else)
			e = b.pop()
			s.elseif = e.(Stmt)
		}
		b.push(s)
	case *ast.ReturnStmt:
		s := ReturnStmt{returnPos: n.Pos()}
		for _, r := range n.Results {
			b.Visit(r)
			e := b.pop()
			s.results = append(s.results, e.(Expr))
		}
		b.push(s)
	case *ast.FuncDecl:
		// any declarations inside the function scope
		b.pushEnv()
		// create pointer to FuncDecl to allow modification later at buildtime
		s := &FuncDecl{fileSet: b.goPkg.Fset}

		b.pushFunc(s, n.Body.List)
		defer b.popFunc()

		if n.Recv != nil {
			b.Visit(n.Recv)
			e := b.pop()
			f := e.(FieldList)
			s.recv = &f
		}
		b.Visit(n.Name)
		e := b.pop()
		i := e.(Ident)
		s.name = &i

		b.Visit(n.Type)
		e = b.pop()
		f := e.(FuncType)
		s.typ = &f

		b.Visit(n.Body)
		e = b.pop()
		blk := e.(BlockStmt)
		s.body = &blk
		b.push(s) // TODO ??

		// store call graph in the FuncDecl
		g := newGraphBuilder(b.goPkg)
		s.graph = s.flow(g)

		// leave the function scope
		b.popEnv()

		if pe, ok := b.env.(*PkgEnvironment); ok {
			if n.Recv != nil {
				pe.addMethod(s)
			} else {
				if n.Name.Name == "init" {
					pe.addInit(s)
				}
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
			f.ensureNamedFields("result")
			s.Results = &f
		}
		b.push(s)
	case *ast.FieldList:
		s := FieldList{OpeningPos: n.Opening}
		for _, field := range n.List {
			b.Visit(field)
			e := b.pop()
			f := e.(Field)
			s.List = append(s.List, &f)
		}
		b.push(s)
	case *ast.Field:
		s := Field{}
		for _, name := range n.Names {
			b.Visit(name)
			e := b.pop()
			i := e.(Ident)
			s.names = append(s.names, &i)
		}
		b.Visit(n.Type)
		e := b.pop()
		s.typ = e.(Expr)
		if n.Tag != nil {
			b.Visit(n.Tag)
			e := b.pop().(BasicLit)
			v := e.value.Interface().(string)
			s.tag = &v
		}
		b.push(s)
	case *ast.GenDecl:
		// IMPORT, CONST, TYPE, or VAR
		switch n.Tok {
		case token.CONST:
			if len(b.funcStack) > 0 {
				// inside function, handle iota differently
				decl := ConstDecl{}
				b.constDecl = &decl // set current const decl for iota tracking
				var lastExpr Expr
				for _, each := range n.Specs {
					b.Visit(each)
					// must be ValueSpec because CONST
					vs := b.pop().(ValueSpec)
					if len(vs.Values) == 0 {
						vs.Values = append(vs.Values, lastExpr)
					} else {
						lastExpr = vs.Values[0]
					}
					// store call graph in the ValueSpec for initialization
					g := newGraphBuilder(b.goPkg)
					vs.graph = vs.flow(g)
					decl.specs = append(decl.specs, vs)
				}
				b.push(decl)
				b.constDecl = nil // clear current const decl
				break
			}
			// set iota for package level const block
			// inside function, handle iota differently
			decl := ConstDecl{}
			b.constDecl = &decl // set current const decl for iota tracking
			var lastExpr Expr
			for _, each := range n.Specs {
				b.Visit(each)
				// must be ValueSpec because CONST
				vs := b.pop().(ValueSpec)
				if len(vs.Values) == 0 {
					vs.Values = append(vs.Values, lastExpr)
				} else {
					lastExpr = vs.Values[0]
				}
				// store call graph in the ValueSpec for initialization
				g := newGraphBuilder(b.goPkg)
				vs.graph = vs.flow(g)
				decl.specs = append(decl.specs, vs)
			}
			// store call graph in the ConstDecl for initialization
			g := newGraphBuilder(b.goPkg)
			decl.graph = decl.flow(g)
			b.constDecl = nil // clear current const decl
			b.env.addCanDeclare(decl)
		case token.VAR:
			for _, each := range n.Specs {
				b.Visit(each)
				e := b.pop()
				c := e.(ValueSpec)
				g := newGraphBuilder(b.goPkg)
				c.graph = c.flow(g)
				// let the environment know
				b.env.addCanDeclare(c)
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
		s := DeclStmt{}
		b.Visit(n.Decl)
		e := b.pop()
		s.decl = e.(Decl)
		b.push(s)
	case *ast.CompositeLit:
		s := CompositeLit{Lbrace: n.Lbrace}
		if n.Type != nil {
			b.Visit(n.Type)
			e := b.pop()
			s.Type = e.(Expr)
			s.ParserType = b.goPkg.TypesInfo.TypeOf(n.Type)
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
		s := ArrayType{lbrackPos: n.Lbrack}
		if n.Len != nil {
			b.Visit(n.Len)
			e := b.pop()
			s.len = e.(Expr)
		}
		b.Visit(n.Elt)
		e := b.pop()
		s.elt = e.(Expr)
		b.push(s)
	case *ast.KeyValueExpr:
		s := KeyValueExpr{colonPos: n.Colon}
		b.Visit(n.Key)
		e := b.pop()
		s.key = e.(Expr)
		b.Visit(n.Value)
		e = b.pop()
		s.Value = e.(Expr)
		b.push(s)
	case *ast.TypeSpec:
		s := TypeSpec{assignPos: n.Assign}
		if n.Name != nil {
			b.Visit(n.Name)
			e := b.pop().(Ident)
			s.name = &e
		}
		if n.TypeParams != nil {
			b.Visit(n.TypeParams)
			e := b.pop().(FieldList)
			s.typeParams = &e
		}
		b.Visit(n.Type)
		e := b.pop().(Expr)
		s.typ = e
		if st, ok := e.(StructType); ok {
			// set the name of the struct type
			st.name = s.name.name
			b.envSet(s.name.name, reflect.ValueOf(st))
		} else if idn, ok := e.(Ident); ok {
			ext := newExtendedType(idn)
			b.envSet(s.name.name, reflect.ValueOf(ext))
		} else if se, ok := e.(StarExpr); ok {
			// first make it work TODO
			// assume StarExpr.X of Ident for now
			ext := newExtendedType(se.x.(Ident))
			b.envSet(s.name.name, reflect.ValueOf(ext))
		} else {
			panic("unsupported type spec type")
		}
		b.push(s)
	case *ast.StructType:
		s := makeStructType(n)
		if n.Fields != nil {
			b.Visit(n.Fields)
			e := b.pop().(FieldList)
			s.fields = &e
		}
		b.push(s)
	case *ast.InterfaceType:
		s := InterfaceType{InterfacePos: n.Interface}
		if n.Methods != nil {
			b.Visit(n.Methods)
			e := b.pop().(FieldList)
			s.Methods = &e
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
		s := IndexExpr{lbrackPos: n.Lbrack}
		b.Visit(n.X)
		e := b.pop()
		s.x = e.(Expr)
		b.Visit(n.Index)
		e = b.pop()
		s.index = e.(Expr)
		b.push(s)
	case *ast.LabeledStmt:
		s := LabeledStmt{colonPos: n.Pos()}
		if n.Label != nil {
			b.Visit(n.Label)
			e := b.pop().(Ident)
			s.label = &e
		}
		b.Visit(n.Stmt)
		e := b.pop()
		s.statement = e.(Stmt)
		b.push(s)

		// TODO
		// here we are creating a step that actually should happen
		// when building the flow. So perhaps we need to store the statementReference
		// elsewhere and not set a labeledStep now.

		// add label -> statement by index mapping in current function
		index := slices.Index(b.funcStack.top().bodyList, ast.Stmt(n))
		refStep := new(labeledStep)
		refStep.label = s.label.name
		refStep.pos = s.Pos()
		ref := stmtReference{index: index, step: refStep} // has no ID
		b.funcStack.top().fn.putGotoReference(s.label.name, ref)
	case *ast.BranchStmt:
		s := BranchStmt{TokPos: n.TokPos, Tok: n.Tok}
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
