package pkg

import (
	"math"
	"reflect"
	"strconv"
	"testing"
)

func evalExpr(expr Expr) reflect.Value {
	g := newGraphBuilder(nil)
	head := expr.flow(g)
	vm := NewVM(newEnvironment(nil))
	vm.takeAllStartingAt(head)
	result := vm.popOperand()
	return result
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

func TestNewNumberPointersEqual(t *testing.T) {
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
