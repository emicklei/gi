package pkg

import (
	"reflect"
	"testing"
)

func TestNewHeap(t *testing.T) {
	h := newHeap()
	if h.values == nil {
		t.Error("expected heap values map to be initialized")
	}
	if h.counter != 1 {
		t.Errorf("expected heap counter to start at 1, got %d", h.counter)
	}
}

func TestAllocHeapValue(t *testing.T) {
	h := newHeap()
	val := reflect.ValueOf(42)
	hp := h.allocHeapValue(val)

	if hp.Addr != 1 {
		t.Errorf("expected address 1, got %d", hp.Addr)
	}
	if hp.Type != val.Type() {
		t.Errorf("expected type %v, got %v", val.Type(), hp.Type)
	}
	if h.values[1].Int() != 42 {
		t.Errorf("expected value 42 stored at address 1, got %v", h.values[1])
	}
}

func TestHeapReadWrite(t *testing.T) {
	h := newHeap()
	val := reflect.ValueOf("test")
	hp := h.allocHeapValue(val)

	// Test Read
	readVal := h.read(hp)
	if readVal.String() != "test" {
		t.Errorf("expected read value 'test', got %v", readVal)
	}

	// Test Write
	newVal := reflect.ValueOf("updated")
	h.write(hp, newVal)

	readVal = h.read(hp)
	if readVal.String() != "updated" {
		t.Errorf("expected read value 'updated', got %v", readVal)
	}
}

func TestHeapReadWriteEnv(t *testing.T) {
	h := newHeap()
	env := newEnvironment(nil)
	env.set("x", reflect.ValueOf(10))

	hp := h.allocHeapVar(env, "x", reflect.TypeOf(10))

	// Test Read from Env
	readVal := h.read(hp)
	if readVal.Int() != 10 {
		t.Errorf("expected read value 10, got %v", readVal)
	}

	// Test Write to Env
	h.write(hp, reflect.ValueOf(20))
	if env.valueLookUp("x").Int() != 20 {
		t.Errorf("expected env value updated to 20, got %v", env.valueLookUp("x"))
	}
}

func TestIsHeapPointer(t *testing.T) {
	h := newHeap()
	val := reflect.ValueOf(100)
	hp := h.allocHeapValue(val)

	// Wrap in interface to simulate real usage
	var iface interface{} = hp
	rv := reflect.ValueOf(iface)

	gotHP, ok := isHeapPointer(rv)
	if !ok {
		t.Error("expected isHeapPointer to return true")
	}
	if gotHP != hp {
		t.Error("expected returned HeapPointer to match original")
	}

	// Test non-heap pointer
	x := 10
	rv2 := reflect.ValueOf(&x)
	_, ok = isHeapPointer(rv2)
	if ok {
		t.Error("expected isHeapPointer to return false for standard pointer")
	}
}

func TestHeapPointerString(t *testing.T) {
	hp := &HeapPointer{Addr: 0x123}
	if hp.String() != "0x123" {
		t.Errorf("expected '0x123', got %s", hp.String())
	}

	hpEnv := &HeapPointer{Addr: 0x456, EnvVarName: "myVar", EnvRef: newEnvironment(nil)}
	expected := "0x456 (myVar)"
	if hpEnv.String() != expected {
		t.Errorf("expected '%s', got %s", expected, hpEnv.String())
	}
}

func TestHeapPointerUnmarshalJSON(t *testing.T) {
	// Setup environment with a value
	env := newEnvironment(nil)
	type Person struct {
		Name string
	}
	p := Person{Name: "Alice"}
	env.set("p", reflect.ValueOf(p))

	hp := &HeapPointer{
		EnvRef:     env,
		EnvVarName: "p",
	}

	jsonData := []byte(`{"Name": "Bob"}`)
	err := hp.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the value in environment was updated
	updated := env.valueLookUp("p").Interface().(Person)
	if updated.Name != "Bob" {
		t.Errorf("expected Name to be 'Bob', got %s", updated.Name)
	}
}

func TestHeapInvalidRead(t *testing.T) {
	h := newHeap()
	hp := &HeapPointer{Addr: 999} // Invalid address

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on invalid read")
		}
	}()
	h.read(hp)
}

func TestHeapInvalidWrite(t *testing.T) {
	h := newHeap()
	hp := &HeapPointer{Addr: 999} // Invalid address

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on invalid write")
		}
	}()
	h.write(hp, reflect.ValueOf(1))
}
