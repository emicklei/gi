package internal

import "testing"

func TestTypeAlias(t *testing.T) {
	testMain(t, `package main

type MyInt = int

func main() {
	var a MyInt = 1
	print(a)
}`, "1")
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

func TestNewTypeUnexported(t *testing.T) {
	testMain(t, `package main

type airplane struct {
	capacity int
}
func main() {
	heli := airplane{capacity: 50}
	print(heli.capacity)
}`, "50")
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

func TestMethod(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func (_ Airplane) S() string { return "airplane" } // put before type on purpose
type Airplane struct {}
func main() {
	print(Airplane{}.S())
}`, "airplane")
}
