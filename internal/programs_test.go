package internal

import (
	"fmt"
	"strings"
	"testing"
)

func TestProgramTypeConvert(t *testing.T) {
	tests := []struct {
		typeName string
	}{
		{"int8"},
		{"int16"},
		{"int32"},
		{"int64"},
		{"int"},
	}
	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`package main
			func main() {
				a := %s(1) + 2
				print(a)
			}`, tt.typeName)
			out := parseAndWalk(t, src)
			if got, want := out, "3"; got != want {
				t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		})
	}
}

func TestProgramTypeUnsignedConvert(t *testing.T) {
	tests := []struct {
		typeName string
	}{
		{"uint8"},
		{"uint16"},
		{"uint32"},
		{"uint64"},
		{"uint"},
	}
	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`
			package main

			func main() {
				a := %s(1) + %s(2)
				print(a)
			}`, tt.typeName, tt.typeName)
			out := parseAndWalk(t, src)
			if got, want := out, "3"; got != want {
				t.Errorf("[step] got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		})
	}
}

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
}`, "'e'")
}

func TestEarlyReturn(t *testing.T) {
	testMain(t, `package main

func main() {
	if true {
		print("2")
		return
	}
	print("0")
	return
}`, "2")
}

func TestFor(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := 0; i < 10; i++ {
		print(i)
	}
	for i := 9; i > 0; i-- {
		print(i)
	}
}`, "0123456789987654321")
}

func TestForScope(t *testing.T) {
	testMain(t, `package main

func main() {
	j := 1
	for i := 0; i < 3; i++ {
		j = i
		print(i)
	}
	print(j)
}`, "0122")
}
func TestForScopeDefine(t *testing.T) {
	testMain(t, `package main

func main() {
	j := 1
	for i := 0; i < 3; i++ {
		j := i
		print(j)
	}
	print(j)
}`, "0121")
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

func TestMapClear(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"A":1, "B":2}
	clear(m)
	print(len(m))
}`, "0")
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

func TestJSONMarshal(t *testing.T) {
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

func TestGoto(t *testing.T) {
	testMain(t, `package main

func main() {
	s := 1
one:
	print(s)
	s++
	if s == 4 {
		return
	} else {
		goto two
	}
	print("unreachable")
two:
	print(s)
	s++
	goto one
}
`, "123")
}

func TestTwoPrints(t *testing.T) {
	testMain(t, `package main

func main() {
	print("one")
	print("two")
}`, "onetwo")
}

func TestDefer(t *testing.T) {
	testMain(t, `package main

func main() {
	a := 1
	defer print(a)
	a++
	defer print(a)
}`, "21")
}

func TestDeferFuncLiteral(t *testing.T) {
	testMain(t, `package main

func main() {
	defer print(0)
	f := func() {
		defer print(1)
	}
	defer print(2)
	f()
	defer print(3)
}`, "1320")
}

// https://go.dev/ref/spec#Defer_statements
func TestDeferReturnUpdateTestNestedLoop(t *testing.T) {
	// TODO
	// currently the returns puts all values on the top stackframe
	// but the defer can change the value from the environment
	// so we need to adjust the return value accordingly somehow
	t.Skip()
	testMain(t, `package main

func f() (result int) {
	defer func() {
		// result is accessed after it was set to 6 by the return statement
		result *= 7
	}()
	return 6
}
func main(){
	print(f())
}`, "42")
}

func TestDeferInLoop(t *testing.T) {
	// i must be captured by value in the defer
	testMain(t, `package main	

func main(){
	for i := 0; i <= 3; i++ {
		defer print(i)
	}
}`, "3210")
}

func TestDeferInLoopInLiteral(t *testing.T) {
	testMain(t, `package main

func main(){
	f := func() {
		for i := 0; i <= 3; i++ {
			defer print(i)
		}
	}
	f()
}`, "3210")
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

func TestMinMax(t *testing.T) {
	testMain(t, `package main

func main() {
	print(min(1,2), max(1,2))
}`, "12")
}
func TestMaxAtLeast(t *testing.T) {
	testMain(t, `package main

func main() {
	print(max(1,2,10))
	print(max(1,5,3))
}`, "105")
}

func TestMinAtMost(t *testing.T) {
	testMain(t, `package main

func main() {
	print(min(3,2,1))
	print(min(4,2,3))
}`, "12")
}

func TestMaxString(t *testing.T) {
	testMain(t, `package main

func main() {
	print(max("", "foo", "bar"))
}`, "foo")
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if got, want := fmt.Sprint(r), "oops"; got != want {
				t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		}
	}()
	testMain(t, `package main

func main() {
	panic("oops")
}`, "")
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

func TestRecover(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	defer func() {
		r := recover()
		print(r)
	}()
	panic("0")
}`, "0")
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
	t.Skip()
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
}`, "<nil>")
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

func TestPointerMethodWithFunctionLiteralArgument(t *testing.T) {
	t.Skip()
	testMain(t, `package main

import "sync"
func main() {
	var wg sync.WaitGroup
	wg.Go(func(){})
	print("done")
}`, "done")
}

func TestPointerMethod(t *testing.T) {
	testMain(t, `package main

import "sync"
func main() {
	var wg sync.WaitGroup
	wg.Wait()
	print("done")
}`, "done")
}

func TestNoInitStdtype(t *testing.T) {
	testMain(t, `package main

import "html/template"
func main() {
	var h template.HTML
	print(h)
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
