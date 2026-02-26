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

type ConstVar struct {
	ident   Ident
	namePos token.Pos
	typ     Expr
	value   Expr
}

// push the result (true,false) of the declaration onto the stack
func (cv ConstVar) eval(vm *VM) {
	// value is on the stack
	if cv.typ == nil {
		val := vm.popOperand()
		if isUndeclared(val) {
			// this happens when the value expression is referencing an undeclared variable
			vm.pushOperand(reflectFalse)
			return
		}
		vm.currentEnv().valueSet(cv.ident.name, val)
		vm.pushOperand(reflectTrue)
		return
	}
	if cv.value != nil {
		val := vm.popOperand()
		if val == reflectNil {
			typ := makeType(vm, cv.typ)
			zv := reflect.Zero(typ)
			vm.currentEnv().valueSet(cv.ident.name, zv)
			vm.pushOperand(reflectTrue)
			return
		}
		if val.Interface() == untypedNil {
			typ := makeType(vm, cv.typ)
			zv := reflect.Zero(typ)
			vm.currentEnv().valueSet(cv.ident.name, zv)
			vm.pushOperand(reflectTrue)
			return
		}
		typ := typeMaker(vm, cv.typ)
		mv := typ.makeValue(vm, 0, []reflect.Value{val})
		vm.currentEnv().valueSet(cv.ident.name, mv)
	} else {
		// if nil then zero
		if z, ok := cv.typ.(CanMake); ok {
			zv := z.makeValue(vm, 0, nil)
			vm.currentEnv().valueSet(cv.ident.name, zv)
			vm.pushOperand(reflectTrue)
			return
		}
		// zero value
		typ := makeType(vm, cv.typ)
		zv := reflect.Zero(typ)
		vm.currentEnv().valueSet(cv.ident.name, zv)
	}
	vm.pushOperand(reflectTrue)
}

func (cv ConstVar) flow(g *graphBuilder) (head Step) {
	if cv.value != nil {
		head = cv.value.flow(g)
	}
	g.next(cv)
	if head == nil {
		head = g.current
	}
	return
}
func (cv ConstVar) pos() token.Pos {
	return cv.namePos
}
func (cv ConstVar) String() string {
	return fmt.Sprintf("ConstVar(%s)", cv.ident)
}

var _ Decl = ValueSpec{}
var _ CanDeclare = ValueSpec{}

var _ Stmt = ValueSpec{}

// Const or Var declaration
type ValueSpec struct {
	names   []Ident
	namePos token.Pos
	typ     Expr
	values  []Expr
	graph   Step
}

func (v ValueSpec) stmtStep() Evaluable  { return nil } //  unused
func (v ValueSpec) declStep() CanDeclare { return v }

func (v ValueSpec) callGraph() Step {
	return v.graph
}

func (v ValueSpec) declare(vm *VM) bool {
	vm.stepThrough(v.graph) // TODO
	return v.processLHS(vm)
}

func (v ValueSpec) processLHS(vm *VM) bool {
	if v.typ == nil {
		for _, idn := range v.names {
			val := vm.popOperand()
			if isUndeclared(val) {
				// this happens when the value expression is referencing an undeclared variable
				return false
			}
			vm.currentEnv().valueSet(idn.name, val)
		}
		return true
	}
	typ := typeMaker(vm, v.typ)

	// left to right, see Flow
	for _, idn := range v.names {
		if v.values != nil {
			val := vm.popOperand()
			if val == reflectNil {
				typ := makeType(vm, v.typ)
				zv := reflect.Zero(typ)
				vm.currentEnv().valueSet(idn.name, zv)
				continue
			}
			if val.Interface() == untypedNil {
				typ := makeType(vm, v.typ)
				zv := reflect.Zero(typ)
				vm.currentEnv().valueSet(idn.name, zv)
				continue
			}
			mv := typ.makeValue(vm, 0, []reflect.Value{val})
			vm.currentEnv().valueSet(idn.name, mv)
		} else {
			// if nil then zero
			if z, ok := v.typ.(CanMake); ok {
				zv := z.makeValue(vm, 0, nil)
				vm.currentEnv().valueSet(idn.name, zv)
				continue
			}
			// zero value
			typ := makeType(vm, v.typ)
			zv := reflect.Zero(typ)
			vm.currentEnv().valueSet(idn.name, zv)
		}
	}
	return true
}

func (v ValueSpec) eval(vm *VM) {
	// process all declarations results and push true/false on stack
	// TODO optimize
	result := reflectTrue
	for range v.names {
		declared := vm.popOperand()
		if declared == reflectFalse {
			result = reflectFalse
			// continue processing to pop all declaration results, but the overall result is false
		}
	}
	vm.pushOperand(result)
}

// var a,b int = 1,2 => var a = 1 ; var b = 2
func (v ValueSpec) flow(g *graphBuilder) (head Step) {
	for i, name := range v.names {
		cv := ConstVar{
			namePos: v.namePos,
			ident:   name,
			typ:     v.typ,
		}
		if v.values != nil {
			cv.value = v.values[i]
		}
		cvFlow := cv.flow(g)
		if i == 0 {
			head = cvFlow
		}
	}
	g.next(v)
	return

	// if v.values != nil {
	// 	// reverse the order, right-to-left, to have first value on top of stack
	// 	for i := len(v.values) - 1; i >= 0; i-- {
	// 		valFlow := v.values[i].flow(g)
	// 		if i == len(v.values)-1 {
	// 			head = valFlow
	// 		}
	// 	}
	// }
	// if head == nil {
	// 	head = g.current
	// }
	// return
}

func (v ValueSpec) pos() token.Pos {
	return v.namePos
}

func (v ValueSpec) String() string {
	return fmt.Sprintf("ValueSpec(%v)", v.names)
}

var _ Expr = new(iotaExpr)

// represents successive untyped integer constants
type iotaExpr struct {
	count   int
	exprPos token.Pos
}

func (i *iotaExpr) reset() {
	i.count = 0
}

func (i *iotaExpr) next() {
	i.count++
}

func (i *iotaExpr) eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(i.count))
}
func (i *iotaExpr) flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}
func (i *iotaExpr) pos() token.Pos {
	return i.exprPos
}
func (i *iotaExpr) String() string {
	return fmt.Sprintf("iota(%d)", i.count)
}
