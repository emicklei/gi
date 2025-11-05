package internal

import (
	"fmt"
	"os"
	"regexp"
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

func TestCompareString(t *testing.T) { // TODO compare for all comparables
	testMain(t, `package main

func main() {
	print("gi" == "flow")
}`, "false")
}

func TestMultiAssign(t *testing.T) {
	testMain(t, `package main	
func main() {
	in1, in2 := "gi", "flow"
	print(in1, in2)
}`, "giflow")
}

func TestAssignToStructField(t *testing.T) {
	testMain(t, `package main

type Point struct {
	X int
	Y int
}
func main() {
	x := 5
	p := Point{X: 10}
	p.X = x
	p.Y = 20
	print(p.X, p.Y)
}`, "520")
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
func TestNumbers(t *testing.T) {
	testMain(t, `package main

func main() {
	print(-1,+3.14,0.1e10)
}`, "-13.141e+09")
}
func TestFunc(t *testing.T) {
	testMain(t, `package main

func plus(a int, b int) int {
	return a + b
}
func main() {
	result := plus(2, 3)
	print(result)
}`, "5")
}

func TestFuncMultiReturn(t *testing.T) {
	testMain(t, `package main

func ab(a int, b int) (int,int) {
	return a,b
}
func main() {
	a,b := ab(2, 3)
	print(a,b)
}`, "23")
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

func TestConst(t *testing.T) {
	testMain(t, `package main

const (
	C = A+1
	A = 0
	B = 1
)
func main() {
	print(A,B,C)
}`, "011")
}

func TestVar(t *testing.T) {
	testMain(t, `package main

var (
	a = 1
	s string
	b bool
)
func main() {
	print(a,s,b)
}`, "1false")
}

func TestConstScope(t *testing.T) {
	testMain(t, `package main

var b = a
func main() {
	var b = a
	const a = 2
	print(a, b)
}
const a = 1`, "21")
}

func TestDeclareAndInit(t *testing.T) {
	testMain(t, `package main

func main() {
	var s string = "gi"
	print(s)
}`, "gi")
}

func TestSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	print([]int{1, 2})
}`, "[1 2]")
}

func TestSliceLen(t *testing.T) {
	testMain(t, `package main

func main() {
	print(len([]int{1}))
}`, "1")
}

func TestSliceCap(t *testing.T) {
	testMain(t, `package main

func main() {
	print(cap([]int{1}))
}`, "4")
}

func TestArray(t *testing.T) {
	testMain(t, `package main

func main() {
	print([2]string{"A", "B"})
}`, "[A B]")
}

func TestArrayLen(t *testing.T) {
	testMain(t, `package main

func main() {
	print(len([2]string{"A", "B"}))
}`, "2")
}

func TestArrayCap(t *testing.T) {
	testMain(t, `package main

func main() {
	print(cap([2]string{"A", "B"}))
}`, "2")
}

func TestSliceClear(t *testing.T) {
	testMain(t, `package main

func main() {
	s := []int{1,2,3}
	clear(s)
	print(len(s))
}`, "3")
}

func TestMapClear(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"A":1, "B":2}
	clear(m)
	print(len(m))
}`, "0")
}

func TestSliceAppendAndIndex(t *testing.T) {
	testMain(t, `package main

func main() {
	list := []int{}
	list = append(list, 1, 2)
	print(list[0], list[1])
}`, "12")
}

func TestSubSlice(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	list := []int{1,2,3}
	print(list[1:2])
}`, "[23]")
}

func TestAppend(t *testing.T) {
	testMain(t, `package main

func main() {
	list := []int{}
	print(list)
	list = append(list, 4, 5)
	print(list)
}`, "[][4 5]")
}

func TestAppendStringToByteSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	var b []byte
	b = append(b, "bar"...)
	print(string(b))
}`, "bar")
}

// https://go.dev/ref/spec#Appending_and_copying_slices
func TestCopy(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {	
	var a = [...]int{0, 1, 2, 3, 4, 5, 6, 7}
	var s = make([]int, 6)
	var b = make([]byte, 5)
	n1 := copy(s, a[0:])            // n1 == 6, s is []int{0, 1, 2, 3, 4, 5}
	n2 := copy(s, s[2:])            // n2 == 4, s is []int{2, 3, 4, 5, 4, 5}
	n3 := copy(b, "Hello, World!")  // n3 == 5, b is []byte("Hello")
	print(n1, n2, n3)
}`, "645")
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

func TestNewType(t *testing.T) {
	testMain(t, `package main

type Airplane struct {
	Model string
}
func main() {
	heli := Airplane{Model:"helicopter"}
	print(heli.Model)
}`, "helicopter")
}

func TestAddressOfType(t *testing.T) {
	testMain(t, `package main

type Airplane struct {
	Model string
}
func main() {
	heli := &Airplane{Model:"helicopter"}
	print(heli.Model)
}`, "helicopter")
}

func TestAddressOfInt(t *testing.T) {
	testMain(t, `package main

func main() {
	i := 42
	print(&i)
}`, func(out string) bool { return strings.HasPrefix(out, "0x") })
}

func TestRangeOfStrings(t *testing.T) {
	testMain(t, `package main

func main() {
	strings := []string{"hello", "world"}
	for i,s := range strings {
		print(i,s)
	}
}`, "0hello1world")
}

func TestRangeOfStringsNoValue(t *testing.T) {
	testMain(t, `package main

func main() { 
	for i := range [2]string{} {
		print(i)
	}
}`, "01")
}

func TestRangeOfIntNoKey(t *testing.T) {
	testMain(t, `package main

func main() {
	for range 2 {
		print("a")
	}
}`, "aa")
}

func TestRangeOfIntWithKey(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := range 2 {
		print(i)
	}
}`, "01")
}

func TestRangeOfMap(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"a":1, "b":2}
	for k,v := range m {
		print(k,v)
	}
}`, func(out string) bool { return out == "a1b2" || out == "b2a1" })
}

func TestRangeNested(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"a": 1, "b": 2}
	for j := range []int{0, 1} {
		for range j {
			for i := range 2 {
				for k, v := range m {
					print(i)
					print(k)
					print(v)
				}
			}
		}
	}
}`, func(out string) bool {
		// because map iteration is random we need to match all possibilities
		ok, _ := regexp.MatchString("^(?:0a10b2|0b20a1)(?:1a11b2|1b21a1)$", out)
		return ok
	})
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

func TestMethod(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func (_ Airplane) S() string { return "airplane" } // put before type on purpose
type Airplane struct {}
func main() {
	print(Airplane{}.S())
}`, "airplane")
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

func TestMap(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{}
	m["a"] = 1
	m["b"] = 2
	print(m["a"] + m["b"])
}`, "3")
}

func TestMapOk(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{}
	m["a"] = 1
	a, ok := m["a"]
	print(a,ok)
}`, "1true")
}

func TestMapInitialized(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"a":1, "b":2}
	print(len(m))
}`, "2")
}

func TestMapDelete(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"a":1, "b":2}
	delete(m, "a")
	print(len(m))
}`, "1")
}

func TestIfElseIfElse(t *testing.T) {
	testMain(t, `package main

func main() {
	if 1 == 2 {
		print("unreachable 1")
	} else if 2 == 2 {
		print("gi")
	} else {
		print("unreachable 2")
	}
}`, "gi")
}

func TestIfIf(t *testing.T) {
	testMain(t, `package main

func main() {
	if 1 == 2 {
		print("unreachable")
	} 
	if 2 == 2 {
		print("gi")
	}
}`, "gi")
}

func TestTwoPrints(t *testing.T) {
	testMain(t, `package main

func main() {
	print("one")
	print("two")
}`, "onetwo")
}

func TestVariadicFunction(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func main() {
	print(sum(1, 2, 3))
}`, "6")
}

func TestDefer(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	defer print(1)
	defer print(2)
}`, "12")
}

func TestNamedReturn(t *testing.T) {
	testMain(t, `package main
		
func f() (result int) {
	return 1 
}
func main(){
	print(f())
}`, "1")
}

// https://go.dev/ref/spec#Defer_statements
func TestDeferReturn(t *testing.T) {
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
	t.Skip()
	// i must be captured by value in the defer
	testMain(t, `package main	

import "fmt"
func main(){
	for i := 0; i <= 3; i++ {
		defer fmt.Print(i)
	}
}`, "3210")
}

func TestDeferInLoopInLiteral(t *testing.T) {
	t.Skip()
	testMain(t, `package main

import "fmt"
func main(){
	f := func() {
		for i := 0; i <= 3; i++ {
			defer fmt.Print(i)
		}
	}
	f()
}`, "3210")
}

func TestMinMax(t *testing.T) {
	testMain(t, `package main

func main() {
	print(min(1,2), max(1,2))
}`, "12")
}

func TestTypeAlias(t *testing.T) {
	testMain(t, `package main

type MyInt = int

func main() {
	var a MyInt = 1
	print(a)
}`, "1")
}

func TestFunctionLiteral(t *testing.T) {
	testMain(t, `package main

func main() {
	f := func(a int) int { return a }
	print(f(1))
}`, "1")
}

func TestSwitchOnBool(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int = 1
	switch {
	case a == 1:
		print(a)
	}
}`, "1")
}

func TestSwitchOnLiteral(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int
	switch a = 1; a {
	case 1:
		print(a)
	}
}`, "1")
}

func TestSwitchDefault(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int
	switch a {
	case 2:
	default:
		print(3)
	}
}`, "3")
}

func TestSwitch(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int
	switch a = 1; a {
	case 1:
		print(a)
	}
	switch a {
	case 2:
	default:
		print(3)
	}
}`, "13")
}

/**
a = 1
if a == 1 {
	print(a)
	goto end
}
if a == 2 {
	print(a)
	goto end
}
print(2)
end:
**/

func TestSwitchType(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	var v any
	v = "gi"
	switch v := v.(type) {
	case int:
		print("int:", v)
	case string:
		print("string:", v)
	default:
		print("unknown:", v)
	}
}`, "string:gi")
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

func TestFuncAsPackageVar(t *testing.T) {
	testMain(t, `package main

const h = "1"
var f = func() string { return h }

func main() {
	print(f())
}`, "1")
}

func TestIfMultiAssign(t *testing.T) {
	testMain(t, `package main

func main() {
	if got, want := min(1,2), 1; got == want {
		print("min")
	}
}`, "min")
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

func TestUnaries(t *testing.T) {
	tests := []struct {
		src  string
		op   string
		want string
	}{
		{"true", "!", "false"},
		{"int(1)", "^", "-2"},
		{"int8(1)", "^", "-2"},
		{"int16(1)", "^", "-2"},
		{"int32(1)", "^", "-2"},
		{"int64(1)", "^", "-2"},
		{"uint64(1)", "^", "18446744073709551614"},
		{"uint32(1)", "^", "4294967294"},
		{"uint16(1)", "^", "65534"},
		{"uint8(1)", "^", "254"},
		{"uint(1)", "^", "18446744073709551614"},
		{"int(1)", "+", "1"},
		{"int8(1)", "+", "1"},
		{"int16(1)", "+", "1"},
		{"int32(1)", "+", "1"},
		{"int64(1)", "+", "1"},
		{"uint64(1)", "+", "1"},
		{"uint32(1)", "+", "1"},
		{"uint16(1)", "+", "1"},
		{"uint8(1)", "+", "1"},
		{"uint(1)", "+", "1"},
		{"int(1)", "+", "1"},
		{"int8(1)", "-", "-1"},
		{"int16(1)", "-", "-1"},
		{"int32(1)", "-", "-1"},
		{"int64(1)", "-", "-1"},
		{"uint64(1)", "-", "18446744073709551615"},
		{"uint32(1)", "-", "4294967295"},
		{"uint16(1)", "-", "65535"},
		{"uint8(1)", "-", "255"},
		{"uint(1)", "-", "18446744073709551615"},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`
			package main

			func main() {
				v := %s
				print(%sv)
			}`, tt.src, tt.op)
			out := parseAndWalk(t, src)
			if got, want := out, tt.want; got != want {
				t.Errorf("%s got [%[1]v:%[1]T] want [%[2]v:%[2]T]", tt.src, got, want)
			}
		})
	}
}

func TestImaginary(t *testing.T) {
	testMain(t, `package main

func main() {
	a := 1+2i
	print(a)
	print(real(a))
	print(imag(a))
}`, "(1+2i)12")
}

// https://go.dev/ref/spec#Package_initialization
func TestDeclarationExample(t *testing.T) {
	testMain(t, `package main

var (
	a = c + b  // == 9
	b = f()    // == 4
	c = f()    // == 5
	d = 3      // == 5 after initialization has finished
)

func f() int {
	d++
	return d
}
func main() {
	print(a,b,c,d)
}`, "9455")
}

func TestSubpackage(t *testing.T) {
	if os.Getenv("GI_TRACE") == "" {
		t.Skip("GI_TRACE not set")
	}
	testProgramIn(t, "../examples/subpkg", "yet unchecked")
}

func TestNestedLoop(t *testing.T) {
	t.Skip()
	if os.Getenv("GI_TRACE") == "" {
		t.Skip("GI_TRACE not set")
	}
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

func TestPointerBasic(t *testing.T) {
	testMain(t, `package main

func main() {
	x := 42
	p := &x
	print(*p)
}`, "42")
}

func TestPointerAssignment(t *testing.T) {
	testMain(t, `package main

func main() {
	x := 10
	p := &x
	*p = 20
	print(x)
}`, "20")
}

func TestPointerMultipleAssignments(t *testing.T) {
	testMain(t, `package main

func main() {
	x := 1
	y := 2
	px := &x
	py := &y
	*px = 100
	*py = 200
	print(*px, *py)
}`, "100200")
}

func TestPointerToString(t *testing.T) {
	testMain(t, `package main

func main() {
	s := "hello"
	p := &s
	*p = "world"
	print(*p)
}`, "world")
}

func TestPointerSwap(t *testing.T) {
	testMain(t, `package main

func swap(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}
func main() {
	x := 5
	y := 10
	swap(&x, &y)
	print(x, y)
}`, "105")
}

func TestPointerToLocalVariable(t *testing.T) {
	testMain(t, `package main
func main() {
    // Test 1: Pointer reflects variable changes
    a := 1
    b := &a
    print(*b)  // Should be 1
    print(" ")
    a = 2
    print(*b)  // Should be 2 (pointer reflects the change)
    print(" ")
    // Test 2: Writing through pointer updates the variable
    *b = 42
    print(a)   // Should be 42
    print(" ")
    print(*b)  // Should be 42
    print(" ")
    // Test 3: Multiple pointers to same variable
    c := &a
    *c = 100
    print(a)   // Should be 100
    print(" ")
    print(*b)  // Should be 100 (b also points to a)
    print(" ")
    print(*c)  // Should be 100
}`, "1 2 42 42 100 100 100")
}

func TestPointerEscapeFromFunction(t *testing.T) {
	testMain(t, `package main
func ReturnPointer(val int) *int {
	return &val
}
func ModifyThroughPointer(p *int) {
	*p = 999
}
func main() {
	// Test 1: Pointer escapes function scope
	ptr := ReturnPointer(42)
	print(*ptr)  // Should be 42
	print(" ")
	
	// Test 2: Modify through returned pointer
	*ptr = 100
	print(*ptr)  // Should be 100
	print(" ")
	
	// Test 3: Pass pointer to function
	x := 5
	ModifyThroughPointer(&x)
	print(x)  // Should be 999
	print(" ")
	
	// Test 4: String pointers
	s := "hello"
	ps := &s
	print(*ps)  // Should be "hello"
	print(" ")
	s = "world"
	print(*ps)  // Should be "world" (pointer reflects change)
}`, "42 100 999 hello world")
}

func TestComplexPointerScenarios(t *testing.T) {
	testMain(t, `package main
type Point struct {
	X int
	Y int
}
func main() {
	// Test 1: Pointer to struct
	p := Point{X: 10, Y: 20}
	ptr := &p
	print(ptr.X)  // Should be 10
	print(" ")
	p.X = 30
	print(ptr.X)  // Should be 30 (pointer reflects change)
	print(" ")

	// Test 2: Modify struct through pointer
	ptr.Y = 99
	print(p.Y)  // Should be 99
	print(" ")
	
	// Test 3: Pointer arithmetic with different types
	i := 42
	pi := &i
	*pi = *pi + 8
	print(i)  // Should be 50
	print(" ")
	print(*pi)  // Should be 50
}`, "10 30 99 50 50")
}

func TestNewStandardType(t *testing.T) {
	testMain(t, `package main

import "sync"
func main() {
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(1)
	wg.Done()
	print("done")
}`, "done")
}

func TestBlankIdentifier(t *testing.T) {
	testMain(t, `package main

func main() {
	_, h, _ := "gi", "flow", "!"
	print(h)
}`, "flow")
}
