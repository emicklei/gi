package internal

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestBasicLit_Eval(t *testing.T) {
	cases := []struct {
		lit      *ast.BasicLit
		expected any
	}{
		{
			lit:      &ast.BasicLit{Kind: token.INT, Value: "42"},
			expected: 42,
		},
		{
			lit:      &ast.BasicLit{Kind: token.FLOAT, Value: "3.14"},
			expected: 3.14,
		},
		{
			lit:      &ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
			expected: "hello",
		},
		{
			lit:      &ast.BasicLit{Kind: token.CHAR, Value: "'a'"},
			expected: "'a'",
		},
	}
	for _, tt := range cases {
		t.Run(tt.lit.Kind.String(), func(t *testing.T) {
			vm := newVM(newEnvironment(nil))
			bl := BasicLit{BasicLit: tt.lit}
			result := vm.returnsEval(bl)
			if result.Interface() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result.Interface())
			}
		})
	}
}

func TestCompositeArrayLit_Eval(t *testing.T) {
	t.Run("array literal", func(t *testing.T) {
		// mock array type
		at := ArrayType{
			ArrayType: &ast.ArrayType{
				Len: &ast.BasicLit{Kind: token.INT, Value: "2"},
			},
			Len: BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "2"}},
			// a real element type would be needed for a full test
			Elt: Ident{Ident: &ast.Ident{Name: "int"}},
		}
		cl := CompositeLit{
			CompositeLit: &ast.CompositeLit{},
			Type:         at,
			Elts: []Expr{
				BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "1"}},
				BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "2"}},
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
			ArrayType: &ast.ArrayType{
				Len: nil,
			},
			// a real element type would be needed for a full test
			Elt: &Ident{Ident: &ast.Ident{Name: "int"}},
		}
		cl := CompositeLit{
			CompositeLit: &ast.CompositeLit{},
			Type:         at,
			Elts: []Expr{
				BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "1"}},
				BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "2"}},
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
