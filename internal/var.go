package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Decl = ValueSpec{}
var _ CanDeclare = ValueSpec{}

// Const or Var declaration
type ValueSpec struct {
	NamePos   token.Pos
	Names     []*Ident
	Type      Expr
	Values    []Expr
	callGraph Step
}

func (v ValueSpec) declStep() CanDeclare { return v }

func (v ValueSpec) CallGraph() Step {
	return v.callGraph
}

func (v ValueSpec) Declare(vm *VM) bool {
	if v.Type == nil {
		for i, idn := range v.Names {
			var val reflect.Value
			// no operand could mean iota active
			if len(vm.callStack.top().operands) == 0 {
				if vm.declIota != nil {
					val = reflect.ValueOf(vm.declIota.value())
				} else {
					vm.fatal("todo")
				}
			} else {
				val = vm.callStack.top().pop()
				if val == reflectUndeclared {
					// this happens when the value expression is referencing an undeclared variable
					return false
				}
				// if itoa was not used but active, fix it
				if vm.declIota != nil {
					// TODO this check is not perfect
					if _, ok := v.Values[i].(*Iota); !ok {
						vm.declIota.fixedValue = val.Interface().(int)
						vm.declIota.fixed = true
					}
				}
			}
			vm.localEnv().set(idn.Name, val)
		}
		return true
	}
	typ := vm.returnsType(v.Type)
	// left to right, see Flow
	for _, idn := range v.Names {
		if v.Values != nil {
			val := vm.callStack.top().pop()
			if !val.IsValid() { // TODO check undeclared?
				return false
			}
			if val.Interface() == untypedNil {
				zv := reflect.Zero(typ)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			vm.localEnv().set(idn.Name, val)
		} else {
			// if nil then zero
			if z, ok := v.Type.(CanInstantiate); ok {
				zv := z.Instantiate(vm, 0, nil)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			// zero primitive
			zv := reflect.Zero(typ)
			vm.localEnv().set(idn.Name, zv)
		}
	}
	return true
}

func (v ValueSpec) Eval(vm *VM) {
	typ := vm.returnsType(v.Type)
	// left to right, see Flow
	for _, idn := range v.Names {
		if v.Values != nil {
			val := vm.callStack.top().pop()
			if !val.IsValid() {
				return
			}
			if val.Interface() == untypedNil {
				zv := reflect.Zero(typ)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			vm.localEnv().set(idn.Name, val)
		} else {
			// if nil then zero
			if z, ok := v.Type.(CanInstantiate); ok {
				zv := z.Instantiate(vm, 0, nil)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			// zero primitive
			typ := vm.returnsType(v.Type)
			zv := reflect.Zero(typ)
			vm.localEnv().set(idn.Name, zv)
		}
	}
}

func (v ValueSpec) Flow(g *graphBuilder) (head Step) {
	if v.Values != nil {
		// reverse the order to have first value on top of stack
		for i := len(v.Values) - 1; i >= 0; i-- {
			valFlow := v.Values[i].Flow(g)
			if i == len(v.Values)-1 {
				head = valFlow
			}
		}
	}
	if head == nil {
		head = g.current
	}
	return
}

func (v ValueSpec) Pos() token.Pos {
	return v.NamePos
}

func (v ValueSpec) String() string {
	return fmt.Sprintf("ValueSpec(len=%d)", len(v.Names))
}

var _ Expr = new(Iota)

// represents successive untyped integer constants
type Iota struct {
	pos        token.Pos
	count      int
	fixed      bool // set to true when iota value is overridden with explicit value
	fixedValue int
}

func (i *Iota) value() int {
	if i.fixed {
		return i.fixedValue
	}
	return i.count
}

func (i *Iota) next() int {
	if i.fixed {
		return i.fixedValue
	}
	i.count++
	return i.count
}
func (i *Iota) Eval(vm *VM) {
	if vm.declIota == nil {
		vm.declIota = i
	} else {
		// replacing?
		if vm.declIota != i {
			// copy state
			i.count = vm.declIota.count + 1
			vm.declIota = i
		}
	}
	if vm.declIota.fixed {
		vm.declIota.fixed = false
	}
	// use the VM's iota value
	vm.pushOperand(reflect.ValueOf(vm.declIota.value()))
}
func (i *Iota) Flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}
func (i *Iota) Pos() token.Pos {
	return i.pos
}
