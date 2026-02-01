package pkg

import (
	"reflect"
	"testing"
)

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

func TestIntExtended(t *testing.T) {
	testMain(t, `package main

type MyInt = int

func main() {
	var a MyInt = 1
	print(a)
}`, "1")
}

func TestPointerIntExtended(t *testing.T) {
	testMain(t, `package main

// type pri = *int
type pri *int

func main() {
	n := 42
	var a pri = &n
	print(*a)
}`, "42")
}

func TestExtendedString(t *testing.T) {
	testMain(t, `package main

type HTML string

func main() {
	print(HTML("gi"))
}`, "gi")
}

func TestTypeDecoratedConstIota(t *testing.T) {
	testMain(t, `package main

type Count int

const (
	Zero Count = iota
	One
)

func main() {
	print(Zero, One)
}`, "01")
}

func TestTypeDecoratedConstIotaWithMethod(t *testing.T) {
	testMain(t, `package main

import "fmt"

type Count int

func (c Count) String() string {
	return fmt.Sprintf("%d",c)
}

const (
	Zero Count = iota
)

func main() {
	print(Zero.String())
}`, "0")
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

func TestTypeAssert(t *testing.T) {
	testMain(t, `package main
func main() {
	var i any
	i = 42
	j := i.(int)
	print(j)
}`, "42")
}

func TestTypeAssertOk(t *testing.T) {
	testMain(t, `package main
func main() {
	var i any
	i = 42
	if j,ok := i.(int); ok {
		print(j)
	}
}`, "42")
}

func TestConvertArgumentType(t *testing.T) {
	testMain(t, `package main
import "math"
func main() {
	print(math.Sin(1))
}`, "0.8414709848078965")
}

func TestConvertArgumentPointerType(t *testing.T) {
	testMain(t, `package main
import "flag"
func main() {
	var name string
	flag.StringVar(&name, "name", "World", "a name to say hello to")
	flag.Parse()
	print(name)
}`, "World")
}

func TestNewNumber(t *testing.T) {
	testMain(t, `package main
 
func main() {
	x := new(int64)
	*x = 40
	print(*x)	
}`, "40")
}

func TestGoType(t *testing.T) {
	gt := SDKType{typ: reflect.TypeOf(42)}
	rv := gt.makeValue(nil, 0, nil)
	t.Log(rv)
	{
		pi := new(int)
		*pi = 42
		gt := SDKType{typ: reflect.TypeOf(pi)}
		rv := gt.makeValue(nil, 0, nil)
		t.Log(rv)
	}
}
