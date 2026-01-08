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
	NamePos token.Pos
	Names   []Ident
	Type    Expr
	Values  []Expr
	graph   Step
}

func (v ValueSpec) declStep() CanDeclare { return v }

func (v ValueSpec) callGraph() Step {
	return v.graph
}

func (v ValueSpec) declare(vm *VM) bool {
	vm.takeAllStartingAt(v.graph)
	if v.Type == nil {
		for _, idn := range v.Names {
			val := vm.popOperand()
			if isUndeclared(val) {
				// this happens when the value expression is referencing an undeclared variable
				return false
			}
			vm.localEnv().set(idn.Name, val)
		}
		return true
	}
	typ := vm.returnsType(v.Type)
	// left to right, see Flow
	for _, idn := range v.Names {
		if v.Values != nil {
			val := vm.popOperand()
			if !val.IsValid() { // TODO check undeclared?
				return false
			}
			if val.Interface() == untypedNil {
				zv := reflect.Zero(typ)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			// need conversion?
			// if val.CanConvert(typ) {
			// 	cv := val.Convert(typ)
			// 	vm.localEnv().set(idn.Name, cv)
			// 	continue
			// }
			// if cm, ok := typ.(CanMake); ok {
			// 	mv := cm.makeValue(vm, 0, []reflect.Value{val})
			// 	vm.localEnv().set(idn.Name, mv)
			// 	continue
			// }
			// store as is
			vm.localEnv().set(idn.Name, val)
		} else {
			// if nil then zero
			if z, ok := v.Type.(CanMake); ok {
				zv := z.makeValue(vm, 0, nil)
				vm.localEnv().set(idn.Name, zv)
				continue
			}
			// zero value
			zv := reflect.Zero(typ)
			vm.localEnv().set(idn.Name, zv)
		}
	}
	return true
}

func (v ValueSpec) Eval(vm *VM) {}

func (v ValueSpec) flow(g *graphBuilder) (head Step) {
	if v.Values != nil {
		// reverse the order to have first value on top of stack
		for i := len(v.Values) - 1; i >= 0; i-- {
			valFlow := v.Values[i].flow(g)
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
	return fmt.Sprintf("ValueSpec(%v)", v.Names)
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
