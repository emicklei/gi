package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

func isUndeclared(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	return v == reflectUndeclared
}

var _ Decl = ValueSpec{}
var _ CanDeclare = ValueSpec{}

// Const or Var declaration
type ValueSpec struct {
	namePos token.Pos
	names   []Ident
	typ     Expr
	values  []Expr
	graph   Step
}

func (v ValueSpec) declStep() CanDeclare { return v }

func (v ValueSpec) callGraph() Step {
	return v.graph
}

func (v ValueSpec) declare(vm *VM) bool {
	vm.takeAllStartingAt(v.graph)
	if v.typ == nil {
		for _, idn := range v.names {
			val := vm.popOperand()
			if isUndeclared(val) {
				// this happens when the value expression is referencing an undeclared variable
				return false
			}
			vm.localEnv().set(idn.name, val)
		}
		return true
	}
	typ := vm.proxyType(v.typ)

	// left to right, see Flow
	for _, idn := range v.names {
		if v.values != nil {
			val := vm.popOperand()
			if val == reflectNil {
				typ := vm.makeType(v.typ)
				zv := reflect.Zero(typ)
				vm.localEnv().set(idn.name, zv)
				continue
			}
			if val.Interface() == untypedNil {
				typ := vm.makeType(v.typ)
				zv := reflect.Zero(typ)
				vm.localEnv().set(idn.name, zv)
				continue
			}
			mv := typ.makeValue(vm, 0, []reflect.Value{val})
			vm.localEnv().set(idn.name, mv)
		} else {
			// if nil then zero
			if z, ok := v.typ.(CanMake); ok {
				zv := z.makeValue(vm, 0, nil)
				vm.localEnv().set(idn.name, zv)
				continue
			}
			// zero value
			typ := vm.makeType(v.typ)
			zv := reflect.Zero(typ)
			vm.localEnv().set(idn.name, zv)
		}
	}
	return true
}

func (v ValueSpec) Eval(vm *VM) {}

func (v ValueSpec) flow(g *graphBuilder) (head Step) {
	if v.values != nil {
		// reverse the order to have first value on top of stack
		for i := len(v.values) - 1; i >= 0; i-- {
			valFlow := v.values[i].flow(g)
			if i == len(v.values)-1 {
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
	return v.namePos
}

func (v ValueSpec) String() string {
	return fmt.Sprintf("ValueSpec(%v)", v.names)
}

var _ Expr = new(iotaExpr)

// represents successive untyped integer constants
type iotaExpr struct {
	pos   token.Pos
	count int
}

func (i *iotaExpr) reset() {
	i.count = 0
}

func (i *iotaExpr) next() {
	i.count++
}

func (i *iotaExpr) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(i.count))
}
func (i *iotaExpr) flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}
func (i *iotaExpr) Pos() token.Pos {
	return i.pos
}
func (i *iotaExpr) String() string {
	return fmt.Sprintf("iota(%d)", i.count)
}
