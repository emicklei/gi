package internal

import (
	"fmt"
	"testing"
)

func TestTypeAlias(t *testing.T) {
	testMain(t, `package main

type MyInt = int

func main() {
	var a MyInt = 1
	print(a)
}`, "1")
}

func TestTypeAlias2(t *testing.T) {
	testMain(t, `package main

type HTML string

func main() {
	print(HTML("gi"))
}`, "gi")
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

func TestNewType(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func main() {
	heli := Aircraft{Model:"helicopter"}
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

func TestTypeMarshalJSON(t *testing.T) {
	t.Skip() // AST issue
	testMain(t, fmt.Sprintf(`package main

import "encoding/json"

type Aircraft struct {
	Model string %s
	Registration string %s
	Brand string %s
	owner string
}
func main() {
	heli := Aircraft{Model:"helicopter", Registration:"PH-EMM"}
	data, _ := json.Marshal(heli)
	print(string(data))
}`, "`json:\"model\"`", "`json:\"-\"`", "`json:\"brand,omitempty\"`"), `{"model":"helicopter"}`)
}

func TestTypeUnmarshalJSON(t *testing.T) {
	t.Skip() // AST issue
	testMain(t, fmt.Sprintf(`package main

import "encoding/json"

type Aircraft struct {
	Model string %s
	Registration string %s
	Brand string %s
	owner string
}
func main() {
	content := %s
	heli := Aircraft{}
	json.Unmarshal([]byte(content), &heli)
	print(heli.Model)
}`, "`json:\"model\"`",
		"`json:\"-\"`",
		"`json:\"brand,omitempty\"`",
		"`{\"model\":\"helicopter\"}`"), "helicopter")
}
func TestTypeMarshalXML(t *testing.T) {
	t.Skip()
	testMain(t, fmt.Sprintf(`package main

import "encoding/xml"

type Aircraft struct {
	Model string %s
	Registration string %s
	Brand string %s
	owner string
}
func main() {
	heli := Aircraft{Model:"helicopter", Registration:"PH-EMM"}
	data, _ := xml.Marshal(heli)
	print(string(data))
}`, "`xml:\"model\"`", "`xml:\"-\"`", "`xml:\"brand,omitempty\"`"), `<Aircraft><model>helicopter</model></Aircraft>`)
}

func TestAddressOfType(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func main() {
	heli := &Aircraft{Model:"helicopter"}
	print(heli.Model)
}`, "helicopter")
}

func TestMethod(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func (_ Aircraft) S() string { return "aircraft" } // put before type on purpose
type Aircraft struct {}
func main() {
	print(Aircraft{}.S())
}`, "aircraft")
}

func TestFmtFormat(t *testing.T) {
	t.Skip()
	testMain(t, `package main
import "fmt"
// import "bytes"
type Aircraft struct {
	Model string
	Price float32
	hidden int
}
func main() {
	// var buf bytes.Buffer
	fmt.Printf("%#v",Aircraft{Model: "balloon", Price: 3.14})
}`, "")
}
func TestGoFmtFormat(t *testing.T) {
	type Aircraft struct {
		Model  string
		Price  float32
		hidden int
	}
	a := Aircraft{Model: "balloon", Price: 3.14}
	t.Logf("%#v", a)
}
