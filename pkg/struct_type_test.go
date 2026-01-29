package pkg

import (
	"fmt"
	"strings"
	"testing"
)

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

func TestInstantiateIType(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func main() {
	heli := Aircraft{}
	print(heli)
}`, `Aircraft{Model:""}`)
}

func TestInstantiateITypeWithField(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func main() {
	heli := Aircraft{Model:"heli"}
	print(heli)
}`, `Aircraft{Model:"heli"}`)
}

func TestNewIType(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func main() {
	heli := new(Aircraft)
	print(heli)
}`, func(out string) bool { return strings.HasPrefix(out, "0x") })
}

func TestNewITypeSetField(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {
	Model string
}
func main() {
	heli := new(Aircraft)
	heli.Model = "heli"
	print(heli.Model)
}`, "heli")
}

func TestITypeMethodNoReceiverRef(t *testing.T) {
	testMain(t, `package main

func (_ Aircraft) S() string { return "aircraft" } // put before type on purpose
type Aircraft struct {}
func main() {
	print(Aircraft{}.S())
}`, "aircraft")
}

func TestMITypeethodAccessingField(t *testing.T) {
	testMain(t, `package main

func (a Aircraft) S() string { return a.Model } // put before type on purpose
type Aircraft struct {
	Model string
}
func main() {
	print(Aircraft{Model:"heli"}.S())
}`, "heli")
}

func TestITypeMethodReadingFieldWithArgument(t *testing.T) {
	testMain(t, `package main

func (a Aircraft) S(prefix string) string { return prefix + a.Model } // put before type on purpose
type Aircraft struct {
	Model string
}
func main() {
	print(Aircraft{Model:"heli"}.S("prefix-"))
}`, "prefix-heli")
}

func TestITypeNonPointerMethodWritingFieldWithArgument(t *testing.T) {
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

func TestITypePointerFunctionWritingFieldWithArgument(t *testing.T) {
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

func TestITypeNonPointerFunctionWritingFieldWithArgument(t *testing.T) {
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

func TestITypeFmtFormat(t *testing.T) {
	testMain(t, `package main
import "fmt"
type Aircraft struct {
	Model string
	Price float32
	hidden int
}
func main() {
	fmt.Printf("%#v\n",Aircraft{Model: "balloon", Price: 3.14})
}`, "")
}

func TestSameVarAndField(t *testing.T) {
	testMain(t, `package main

type Aircraft struct {model string}
func main() {
	model := "plane"
	a := Aircraft{model:model}
	b := map[string]string{model: model}
	
	print(a.model)
	print(b[model])
}`, "planeplane")
}

func TestITypeArray(t *testing.T) {
	t.Skip()
	testMain(t, `package main
type Aircraft struct {Model string}
func main() {
	a := [2]Aircraft{{Model:"a"},{Model:"b"}}
	print(a[0].Model)
	print(a[1].Model)
}`, "ab")
}

func TestITypeAsWriter(t *testing.T) {
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

// panic: reflect.Value.SetMapIndex: value of type pkg.ExtendedValue is not assignable to type *pkg.StructValue
func TestExtendedTypeAsMapKey(t *testing.T) {
	t.Skip()
	testMain(t, `package main
type Count int
func main() {
	one := Count(1)
	m := map[Count]int{one:1}
	print(len(m),m[one])
}`, "11")
}
