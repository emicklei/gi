package internal

import (
	"reflect"
	"testing"
)

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

func TestDeclareVarAndInit(t *testing.T) {
	testMain(t, `package main

func main() {
	var s string = "gi"
	print(s)
}`, "gi")
}

func TestDeclareConstInFunc(t *testing.T) {
	testMain(t, `package main

func main() {
	const s = "gi"
	print(s)
}`, "gi")
}

func TestDeclareConstsInFunc(t *testing.T) {
	testMain(t, `package main

func main() {
	const (
		s = "gi"
		n = 42
	)
	print(s,n)
}`, "gi42")
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

func TestMultiVar(t *testing.T) {
	testMain(t, `package main

func main() {	
	var b, c int = 1, 2
	print(b,c)
}`, "12")
}

func TestIota(t *testing.T) {
	testMain(t, `package main

type state int

const (
	a = iota
	b
	c       = 5
	d state = iota
	e
	f = 1
	g
)

func main() {
	print(a, b, c, d, e, f, g)
}`, "0153411")
}

func TestIotaInFunc(t *testing.T) {
	testMain(t, `package main

func main() {
	const (
		a = iota
		b
	)
	print( a, b)
}`, "01")
}
func TestIotaFloat32InFunc(t *testing.T) {
	testMain(t, `package main

func main() {
	const (
		a float32 = iota + 3.14
		b
	)
	print( a, b)
}`, "3.144.140000000000001")
}

func TestUntypedInt(t *testing.T) {
	const (
		a = 1
		b = a + 2
	)
	at := reflect.ValueOf(a).Type()
	bt := reflect.ValueOf(b).Type()
	a2 := reflect.TypeOf(a)
	t.Log(at, bt, a2, a2.Kind())
}
