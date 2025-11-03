package internal

import (
	"fmt"
	"reflect"
)

// HeapPointer represents a pointer to a value stored in the VM's heap.
// This is used to handle pointer escape analysis - when a local variable's
// address is taken, it needs to survive beyond its scope.
type HeapPointer struct {
	Addr        uintptr      // unique address in the heap
	Type        reflect.Type // type of the pointed-to value
	EnvRef      Env          // if non-nil, this points to a variable in an environment
	EnvVarName  string       // the variable name in the environment
}

// String formats the HeapPointer to look like a real pointer address.
func (hp HeapPointer) String() string {
	return fmt.Sprintf("0x%x", hp.Addr)
}

// allocHeap allocates space in the VM heap for a value and returns a HeapPointer to it.
func (vm *VM) allocHeap(v reflect.Value) HeapPointer {
	addr := vm.heapCounter
	vm.heapCounter++
	vm.heap[addr] = v
	return HeapPointer{
		Addr: addr,
		Type: v.Type(),
	}
}

// allocHeapVar allocates a heap pointer that references a variable in an environment.
// This is used when taking the address of a variable (like &a).
func (vm *VM) allocHeapVar(env Env, varName string, varType reflect.Type) HeapPointer {
	addr := vm.heapCounter
	vm.heapCounter++
	return HeapPointer{
		Addr:       addr,
		Type:       varType,
		EnvRef:     env,
		EnvVarName: varName,
	}
}

// readHeap retrieves a value from the VM heap.
func (vm *VM) readHeap(hp HeapPointer) reflect.Value {
	// If this is an environment reference, read from the environment
	if hp.EnvRef != nil {
		return hp.EnvRef.valueLookUp(hp.EnvVarName)
	}
	// Otherwise, read from heap storage
	v, ok := vm.heap[hp.Addr]
	if !ok {
		vm.fatal("heap pointer to invalid address")
	}
	return v
}

// writeHeap updates a value in the VM heap.
func (vm *VM) writeHeap(hp HeapPointer, value reflect.Value) {
	// If this is an environment reference, write to the environment
	if hp.EnvRef != nil {
		hp.EnvRef.set(hp.EnvVarName, value)
		return
	}
	// Otherwise, write to heap storage
	if _, ok := vm.heap[hp.Addr]; !ok {
		vm.fatal("heap pointer to invalid address")
	}
	vm.heap[hp.Addr] = value
}
