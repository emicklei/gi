package internal

import (
	"fmt"
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
	// t.Skip() // AST issue
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
	//t.Skip() // AST issue
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

func TestMethodNoReceiverRef(t *testing.T) {
	testMain(t, `package main

func (_ Aircraft) S() string { return "aircraft" } // put before type on purpose
type Aircraft struct {}
func main() {
	print(Aircraft{}.S())
}`, "aircraft")
}

func TestMethodAccessingField(t *testing.T) {
	testMain(t, `package main

func (a Aircraft) S() string { return a.Model } // put before type on purpose
type Aircraft struct {
	Model string
}
func main() {
	print(Aircraft{Model:"heli"}.S())
}`, "heli")
}

func TestMethodReadingFieldWithArgument(t *testing.T) {
	testMain(t, `package main

func (a Aircraft) S(prefix string) string { return prefix + a.Model } // put before type on purpose
type Aircraft struct {
	Model string
}
func main() {
	print(Aircraft{Model:"heli"}.S("prefix-"))
}`, "prefix-heli")
}

func TestNonPointerMethodWritingFieldWithArgument(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func (a Aircraft) nochange(model string) { a.Model = model }

func main() {
	a := Aircraft{Model:"airplane"}
	a.nochange("balloon")
	print(a.Model)
}`, "airplane")
}

func TestPointerFunctionWritingFieldWithArgument(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func changeModel(a *Aircraft, model string) { a.Model = model }

func main() {
	a := Aircraft{Model:"airplane"}
	changeModel(&a, "heli")
	print(a.Model)
}`, "heli")
}

func TestPointerMethodWritingFieldWithArgument(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func (a *Aircraft) change(model string) { a.Model = model }

func main() {
	a := Aircraft{Model:"airplane"}
	a.change("heli")
	print(a.Model)
}`, "heli")
}

func TestNonPointerFunctionWritingFieldWithArgument(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func nochangeModel(a Aircraft, model string) { a.Model = model }

func main() {
	a := Aircraft{Model:"airplane"}
	nochangeModel(a, "heli")
	print(a.Model)
}`, "airplane")
}

func TestFmtFormat(t *testing.T) {
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
	fmt.Printf("%#v\n",Aircraft{Model: "balloon", Price: 3.14})
}`, "")
}
func TestGoFmtFormat(t *testing.T) {
	type Aircraft struct {
		Model string
		Price float32
		//lint:ignore U1000 // unused
		hidden int
	}
	a := Aircraft{Model: "balloon", Price: 3.14}
	t.Logf("%#v", a)
}

func TestCustomTypeAsWriter(t *testing.T) {
	t.Skip()
	testMain(t, `package main

import "fmt"
	
type writer struct {
	written []byte
}
func (w *writer) Write(p []byte) (n int, err error) {
	w.written = p
	return len(p),nil
}
func main() {
	w := new(writer)
	fmt.Fprint(w,"gi")
	print(string(w.written))
}`, "gi")
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
