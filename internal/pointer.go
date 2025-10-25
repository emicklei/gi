package internal

import (
	"fmt"
	"reflect"
)

// VarPointer represents a pointer to a variable in the environment.
// This is needed because reflect.Value from a map is not addressable,
// so we need to track the environment and variable name to enable
// proper pointer semantics for assignment.
type VarPointer struct {
	env  Env
	name string
}

// Deref returns the value that this pointer points to.
func (vp *VarPointer) Deref() reflect.Value {
	return vp.env.valueLookUp(vp.name)
}

// Assign sets the value that this pointer points to.
func (vp *VarPointer) Assign(value reflect.Value) {
	owner := vp.env.valueOwnerOf(vp.name)
	if owner == nil {
		panic("undefined identifier: " + vp.name)
	}
	owner.set(vp.name, value)
}

func (vp *VarPointer) String() string {
	return fmt.Sprintf("&%s", vp.name)
}
