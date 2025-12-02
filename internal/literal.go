package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
	"strings"
)

var _ Expr = BasicLit{}

type BasicLit struct {
	pos   token.Pos     // literal position
	value reflect.Value // cached evaluated value
}

func newBasicLit(pos token.Pos, v reflect.Value) BasicLit {
	return BasicLit{pos: pos, value: v}
}

func (s BasicLit) Eval(vm *VM) {
	vm.pushOperand(s.value)
}

func basicLitValue(s *ast.BasicLit) reflect.Value {
	switch s.Kind {
	case token.INT:
		i, _ := strconv.Atoi(s.Value)
		return reflect.ValueOf(i)
	case token.STRING:
		unq := strings.Trim(s.Value, "`\"")
		return reflect.ValueOf(unq)
	case token.FLOAT:
		f, _ := strconv.ParseFloat(s.Value, 64)
		return reflect.ValueOf(f)
	case token.CHAR:
		// a character literal is a rune, which is an alias for int32
		return reflect.ValueOf(s.Value)
	case token.IMAG:
		i, _ := strconv.ParseComplex(s.Value, 128)
		return reflect.ValueOf(i)
	default:
		panic("not implemented: basicLitValue:" + s.Kind.String())
	}
}

func (s BasicLit) Flow(g *graphBuilder) (head Step) {
	g.next(s)
	return g.current
}

func (s BasicLit) Pos() token.Pos { return s.pos }

func (s BasicLit) String() string {
	return fmt.Sprintf("BasicLit(%v)", s.value.Interface())
}

var _ Flowable = CompositeLit{}
var _ Expr = CompositeLit{}

type CompositeLit struct {
	Lbrace     token.Pos  // position of "{"
	Type       Expr       // literal type; or nil
	ParserType types.Type // literal type; or nil
	Elts       []Expr     // list of composite elements; or nil
}

func (s CompositeLit) Eval(vm *VM) {

	// if Type is not present, we put all values on the stack as is
	if s.Type == nil {
		values := make([]reflect.Value, len(s.Elts))
		for i := range s.Elts {
			val := vm.callStack.top().pop()
			values[i] = val
		}
		vm.pushOperand(reflect.ValueOf(values))
		return
	}

	values := make([]reflect.Value, len(s.Elts))
	for i := range s.Elts {
		val := vm.callStack.top().pop()
		values[i] = val
	}
	typeOrValue := vm.callStack.top().pop().Interface()
	if inst, ok := typeOrValue.(CanInstantiate); ok {
		instance := inst.Instantiate(vm, len(values), nil)
		result := inst.LiteralCompose(instance, values)
		vm.pushOperand(result)
	} else {
		vm.fatal("unhandled type")
	}
}

func (s CompositeLit) Flow(g *graphBuilder) (head Step) {
	if s.Type != nil {
		head = s.Type.Flow(g)
	}
	// reverse order to have the first element on top of the stack
	for i := len(s.Elts) - 1; i >= 0; i-- {
		eltFlow := s.Elts[i].Flow(g)
		if i == len(s.Elts)-1 {
			if head == nil {
				head = eltFlow
			}
		}
	}
	g.next(s)
	if head == nil {
		head = g.current
	}
	return
}

func (s CompositeLit) Pos() token.Pos { return s.Lbrace }

func (s CompositeLit) String() string {
	return fmt.Sprintf("CompositeLit(%v,%v)", s.Type, s.Elts)
}

var _ Expr = FuncLit{}

type FuncLit struct {
	*ast.FuncLit
	Type      *FuncType
	Body      *BlockStmt // TODO not sure what to do when Body and/or Type is nil
	callGraph Step
}

func (s FuncLit) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(s))
}

func (s FuncLit) Flow(g *graphBuilder) (head Step) {
	g.next(s)
	return g.current
}

func (s FuncLit) String() string {
	return fmt.Sprintf("FuncLit(%v,%v)", s.Type, s.Body)
}
