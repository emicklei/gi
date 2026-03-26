package pkg

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Heap struct {
	values  map[uintptr]reflect.Value // heap storage for escaped pointers
	counter uintptr                   // counter for generating unique heap addresses
}

func newHeap() *Heap {
	return &Heap{
		values:  make(map[uintptr]reflect.Value),
		counter: 1, // start at 1 to avoid zero-value confusion
	}
}

// HeapPointer represents a pointer to a value stored in the VM's heap.
// This is used to handle pointer escape analysis - when a local variable's
// address is taken, it needs to survive beyond its scope (in gi terms: environment).
type HeapPointer struct {
	addr       uintptr      // unique address in the heap
	typ        reflect.Type // type of the pointed-to value
	env        Env          // if non-nil, this points to a variable in an environment
	envVarName string       // the variable name in the environment
	// tryout
	ptrValue reflect.Value
}

func (hp *HeapPointer) UnmarshalJSON(data []byte) error {
	if hp.env == nil {
		return nil
	}
	deref := hp.env.valueLookUp(hp.envVarName)
	val := deref.Interface()
	if i, ok := val.(StructValue); ok {
		return i.UnmarshalJSON(data)
	}
	// For standard types, we need to update the value in the environment
	ptr := reflect.New(deref.Type())
	ptr.Elem().Set(deref)
	if err := json.Unmarshal(data, ptr.Interface()); err != nil {
		return err
	}
	hp.env.valueSet(hp.envVarName, ptr.Elem())
	return nil
}

// String formats the HeapPointer to look like a real pointer address.
func (hp *HeapPointer) String() string {
	if hp.env != nil {
		return fmt.Sprintf("0x%x (%s) {env=%p} {ptr=%p}", hp.addr, hp.envVarName, hp.env, hp)
	}
	return fmt.Sprintf("0x%x", hp.addr)
}

// TODO inline?
func asHeapPointer(rv reflect.Value) (hp *HeapPointer, ok bool) {
	if rv.Kind() != reflect.Pointer {
		return nil, false
	}
	if rv.IsZero() {
		return nil, false
	}
	if !rv.CanInterface() {
		return nil, false
	}
	hp, ok = rv.Interface().(*HeapPointer)
	return
}

// allocHeapValue allocates space in the VM heap for a value and returns a HeapPointer to it.
func (h *Heap) allocHeapValue(v reflect.Value) *HeapPointer {
	addr := h.counter
	h.counter++
	h.values[addr] = v
	return &HeapPointer{
		addr:     addr,
		typ:      v.Type(),
		ptrValue: reflect.New(v.Type()),
	}
}

// allocHeapVar allocates a heap pointer that references a variable in an environment.
// This is used when taking the address of a variable (like &a).
func (h *Heap) allocHeapVar(env Env, varName string, varType reflect.Type) *HeapPointer {
	addr := h.counter
	h.counter++
	return &HeapPointer{
		addr:       addr,
		typ:        varType,
		env:        env,
		envVarName: varName,
		ptrValue:   reflect.New(varType),
	}
}

// read retrieves a value from the VM heap.
func (h *Heap) read(hp *HeapPointer) reflect.Value {
	// If this is an environment reference, read from the environment
	if hp.env != nil {
		val := hp.env.valueLookUp(hp.envVarName)
		hp.ptrValue.Elem().Set(val)
		return val
	}
	// Otherwise, read from heap storage
	v, ok := h.values[hp.addr]
	if !ok {
		panic("invalid heap pointer address")
	}
	return v
}

// write updates a value in the VM heap.
func (h *Heap) write(hp *HeapPointer, value reflect.Value) {
	// If this is an environment reference, write to the environment
	if hp.env != nil {
		hp.env.valueSet(hp.envVarName, value)
		return
	}
	// Otherwise, write to heap storage
	if _, ok := h.values[hp.addr]; !ok {
		panic("invalid heap address")
	}
	h.values[hp.addr] = value
}
