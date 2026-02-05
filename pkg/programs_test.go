package pkg

import (
	"fmt"
	"strings"
	"testing"
)

func TestAssignmentOperators(t *testing.T) {
	tests := []struct {
		op   string
		want string
	}{
		{"+=", "3"},
		{"-=", "-1"},
		{"*=", "2"},
		{"/=", "0"},
		{"%=", "1"},
		{"&=", "0"},
		{"|=", "3"},
		{"^=", "3"},
		{"<<=", "4"},
		{">>=", "0"},
		{"&^=", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			src := fmt.Sprintf(`
			package main

			func main() {
				a := 1
				a %s 2
				print(a)
			}`, tt.op)
			out := parseAndWalk(t, src)
			if got, want := out, tt.want; got != want {
				t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	testMain(t, `package main

func main() {
	print("gi")
	print("flow")
}`, "giflow")
}

func TestRuneOfString(t *testing.T) {
	testMain(t, `package main

func main() {
	print(len("สวัสดี"))
	print("สวัสดี"[0])
}`, "18224")
}

func TestCompareString(t *testing.T) { // TODO compare for all comparables
	testMain(t, `package main

func main() {
	print("gi" == "flow")
}`, "false")
}

func TestTrueFalse(t *testing.T) {
	testMain(t, `package main

func main() {
	print(true, false)
}`, "truefalse")
}

func TestRune(t *testing.T) {
	testMain(t, `package main

func main() {
	print('e')
}`, "101")
}

func TestGeneric(t *testing.T) {
	testMain(t, `package main

func Generic[T any](arg T) (*T, error) { return &arg, nil }
func main() {
	h, _ := Generic("hello")
	print(*h)
}`, "hello")
}
func TestDeclare(t *testing.T) {
	testMain(t, `package main

func main() {
	var s string
	print(s)
}`, "")
}

func TestTimeConstant(t *testing.T) {
	testMain(t, `package main

import "time"
func main() {
	r := time.RFC1123
	print(r)
}`, "Mon, 02 Jan 2006 15:04:05 MST")
}

func TestTimeAliasConstant(t *testing.T) {
	testMain(t, `package main

import t "time"
func main() {
	r := t.RFC1123
	print(r)
}`, "Mon, 02 Jan 2006 15:04:05 MST")
}

func TestJSONMarshalString(t *testing.T) {
	testMain(t, `package main

import "encoding/json"
func main() {
	r,_ := json.Marshal("hello")
	print(string(r))
}`, `"hello"`)
}

func TestFloats(t *testing.T) {
	testMain(t, `package main

func main() {
	f32, f64 := float32(3.14), 3.14
	print(f32," ",f64)
}`, "3.14 3.14")
}

func TestAddressOfInt(t *testing.T) {

	testMain(t, `package main

func main() {
	i := 42
	print(&i)
}`, func(out string) bool { return strings.HasPrefix(out, "0x") })
}

func TestInit(t *testing.T) {
	testMain(t, `package main

func init() {
	print("0")
}
func init() {
	print("1")
}
func main() {}`, "01")
}

func TestTwoPrints(t *testing.T) {
	testMain(t, `package main

func main() {
	print("one")
	print("two")
}`, "onetwo")
}

func TestStringToByteSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	print([]byte("go"))
}`, "[103 111]")
}

func TestByteSliceToString(t *testing.T) {
	testMain(t, `package main

func main() {
	print(string([]byte{103,111}))
}`, "go")
}

func TestMakeMap(t *testing.T) {
	testMain(t, `package main

func main() {
	c := make(map[string]int)
	print(len(c))
	c2 := make(map[string]int,10)
	print(len(c2))
}`, "00")
}

func TestImaginary(t *testing.T) {
	testMain(t, `package main
const (
	c4 = len([10]float64{imag(2i)})  // imag(2i) is a constant and no function call is issued
)
func main() {
	a := 1+2i
	print(a)
	print(real(a))
	print(imag(a))
	print(c4)
}`, "(1+2i)1210")
}

func TestSubpackage(t *testing.T) {
	testProgramIn(t, "../examples/subpkg", "yet unchecked")
}

func TestNestedLoop(t *testing.T) {
	testProgramIn(t, "../examples/nestedloop", "todo")
}

func TestNestedLoopFromSource(t *testing.T) {
	testMain(t, `package main 

import (
	"fmt"
)

func squared(n int) int {
	return n * n
}

func main() {
	n := []int{42}
	i, j := 0, 2
	for k := i; k < j; k++ {
		for s := -1; s < k+1; s++ {
			n = append(n, squared(s))
			fmt.Println(i, j, k, s)
		}
	}
	fmt.Println(n)
}
`, func(out string) bool { return true })
}

// about nil
// https://github.com/golang/go/issues/51649
func TestNilError(t *testing.T) {
	testMain(t, `package main

func main() {
	var err error = nil
	print(err)
}`, "(0x0,0x0)")
}

func TestError(t *testing.T) {
	testMain(t, `package main

import "errors"
func main() {
	var err2 error = errors.New("an error")
	err2Msg := err2.Error()
	print(err2Msg)
}`, "an error")
}

func TestPrintError(t *testing.T) {
	testMain(t, `package main

import "errors"
func main() {
	var err2 error = errors.New("an error")
	print(err2.Error())
}`, "an error")
}

func TestPrintOne(t *testing.T) {
	testMain(t, `package main

func main() {
	var i int = 1
	print(i)
}`, "1")
}

func TestPointerMethodWithFunctionLiteralArgument(t *testing.T) {
	t.Skip()
	testMain(t, `package main

import "sync"

func main() {
	var wg sync.WaitGroup
	wg.Go(func() { print("gi") })
	print("done")
}`, "donegi")
}

// Wait has pointer receiver
func TestPointerMethod(t *testing.T) {
	testMain(t, `package main

import "sync"
func main() {
	var wg sync.WaitGroup
	wg.Wait() 
	print("done")
}`, "done")
}

func TestZeroValueStdLibType(t *testing.T) {
	testMain(t, `package main

import "html/template"
func main() {
	var h template.HTML
	print(h)
}`, "")
}

func TestCounterWithInterface(t *testing.T) {
	t.Skip()
	testMain(t, `package main

type counter interface {
	inc()
}
type myCounter int
func (c *myCounter) inc() {
	*c++
}
func main() {
	var c counter = new(myCounter)
	c.inc()
	print(c)
}`, "")
}

func TestCompileTimeMapKey(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{
	"o"+"n"+"e": 1,
}
	print(m["one"])
}`, "1")
}

func TestPrintfNumber(t *testing.T) {
	testMain(t, `package main
import "fmt"
func main() {
	i := 1
	fmt.Printf("Count: %d\n", i)
}`, "")
}

// https://stackoverflow.com/questions/67601236/run-tests-programmatically-in-go
func TestTest(t *testing.T) {
	testMain(t, `package main
import "testing"

func TestSomething(t *testing.T) {
	t.Log("This is a test")
	print("test ran")
}
func main() {
	t := new(testing.T)
	TestSomething(t)
}
`, "test ran")
}

func TestIdentRoles(t *testing.T) {
	testMain(t, `package main

func main() {
	i1 := new(int)
	i2 := int(0)
	var i3 int
	var i4 any = 0
	i5 := i4.(int)
	switch i4.(type) {
	case int:
	case *int:
	}
	i6 := (int)(0)
	i7 := (*int)(nil)
	print(*i1, i2, i3, i5, i6, i7)
}
`, "00000<nil>")
}

func TestIdentOfITypeRoles(t *testing.T) {
	testMain(t, `package main

type count int

func main() {
	i1 := new(count)
	i2 := count(0)
	var i3 count
	var i4 any = 0
	i5 := i4.(count)
	switch i4.(type) {
	case count:
	case *count:
	}
	i6 := (count)(0)
	i7 := (*count)(nil)
	print(*i1, i2, i3, i5, i6, i7)
}
`, "{{0 } map[]}0{}00&[]") // TODO fix expected output
}
