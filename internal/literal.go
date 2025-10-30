package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

var _ Expr = BasicLit{}

type BasicLit struct {
	*ast.BasicLit
}

func (s BasicLit) Eval(vm *VM) {
	switch s.Kind {
	case token.INT:
		i, _ := strconv.Atoi(s.Value)
		vm.pushOperand(reflect.ValueOf(i))
	case token.STRING:
		unq := strings.Trim(s.Value, "`\"")
		vm.pushOperand(reflect.ValueOf(unq))
	case token.FLOAT:
		f, _ := strconv.ParseFloat(s.Value, 64)
		vm.pushOperand(reflect.ValueOf(f))
	case token.CHAR:
		// a character literal is a rune, which is an alias for int32
		vm.pushOperand(reflect.ValueOf(s.Value))
	case token.IMAG:
		i, _ := strconv.ParseComplex(s.Value, 128)
		vm.pushOperand(reflect.ValueOf(i))
	default:
		panic("not implemented: BasicList.Eval:" + s.Kind.String())
	}
}
func (s BasicLit) Loc(f *token.File) string {
	return fmt.Sprintf("%v:BasicLit(%v)", f.Position(s.Pos()), s.Value)
}
func (s BasicLit) String() string {
	return fmt.Sprintf("BasicLit(%v)", s.Value)
}

func (s BasicLit) Flow(g *graphBuilder) (head Step) {
	g.next(s)
	return g.current
}

var _ Flowable = CompositeLit{}
var _ Expr = CompositeLit{}

type CompositeLit struct {
	*ast.CompositeLit
	Type Expr
	Elts []Expr
}

func (s CompositeLit) Eval(vm *VM) {
	internalType := vm.returnsEval(s.Type).Interface()
	i, ok := internalType.(CanInstantiate)
	if !ok {
		vm.fatal(fmt.Sprintf("expected CanInstantiate:%v (%T)", internalType, internalType))
	}
	instance := i.Instantiate(vm)
	values := make([]reflect.Value, len(s.Elts))
	for i, elt := range s.Elts {
		var val reflect.Value
		if vm.isStepping {
			// see Flow for the order of pushing
			val = vm.frameStack.top().pop()
		} else {
			val = vm.returnsEval(elt)
		}
		values[i] = val
	}
	result := i.LiteralCompose(instance, values)
	vm.pushOperand(result)
}

func (s CompositeLit) String() string {
	return fmt.Sprintf("CompositeLit(%v,%v)", s.Type, s.Elts)
}

func (s CompositeLit) Flow(g *graphBuilder) (head Step) {
	// reverse order to have the first element on top of the stack
	for i := len(s.Elts) - 1; i >= 0; i-- {
		each := s.Elts[i]
		if i == len(s.Elts)-1 {
			head = each.Flow(g)
			continue
		}
		each.Flow(g)
	}
	g.next(s)
	// without elements, head is the current step
	if head == nil {
		head = g.current
	}
	return head
}

var _ Expr = FuncLit{}

type FuncLit struct {
	*ast.FuncLit
	Type      *FuncType
	Body      *BlockStmt // TODO not sure what to do when Body and/or Type is nil
	callGraph Step       // TODO used?
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
