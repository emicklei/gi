package pkg

import (
	"go/token"
	"math"
	"reflect"
	"strconv"
	"testing"
)

func TestInterfaceEqualsNilError(t *testing.T) {
	testMain(t, `package main

func main() {
	var e error
	if e == nil {
		print("1")
	}
}`, "1")
}

func TestNewNumberPointersEqual(t *testing.T) {
	//t.Skip()
	testMain(t, `package main

func main() {
	x := new(int64)
	*x = 40
	y := new(int64)
	*y = 40
	print(x == y)
}`, "false")
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
	if nil == e {
		print("2")
	}

}`, "2")
}

func TestNilErrorEqualsNilError(t *testing.T) {
	testMain(t, `package main

func main() {
	var e1 error
	var e2 error
	if e1 == e2 {
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
	if e1 != e2 {
		print("4")
	}
}`, "4")
}

func expectFloat32(t *testing.T, expected float32) func(string) bool {
	return func(out string) bool {
		f, err := strconv.ParseFloat(out, 32)
		if err != nil {
			t.Errorf("failed to parse output: %v", err)
			return false
		}
		if math.Abs(float64(expected)-f) > 1e-6 {
			t.Errorf("got %v want %v", f, expected)
			return false
		}
		return true
	}
}

func TestBinaryExprFloat32(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 1.1
	var y float32 = 2.2
	print(x + y)
}`, expectFloat32(t, 3.3000002))
	})
	t.Run("sub", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 5.5
	var y float32 = 3.3
	print(x - y)
}`, expectFloat32(t, 2.2))
	})
	t.Run("mul", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 2.2
	var y float32 = 3.3
	print(x * y)
}`, expectFloat32(t, 7.2600005))
	})
	t.Run("quo", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 6.6
	var y float32 = 3.3
	print(x / y)
}`, expectFloat32(t, 2))
	})
	t.Run("eql", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 6.6
	var y float32 = 6.6
	print(x == y)
}`, "true")
	})
	t.Run("neq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 6.6
	var y float32 = 3.3
	print(x != y)
}`, "true")
	})
	t.Run("lss", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 3.3
	var y float32 = 6.6
	print(x < y)
}`, "true")
	})
	t.Run("leq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 3.3
	var y float32 = 3.3
	print(x <= y)
}`, "true")
	})
	t.Run("gtr", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 6.6
	var y float32 = 3.3
	print(x > y)
}`, "true")
	})
	t.Run("geq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float32 = 6.6
	var y float32 = 6.6
	print(x >= y)
}`, "true")
	})
}

func expectFloat64(t *testing.T, expected float64) func(string) bool {
	return func(out string) bool {
		f, err := strconv.ParseFloat(out, 64)
		if err != nil {
			t.Errorf("failed to parse output: %v", err)
			return false
		}
		if math.Abs(expected-f) > 1e-9 {
			t.Errorf("got %v want %v", f, expected)
			return false
		}
		return true
	}
}

func TestBinaryExprFloat64(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 1.1
	var y float64 = 2.2
	print(x + y)
}`, expectFloat64(t, 3.3))
	})
	t.Run("sub", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 5.5
	var y float64 = 3.3
	print(x - y)
}`, expectFloat64(t, 2.2))
	})
	t.Run("mul", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 2.2
	var y float64 = 3.3
	print(x * y)
}`, expectFloat64(t, 7.26))
	})
	t.Run("quo", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 6.6
	var y float64 = 3.3
	print(x / y)
}`, expectFloat64(t, 2))
	})
	t.Run("eql", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 6.6
	var y float64 = 6.6
	print(x == y)
}`, "true")
	})
	t.Run("neq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 6.6
	var y float64 = 3.3
	print(x != y)
}`, "true")
	})
	t.Run("lss", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 3.3
	var y float64 = 6.6
	print(x < y)
}`, "true")
	})
	t.Run("leq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 3.3
	var y float64 = 3.3
	print(x <= y)
}`, "true")
	})
	t.Run("gtr", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 6.6
	var y float64 = 3.3
	print(x > y)
}`, "true")
	})
	t.Run("geq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x float64 = 6.6
	var y float64 = 6.6
	print(x >= y)
}`, "true")
	})
}

func TestBinaryExprString(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x string = "hello"
	var y string = " world"
	print(x + y)
}`, "hello world")
	})
	t.Run("eql", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x string = "hello"
	var y string = "hello"
	print(x == y)
}`, "true")
	})
	t.Run("neq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x string = "hello"
	var y string = "world"
	print(x != y)
}`, "true")
	})
	t.Run("lss", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x string = "a"
	var y string = "b"
	print(x < y)
}`, "true")
	})
	t.Run("leq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x string = "a"
	var y string = "a"
	print(x <= y)
}`, "true")
	})
	t.Run("gtr", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x string = "b"
	var y string = "a"
	print(x > y)
}`, "true")
	})
	t.Run("geq", func(t *testing.T) {
		testMain(t, `package main
func main() {
	var x string = "b"
	var y string = "b"
	print(x >= y)
}`, "true")
	})
}

func TestBinaryFuncsSignedIntegers(t *testing.T) {
	t.Parallel()
	const left, right = 42, 5
	runSignedBinFuncTests(t, "int", int(left), int(right))
	runSignedBinFuncTests(t, "int8", int8(left), int8(right))
	runSignedBinFuncTests(t, "int16", int16(left), int16(right))
	runSignedBinFuncTests(t, "int32", int32(left), int32(right))
	runSignedBinFuncTests(t, "int64", int64(left), int64(right))
}

func TestBinaryFuncsUnsignedIntegers(t *testing.T) {
	t.Parallel()
	const left, right = 42, 5
	runUnsignedBinFuncTests(t, "uint", uint(left), uint(right))
	runUnsignedBinFuncTests(t, "uint8", uint8(left), uint8(right))
	runUnsignedBinFuncTests(t, "uint16", uint16(left), uint16(right))
	runUnsignedBinFuncTests(t, "uint32", uint32(left), uint32(right))
	runUnsignedBinFuncTests(t, "uint64", uint64(left), uint64(right))
}

func TestBinaryFuncsFloatTypes(t *testing.T) {
	t.Parallel()
	runFloatBinFuncTests(t, "float32", float32(6.5), float32(2))
	runFloatBinFuncTests(t, "float64", 6.5, 2.0)
}

func TestBinaryFuncsStringType(t *testing.T) {
	t.Parallel()
	runStringBinFuncTests(t, "string", "go", "gi")
}

var integerTokens = []token.Token{
	token.ADD,
	token.SUB,
	token.MUL,
	token.QUO,
	token.REM,
	token.AND,
	token.OR,
	token.XOR,
	token.SHL,
	token.SHR,
	token.AND_NOT,
	token.EQL,
	token.LSS,
	token.GTR,
	token.NEQ,
	token.LEQ,
	token.GEQ,
}

var floatTokens = []token.Token{
	token.ADD,
	token.SUB,
	token.MUL,
	token.QUO,
	token.EQL,
	token.LSS,
	token.GTR,
	token.NEQ,
	token.LEQ,
	token.GEQ,
}

var stringTokens = []token.Token{
	token.ADD,
	token.EQL,
	token.LSS,
	token.GTR,
	token.NEQ,
	token.LEQ,
	token.GEQ,
}

type signedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsignedInteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type floatNumber interface {
	~float32 | ~float64
}

func runSignedBinFuncTests[T signedInteger](t *testing.T, typeName string, left, right T) {
	t.Helper()
	t.Run(typeName, func(t *testing.T) {
		leftTyped, rightTyped := left, right
		for _, tok := range integerTokens {
			want := signedOp(tok, leftTyped, rightTyped)
			assertBinaryFunc(t, typeName, tok, leftTyped, rightTyped, want)
		}
	})
}

func runUnsignedBinFuncTests[T unsignedInteger](t *testing.T, typeName string, left, right T) {
	t.Helper()
	t.Run(typeName, func(t *testing.T) {
		leftTyped, rightTyped := left, right
		for _, tok := range integerTokens {
			want := unsignedOp(tok, leftTyped, rightTyped)
			assertBinaryFunc(t, typeName, tok, leftTyped, rightTyped, want)
		}
	})
}

func runFloatBinFuncTests[T floatNumber](t *testing.T, typeName string, left, right T) {
	t.Helper()
	t.Run(typeName, func(t *testing.T) {
		leftTyped, rightTyped := left, right
		for _, tok := range floatTokens {
			want := floatOp(tok, leftTyped, rightTyped)
			assertBinaryFunc(t, typeName, tok, leftTyped, rightTyped, want)
		}
	})
}

func runStringBinFuncTests(t *testing.T, typeName, left, right string) {
	t.Helper()
	t.Run(typeName, func(t *testing.T) {
		for _, tok := range stringTokens {
			want := stringOp(tok, left, right)
			assertBinaryFunc(t, typeName, tok, left, right, want)
		}
	})
}

func signedOp[T signedInteger](tok token.Token, left, right T) any {
	li := int64(left)
	ri := int64(right)
	switch tok {
	case token.ADD:
		return T(li + ri)
	case token.SUB:
		return T(li - ri)
	case token.MUL:
		return T(li * ri)
	case token.QUO:
		return T(li / ri)
	case token.REM:
		return T(li % ri)
	case token.AND:
		return T(li & ri)
	case token.OR:
		return T(li | ri)
	case token.XOR:
		return T(li ^ ri)
	case token.SHL:
		return T(li << uint64(ri))
	case token.SHR:
		return T(li >> uint64(ri))
	case token.AND_NOT:
		return T(li &^ ri)
	case token.EQL:
		return li == ri
	case token.LSS:
		return li < ri
	case token.GTR:
		return li > ri
	case token.NEQ:
		return li != ri
	case token.LEQ:
		return li <= ri
	case token.GEQ:
		return li >= ri
	default:
		panic("unsupported token for signed integers: " + tok.String())
	}
}

func unsignedOp[T unsignedInteger](tok token.Token, left, right T) any {
	lu := uint64(left)
	ru := uint64(right)
	switch tok {
	case token.ADD:
		return T(lu + ru)
	case token.SUB:
		return T(lu - ru)
	case token.MUL:
		return T(lu * ru)
	case token.QUO:
		return T(lu / ru)
	case token.REM:
		return T(lu % ru)
	case token.AND:
		return T(lu & ru)
	case token.OR:
		return T(lu | ru)
	case token.XOR:
		return T(lu ^ ru)
	case token.SHL:
		return T(lu << ru)
	case token.SHR:
		return T(lu >> ru)
	case token.AND_NOT:
		return T(lu &^ ru)
	case token.EQL:
		return lu == ru
	case token.LSS:
		return lu < ru
	case token.GTR:
		return lu > ru
	case token.NEQ:
		return lu != ru
	case token.LEQ:
		return lu <= ru
	case token.GEQ:
		return lu >= ru
	default:
		panic("unsupported token for unsigned integers: " + tok.String())
	}
}

func floatOp[T floatNumber](tok token.Token, left, right T) any {
	switch tok {
	case token.ADD:
		return left + right
	case token.SUB:
		return left - right
	case token.MUL:
		return left * right
	case token.QUO:
		return left / right
	case token.EQL:
		return left == right
	case token.LSS:
		return left < right
	case token.GTR:
		return left > right
	case token.NEQ:
		return left != right
	case token.LEQ:
		return left <= right
	case token.GEQ:
		return left >= right
	default:
		panic("unsupported token for floats: " + tok.String())
	}
}

func stringOp(tok token.Token, left, right string) any {
	switch tok {
	case token.ADD:
		return left + right
	case token.EQL:
		return left == right
	case token.LSS:
		return left < right
	case token.GTR:
		return left > right
	case token.NEQ:
		return left != right
	case token.LEQ:
		return left <= right
	case token.GEQ:
		return left >= right
	default:
		panic("unsupported token for strings: " + tok.String())
	}
}

func assertBinaryFunc(t *testing.T, typeName string, tok token.Token, left, right any, want any) {
	t.Helper()
	key := typeName + strconv.Itoa(int(tok)) + typeName
	fn, ok := binFuncs[key]
	if !ok {
		t.Fatalf("missing binary func for %s", key)
	}
	got := fn(reflect.ValueOf(left), reflect.ValueOf(right)).Interface()
	if got != want {
		t.Fatalf("unexpected result for %s (%v %s %v): got %v want %v", key, left, tok, right, got, want)
	}
}
