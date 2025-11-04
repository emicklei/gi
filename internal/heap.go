package internal

import (
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
// address is taken, it needs to survive beyond its scope.
type HeapPointer struct {
	Addr       uintptr      // unique address in the heap
	Type       reflect.Type // type of the pointed-to value
	EnvRef     Env          // if non-nil, this points to a variable in an environment
	EnvVarName string       // the variable name in the environment
}

// String formats the HeapPointer to look like a real pointer address.
func (hp HeapPointer) String() string {
	if hp.EnvRef != nil {
		return fmt.Sprintf("0x%x", hp.Addr)
	}
	return fmt.Sprintf("0x%x (%s)", hp.Addr, hp.EnvVarName)
}

// allocHeapValue allocates space in the VM heap for a value and returns a HeapPointer to it.
func (h *Heap) allocHeapValue(v reflect.Value) HeapPointer {
	addr := h.counter
	h.counter++
	h.values[addr] = v
	return HeapPointer{
		Addr: addr,
		Type: v.Type(),
	}
}

// allocHeapVar allocates a heap pointer that references a variable in an environment.
// This is used when taking the address of a variable (like &a).
func (h *Heap) allocHeapVar(env Env, varName string, varType reflect.Type) HeapPointer {
	addr := h.counter
	h.counter++
	return HeapPointer{
		Addr:       addr,
		Type:       varType,
		EnvRef:     env,
		EnvVarName: varName,
	}
}

// read retrieves a value from the VM heap.
func (h *Heap) read(hp HeapPointer) reflect.Value {
	// If this is an environment reference, read from the environment
	if hp.EnvRef != nil {
		return hp.EnvRef.valueLookUp(hp.EnvVarName)
	}
	// Otherwise, read from heap storage
	v, ok := h.values[hp.Addr]
	if !ok {
		panic("invalid heap pointer address")
	}
	return v
}

// write updates a value in the VM heap.
func (h *Heap) write(hp HeapPointer, value reflect.Value) {
	// If this is an environment reference, write to the environment
	if hp.EnvRef != nil {
		hp.EnvRef.set(hp.EnvVarName, value)
		return
	}
	// Otherwise, write to heap storage
	if _, ok := h.values[hp.Addr]; !ok {
		panic("invalid heap address")
	}
	h.values[hp.Addr] = value
}
