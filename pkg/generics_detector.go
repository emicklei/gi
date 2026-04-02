package pkg

import (
	"go/ast"
	"log"

	"golang.org/x/tools/go/packages"
)

var _ ast.Visitor = (*genericsDetector)(nil)

type genericsDetector struct {
	goPkg *packages.Package
}

func newGenericsDetector(goPkg *packages.Package) genericsDetector {
	return genericsDetector{goPkg: goPkg}
}

// Visit implements the ast.Visitor interface
func (d genericsDetector) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return d
	}
	switch n := node.(type) {
	case *ast.CallExpr:
		if d.isGenericSDKCall(n) {
			if trace {
				console("generics call", n.Fun)
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
	case *ast.BranchStmt:
	case *ast.ChanType:
	case *ast.TypeSpec:
	case *ast.ImportSpec:
	case *ast.BasicLit:
	case *ast.SelectorExpr:
	case *ast.Ident:
	default:
		log.Fatalf("detector unhandled:%T\n", node)
	}
	return d
}

func (d genericsDetector) isGenericInterpretedCall(f *ast.CallExpr) bool {
	return false
}

func (d genericsDetector) isGenericSDKCall(f *ast.CallExpr) bool {
	// must be selector expression
	selex, ok := f.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	// must be identifier for package
	ident, ok := selex.X.(*ast.Ident)
	if !ok {
		return false
	}
	// if known function then it is not generic
	if pkg, ok := stdfuncs[ident.Name]; ok {
		if _, ok := pkg[selex.Sel.Name]; ok {
			if trace {
				console("non-generic call", ident.Name, selex.Sel.Name)
			}
			return false
		}
	}
	return true
}
