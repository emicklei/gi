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

func (b BasicLit) Eval(vm *VM) {
	vm.pushOperand(b.value)
}

func basicLitValue(s *ast.BasicLit) reflect.Value {
	switch s.Kind {
	case token.INT:
		i, _ := strconv.Atoi(s.Value)
		return reflect.ValueOf(i)
	case token.STRING:
		unq, err := strconv.Unquote(s.Value)
		if err != nil {
			// fallback for raw strings or edge cases
			unq = strings.Trim(s.Value, "`\"")
		}
		return reflect.ValueOf(unq)
	case token.FLOAT:
		f, _ := strconv.ParseFloat(s.Value, 64)
		return reflect.ValueOf(f)
	case token.CHAR:
		// a character literal is a rune, which is an alias for int32
		return reflect.ValueOf([]rune(s.Value[1 : len(s.Value)-1])[0])
	case token.IMAG:
		i, _ := strconv.ParseComplex(s.Value, 128)
		return reflect.ValueOf(i)
	default:
		panic("not implemented: basicLitValue:" + s.Kind.String())
	}
}

func (b BasicLit) Flow(g *graphBuilder) (head Step) {
	g.next(b)
	return g.current
}

func (b BasicLit) Pos() token.Pos { return b.pos }

func (b BasicLit) String() string {
	return fmt.Sprintf("BasicLit(%v)", b.value.Interface())
}

var _ Flowable = CompositeLit{}
var _ Expr = CompositeLit{}

type CompositeLit struct {
	Lbrace     token.Pos  // position of "{"
	Type       Expr       // literal type; or nil
	ParserType types.Type // literal type; or nil
	Elts       []Expr     // list of composite elements; or nil
}

func (c CompositeLit) Eval(vm *VM) {

	// if Type is not present, we put all values on the stack as is
	if c.Type == nil {
		values := make([]reflect.Value, len(c.Elts))
		for i := range c.Elts {
			val := vm.popOperand()
			values[i] = val
		}
		vm.pushOperand(reflect.ValueOf(values))
		return
	}

	values := make([]reflect.Value, len(c.Elts))
	for i := range c.Elts {
		val := vm.popOperand()
		values[i] = val
	}
	typeOrValue := vm.popOperand().Interface()
	if inst, ok := typeOrValue.(CanMake); ok {
		structVal := inst.Make(vm, len(values), nil)
		result := inst.LiteralCompose(structVal, values)
		vm.pushOperand(result)
	} else {
		vm.fatal("unhandled type")
	}
}

func (c CompositeLit) Flow(g *graphBuilder) (head Step) {
	if c.Type != nil {
		head = c.Type.Flow(g)
	}
	// reverse order to have the first element on top of the stack
	for i := len(c.Elts) - 1; i >= 0; i-- {
		eltFlow := c.Elts[i].Flow(g)
		if i == len(c.Elts)-1 {
			if head == nil {
				head = eltFlow
			}
		}
	}
	g.next(c)
	if head == nil {
		head = g.current
	}
	return
}

func (c CompositeLit) Pos() token.Pos { return c.Lbrace }

func (c CompositeLit) String() string {
	return fmt.Sprintf("CompositeLit(%v,%v)", c.Type, c.Elts)
}
