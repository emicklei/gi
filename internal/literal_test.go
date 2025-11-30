package internal

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestCompositeArrayLit_Eval(t *testing.T) {
	t.Run("array literal", func(t *testing.T) {
		// mock array type
		at := ArrayType{
			Len: newBasicLit(token.NoPos, basicLitValue(&ast.BasicLit{Kind: token.INT, Value: "2"})),
			// a real element type would be needed for a full test
			Elt: Ident{Name: "int"},
		}
		cl := CompositeLit{
			Type: at,
			Elts: []Expr{
				newBasicLit(token.NoPos, basicLitValue(&ast.BasicLit{Kind: token.INT, Value: "1"})),
				newBasicLit(token.NoPos, basicLitValue(&ast.BasicLit{Kind: token.INT, Value: "2"})),
			},
		}
		result := evalExpr(cl)
		arr := result
		if arr.Kind() != reflect.Array {
			t.Fatalf("expected array, got %v", arr.Kind())
		}
		if arr.Len() != 2 {
			t.Fatalf("expected array of length 2, got %d", arr.Len())
		}
		if arr.Index(0).Int() != 1 {
			t.Errorf("expected first element to be 1, got %d", arr.Index(0).Int())
		}
		if arr.Index(1).Int() != 2 {
			t.Errorf("expected second element to be 2, got %d", arr.Index(1).Int())
		}
	})
}

func TestCompositeSliceLit_Eval(t *testing.T) {
	t.Run("slice literal", func(t *testing.T) {
		// mock array type
		at := ArrayType{
			// a real element type would be needed for a full test
			Elt: Ident{Name: "int"},
		}
		cl := CompositeLit{
			Type: at,
			Elts: []Expr{
				newBasicLit(token.NoPos, basicLitValue(&ast.BasicLit{Kind: token.INT, Value: "1"})),
				newBasicLit(token.NoPos, basicLitValue(&ast.BasicLit{Kind: token.INT, Value: "2"})),
			},
		}
		result := evalExpr(cl)
		if result.Kind() != reflect.Slice {
			t.Fatalf("expected slice, got %v", result.Kind())
		}
		if result.Len() != 2 {
			t.Fatalf("expected slice of length 2, got %d", result.Len())
		}
		if result.Index(0).Int() != 1 {
			t.Errorf("expected first element to be 1, got %d", result.Index(0).Int())
		}
		if result.Index(1).Int() != 2 {
			t.Errorf("expected second element to be 2, got %d", result.Index(1).Int())
		}
	})
}

func TestNil(t *testing.T) {
	z := reflect.ValueOf(nil)
	t.Log(z.IsValid()) // false
	// panics
	// t.Log(z.IsZero())
	// panics
	// t.Log(z.IsNil())

	var err error = nil
	e := reflect.ValueOf(err)
	t.Log(e.IsValid()) // false
	// panics
	// t.Log(e.IsZero())
	// panics
	// t.Log(e.IsNil())
}
