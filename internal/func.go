package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

type statementReference struct {
	step  Step
	index int
}

type FuncDecl struct {
	Name *Ident
	Recv *FieldList // non-nil for methods
	Body *BlockStmt
	Type *FuncType
	// control flow graph
	callGraph Step
	// goto targets
	labelToStmt map[string]statementReference // TODO lazy initialization
	// for source access of any statement/expression within this function
	fileSet        *token.FileSet
	hasRecoverCall bool
}

func (f FuncDecl) Eval(vm *VM) {} // noop

func (f FuncDecl) Flow(g *graphBuilder) (head Step) {
	head = g.current
	if f.Body != nil {
		g.funcStack.push(f)
		head = f.Body.Flow(g)
		g.funcStack.pop()
	}
	return
}

func (f FuncDecl) Pos() token.Pos { return f.Type.Pos() }

func (f FuncDecl) String() string {
	return fmt.Sprintf("FuncDecl(%s)", f.Name.Name)
}

var _ Expr = FuncType{}

type FuncType struct {
	FuncPos    token.Pos
	TypeParams *FieldList // type parameters; or nil
	Params     *FieldList // (incoming) parameters; non-nil
	Results    *FieldList // (outgoing) results; or nil
}

func (t FuncType) Eval(vm *VM) {}

func (t FuncType) Flow(g *graphBuilder) (head Step) {
	// TODO
	return g.current
}

func (t FuncType) Pos() token.Pos { return t.FuncPos }

func (t FuncType) String() string {
	return fmt.Sprintf("FuncType(%v,%v,%v)", t.TypeParams, t.Params, t.Results)
}

var _ Expr = Ellipsis{}

type Ellipsis struct {
	*ast.Ellipsis
	Elt Expr // ellipsis element type (parameter lists only); or nil
}

func (e Ellipsis) String() string {
	return fmt.Sprintf("Ellipsis(%v)", e.Elt)
}
func (e Ellipsis) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(e))
}

func (e Ellipsis) Flow(g *graphBuilder) (head Step) {
	if e.Elt != nil {
		head = e.Elt.Flow(g)
	} else {
		g.next(e)
		return g.current
	}
	return
}

// funcInvocation represents a function call instance with its own environment.
// this used to handle defer statements properly.
type funcInvocation struct {
	flow Step
	env  Env
}

func isRecoverCall(expr Expr) bool {
	if ident, ok := expr.(Ident); ok {
		return ident.Name == "recover"
	}
	return false
}
