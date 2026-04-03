package pkg

import (
	"go/ast"
	"go/types"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

type InferredGenericCallExpr struct {
	ImportSpec *ast.ImportSpec
	FuncName   string
	Signature  *types.Signature
	ArgTypes   []types.Type
}

func (i *InferredGenericCallExpr) packageName() string {
	var quoted string
	if i.ImportSpec.Name != nil {
		quoted = i.ImportSpec.Name.Name
	} else {
		quoted = i.ImportSpec.Path.Value
	}
	unq, _ := strconv.Unquote(quoted)
	return unq
}

/*
*
gi.RegisterFunction(

	"slices",
	"Contains[int]",
	reflect.ValueOf(func(a0 []int, a1 int) (bool) {
		return slices.Contains(a0, a1)
	}))
*/
func (i *InferredGenericCallExpr) emitSource(w io.Writer) {
	pkg := i.packageName()
	var b strings.Builder
	b.WriteString("\tgi.RegisterFunction(\n\t\t")
	b.WriteString(strconv.Quote(pkg))
	b.WriteString(",\n\t\t\"")
	b.WriteString(i.FuncName)
	b.WriteString("(")
	// arg types
	for a, arg := range i.ArgTypes {
		if a > 0 {
			b.WriteString(",")
		}
		b.WriteString(arg.String())
	}
	b.WriteString(")\"")
	b.WriteString(",\n\t\treflect.ValueOf(func(")
	// arg
	for a, arg := range i.ArgTypes {
		if a > 0 {
			b.WriteString(",")
		}
		b.WriteString("a")
		b.WriteString(strconv.Itoa(a))
		b.WriteString(" ")
		b.WriteString(arg.String())
	}
	// returns
	b.WriteString(") (")
	results := i.Signature.Results()
	for r := 0; r < results.Len(); r++ {
		if r > 0 {
			b.WriteString(",")
		}
		b.WriteString(results.At(r).Type().String())
	}
	b.WriteString(") {\n\t\t\treturn ")
	b.WriteString(pkg)
	b.WriteString(".")
	b.WriteString(i.FuncName)
	b.WriteString("(")
	// body
	for a := range len(i.ArgTypes) {
		if a > 0 {
			b.WriteString(",")
		}
		b.WriteString("a")
		b.WriteString(strconv.Itoa(a))
	}
	// close
	b.WriteString(")\n\t}))\n")
	io.WriteString(w, b.String())
}

var _ ast.Visitor = (*genericsDetector)(nil)

type genericsDetector struct {
	goPkg   *packages.Package
	imports map[string]*ast.ImportSpec
}

func newGenericsDetector(goPkg *packages.Package) genericsDetector {
	return genericsDetector{goPkg: goPkg, imports: map[string]*ast.ImportSpec{}}
}

// Visit implements the ast.Visitor interface
func (d genericsDetector) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return d
	}
	switch n := node.(type) {
	case *ast.CallExpr:
		if igc := d.asInferredGenericCallExpr(n); igc != nil {
			if trace {
				igc.emitSource(os.Stdout)
			}
		}
		for _, each := range n.Args {
			d.Visit(each)
		}
	case *ast.ForStmt:
		d.Visit(n.Init)
		d.Visit(n.Cond)
		d.Visit(n.Post)
		d.Visit(n.Body)
	case *ast.IfStmt:
		d.Visit(n.Cond)
		d.Visit(n.Body)
		d.Visit(n.Else)
	case *ast.CompositeLit:
		for _, each := range n.Elts {
			d.Visit(each)
		}
	case *ast.ExprStmt:
		d.Visit(n.X)
	case *ast.AssignStmt:
		for _, each := range n.Rhs {
			d.Visit(each)
		}
	case *ast.FuncDecl:
		d.Visit(n.Body)
	case *ast.BlockStmt:
		for _, each := range n.List {
			d.Visit(each)
		}
	case *ast.GenDecl:
		for _, each := range n.Specs {
			d.Visit(each)
		}
	case *ast.ReturnStmt:
		for _, each := range n.Results {
			d.Visit(each)
		}
	case *ast.ParenExpr:
		d.Visit(n.X)
	case *ast.BinaryExpr:
		d.Visit(n.X)
		d.Visit(n.Y)
	case *ast.IndexExpr:
		d.Visit(n.Index)
		d.Visit(n.X)
	case *ast.DeclStmt:
		d.Visit(n.Decl)
	case *ast.ValueSpec:
		for _, each := range n.Values {
			d.Visit(each)
		}
	case *ast.UnaryExpr:
		d.Visit(n.X)
	case *ast.DeferStmt:
		d.Visit(n.Call)
	case *ast.KeyValueExpr:
		d.Visit(n.Key)
		d.Visit(n.Value)
	case *ast.RangeStmt:
		d.Visit(n.Key)
		d.Visit(n.Value)
		d.Visit(n.X)
		d.Visit(n.Body)
	case *ast.FuncLit:
		d.Visit(n.Body)
	case *ast.StarExpr:
		d.Visit(n.X)
	case *ast.SwitchStmt:
		d.Visit(n.Init)
		d.Visit(n.Tag)
		d.Visit(n.Body)
	case *ast.SendStmt:
		d.Visit(n.Chan)
		d.Visit(n.Value)
	case *ast.IncDecStmt:
		d.Visit(n.X)
	case *ast.TypeSwitchStmt:
		d.Visit(n.Init)
		d.Visit(n.Body)
		d.Visit(n.Assign)
	case *ast.CaseClause:
		for _, each := range n.List {
			d.Visit(each)
		}
		for _, each := range n.Body {
			d.Visit(each)
		}
	case *ast.TypeAssertExpr:
		d.Visit(n.X)
	case *ast.MapType:
		d.Visit(n.Key)
		d.Visit(n.Value)
	case *ast.ArrayType:
		d.Visit(n.Len)
		d.Visit(n.Elt)
	case *ast.SliceExpr:
		d.Visit(n.X)
		d.Visit(n.Low)
		d.Visit(n.High)
		d.Visit(n.Max)
	case *ast.LabeledStmt:
		d.Visit(n.Stmt)
	case *ast.ImportSpec:
		if n.Name != nil {
			unq, _ := strconv.Unquote(n.Name.Name)
			d.imports[unq] = n
		} else {
			unq, _ := strconv.Unquote(n.Path.Value)
			d.imports[unq] = n
		}
	case *ast.BranchStmt:
	case *ast.ChanType:
	case *ast.TypeSpec:
	case *ast.BasicLit:
	case *ast.SelectorExpr:
	case *ast.Ident:
	default:
		log.Fatalf("detector unhandled:%T\n", node)
	}
	return d
}

func isGenericCall(f *ast.CallExpr) bool {
	// TODO deduplicate with asInferredGenericCallExpr
	selex, ok := f.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	// must be identifier for package
	ident, ok := selex.X.(*ast.Ident)
	if !ok {
		return false
	}
	vant, ok := importedPkgs[ident.Name]
	if !ok {
		return false
	}
	if _, ok := vant.isGeneric[selex.Sel.Name]; !ok {
		return false
	}
	return true
}

func (d genericsDetector) asInferredGenericCallExpr(f *ast.CallExpr) *InferredGenericCallExpr {
	// must be selector expression
	selex, ok := f.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	// must be identifier for package
	ident, ok := selex.X.(*ast.Ident)
	if !ok {
		return nil
	}
	// if known function then it is not generic
	if pkg, ok := stdfuncs[ident.Name]; ok {
		if _, ok := pkg[selex.Sel.Name]; ok {
			return nil
		}
	}
	ret, ok := d.goPkg.TypesInfo.Types[selex]
	if !ok {
		if trace {
			console("unexpected missing type", selex)
		}
		return nil
	}
	sig := ret.Type.(*types.Signature)
	args := []types.Type{}
	for _, arg := range f.Args {
		typ, ok := d.goPkg.TypesInfo.Types[arg]
		if !ok {
			if trace {
				console("unexpected missing type", arg)
			}
			return nil
		}
		args = append(args, typ.Type)
	}
	// if not registered then not generic ; may fail later
	im, ok := d.imports[ident.Name]
	if !ok {
		if trace {
			console("missing import?", ident.Name)
		}
		return nil
	}
	return &InferredGenericCallExpr{
		ImportSpec: im,
		FuncName:   selex.Sel.Name,
		Signature:  sig,
		ArgTypes:   args,
	}
}
