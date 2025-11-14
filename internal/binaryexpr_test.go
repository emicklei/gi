package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
	"reflect"
	"testing"
)

func evalExpr(expr Expr) reflect.Value {
	g := newGraphBuilder(nil)
	head := expr.Flow(g)
	vm := newVM(newEnvironment(nil))
	vm.takeAllStartingAt(head)
	result := vm.frameStack.top().pop()
	return result
}
func TestBinaryExprValue_Eval(t *testing.T) {
	t.Parallel()
	t.Run("int op int", func(t *testing.T) {
		cases := []struct {
			op       token.Token
			left     string
			right    string
			expected any
		}{
			{token.ADD, "1", "2", int64(3)},
			{token.SUB, "5", "3", int64(2)},
			{token.MUL, "2", "3", int64(6)},
			{token.QUO, "6", "3", int64(2)},
			{token.REM, "7", "3", int64(1)},
			{token.EQL, "3", "3", true},
			{token.EQL, "3", "4", false},
			{token.NEQ, "3", "4", true},
			{token.NEQ, "3", "3", false},
			{token.LSS, "3", "4", true},
			{token.LSS, "3", "3", false},
			{token.LEQ, "3", "3", true},
			{token.LEQ, "3", "4", true},
			{token.LEQ, "4", "3", false},
			{token.GTR, "4", "3", true},
			{token.GTR, "3", "3", false},
			{token.GEQ, "4", "4", true},
			{token.GEQ, "4", "3", true},
			{token.GEQ, "3", "4", false},
		}
		for _, tt := range cases {
			t.Run(fmt.Sprintf("int %s int", tt.op), func(t *testing.T) {
				left := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: tt.left}}
				right := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: tt.right}}
				expr := BinaryExpr{
					X:  left,
					Y:  right,
					Op: tt.op,
				}
				result := evalExpr(expr)

				switch expected := tt.expected.(type) {
				case int64:
					if result.Int() != expected {
						t.Errorf("expected %d, got %d", expected, result.Int())
					}
				case bool:
					if result.Bool() != expected {
						t.Errorf("expected %v, got %v", expected, result.Bool())
					}
				}
			})
		}
	})

	t.Run("float op float", func(t *testing.T) {
		cases := []struct {
			op       token.Token
			left     string
			right    string
			expected float64
		}{
			{token.ADD, "1.1", "2.2", 3.3},
			{token.SUB, "5.5", "3.3", 2.2},
			{token.MUL, "2.2", "3.3", 7.26},
			{token.QUO, "6.6", "3.3", 2.0},
		}
		for _, tt := range cases {
			t.Run(fmt.Sprintf("float %s float", tt.op), func(t *testing.T) {
				left := BasicLit{BasicLit: &ast.BasicLit{Kind: token.FLOAT, Value: tt.left}}
				right := BasicLit{BasicLit: &ast.BasicLit{Kind: token.FLOAT, Value: tt.right}}
				expr := BinaryExpr{
					X:  left,
					Y:  right,
					Op: tt.op,
				}
				result := evalExpr(expr)
				if math.Abs(result.Float()-tt.expected) > 1e-9 {
					t.Errorf("expected %f, got %f", tt.expected, result.Float())
				}
			})
		}
	})

	t.Run("int op float", func(t *testing.T) {
		left := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "42"}}
		right := BasicLit{BasicLit: &ast.BasicLit{Kind: token.FLOAT, Value: "3.14"}}
		expr := BinaryExpr{
			X:  left,
			Y:  right,
			Op: token.ADD,
		}
		result := evalExpr(expr)
		if result.Kind() != reflect.Float64 {
			t.Fatalf("expected float64 result, got %v", result.Kind())
		}
		if result.Float() != 45.14 {
			t.Fatalf("expected 45.14, got %f", result.Float())
		}
	})

	t.Run("float op int", func(t *testing.T) {
		left := BasicLit{BasicLit: &ast.BasicLit{Kind: token.FLOAT, Value: "3.14"}}
		right := BasicLit{BasicLit: &ast.BasicLit{Kind: token.INT, Value: "42"}}
		expr := BinaryExpr{
			X:  left,
			Y:  right,
			Op: token.ADD,
		}
		result := evalExpr(expr)
		if result.Kind() != reflect.Float64 {
			t.Fatalf("expected float64 result, got %v", result.Kind())
		}
		if result.Float() != 45.14 {
			t.Fatalf("expected 45.14, got %f", result.Float())
		}
	})

	t.Run("string op string", func(t *testing.T) {
		left := BasicLit{BasicLit: &ast.BasicLit{Kind: token.STRING, Value: "Hello, "}}
		right := BasicLit{BasicLit: &ast.BasicLit{Kind: token.STRING, Value: "World!"}}
		expr := BinaryExpr{
			X:  left,
			Y:  right,
			Op: token.ADD,
		}
		result := evalExpr(expr)
		if result.Kind() != reflect.String {
			t.Fatalf("expected string result, got %v", result.Kind())
		}
		if result.String() != "Hello, World!" {
			t.Fatalf(`expected "Hello, World!", got %s`, result.String())
		}
	})
}

func TestBinaryExprValue_IntOpInt(t *testing.T) {
	t.Parallel()
	bv := BinaryExprValue{
		op:    token.ADD,
		left:  reflect.ValueOf(int(1)),
		right: reflect.ValueOf(int(2)),
	}
	br := bv.Eval()
	if got, want := br.Kind(), reflect.Int; got != want {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
	}
	bv = BinaryExprValue{
		op:    token.ADD,
		left:  reflect.ValueOf(int64(1)),
		right: reflect.ValueOf(int64(2)),
	}
	br = bv.Eval()
	if got, want := br.Kind(), reflect.Int64; got != want {
		t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
	}
}

func TestInterfaceEqualsNilError(t *testing.T) {
	testMain(t, `package main
 
func main() {
	var e error
	if e == nil {
		print("1")
	}		
}`, "1")
}

func TestBinaryExpr2(t *testing.T) {
	testMain(t, `package main
 
func main() {
	var x int64 = 40
	var y int64 = 2
	print(x + y)	
}`, "42")
}

func TestBinaryExpr2Funcs(t *testing.T) {
	testMain(t, `package main
 
func f1() int64 {
	return 40
}
func f2() int64 {
	return 2
}	

func main() {
	print(f1() + f2())	
}`, "42")
}

func TestUntypedNilEqualsNilError(t *testing.T) {
	testMain(t, `package main
 
func main() {
	var e error
	if nil == e{
		print("2")
	}		
}`, "2")
}
func TestNilErrorEqualsNilError(t *testing.T) {
	testMain(t, `package main
 
func main() {
	var e1 error
	var e2 error
	if e1 == e2{
		print("3")
	}		
}`, "3")
}
func TestNonNilErrorEqualsNilError(t *testing.T) {
	testMain(t, `package main
import "errors"
func main() {
	e1 := errors.New("err")
	var e2 error
	if e1 != e2{
		print("4")
	}		
}`, "4")
}
