package pkg

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
	valuePos token.Pos     // literal position
	value    reflect.Value // cached evaluated value
}

func newBasicLit(pos token.Pos, v reflect.Value) BasicLit {
	return BasicLit{valuePos: pos, value: v}
}

func (b BasicLit) eval(vm *VM) {
	vm.pushOperand(b.value)
}

func basicLitValue(s *ast.BasicLit) reflect.Value {
	switch s.Kind {
	case token.INT:
		i, _ := strconv.ParseInt(s.Value, 0, 64)
		return reflect.ValueOf(int(i))
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

func (b BasicLit) flow(g *graphBuilder) (head Step) {
	g.next(b)
	return g.current
}

func (b BasicLit) pos() token.Pos { return b.valuePos }

func (b BasicLit) String() string {
	return fmt.Sprintf("BasicLit(%s,%v)", b.value.Kind(), b.value.Interface())
}

var _ Flowable = CompositeLit{}
var _ Expr = CompositeLit{}

type CompositeLit struct {
	lbracePos  token.Pos  // position of "{"
	typ        Expr       // literal type; or nil
	parserType types.Type // literal type; or nil
	elts       []Expr     // list of composite elements; or nil
}

func (c CompositeLit) eval(vm *VM) {

	// if Type is not present, we put all values on the stack as is
	if c.typ == nil {
		values := make([]reflect.Value, len(c.elts))
		for i := range c.elts {
			val := vm.popOperand()
			values[i] = val
		}
		vm.pushOperand(reflect.ValueOf(values))
		return
	}

	values := make([]reflect.Value, len(c.elts))
	for i := range c.elts {
		val := vm.popOperand()
		values[i] = val
	}
	typeOrValue := vm.popOperand().Interface()
	if inst, ok := typeOrValue.(CanMake); ok {
		structVal := inst.makeValue(vm, len(values), nil)
		result := inst.literalCompose(vm, structVal, values)
		vm.pushOperand(result)
	} else {
		vm.fatalf("unhandled type")
	}
}

func (c CompositeLit) flow(g *graphBuilder) (head Step) {
	if c.typ != nil {
		head = c.typ.flow(g)
	}
	// reverse order to have the first element on top of the stack
	for i := len(c.elts) - 1; i >= 0; i-- {
		eltFlow := c.elts[i].flow(g)
		if i == len(c.elts)-1 {
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

func (c CompositeLit) pos() token.Pos { return c.lbracePos }

func (c CompositeLit) String() string {
	return fmt.Sprintf("CompositeLit(%v,%v)", c.typ, c.elts)
}

type Int64 int64
