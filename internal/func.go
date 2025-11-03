package internal

import (
	"fmt"
	"go/ast"
	"go/token"
)

type statementReference struct {
	step  Step
	index int
}

type FuncDecl struct {
	*ast.FuncDecl
	Name *Ident
	Recv *FieldList // for methods
	Body *BlockStmt
	Type *FuncType
	// control flow graph
	callGraph Step
	// goto targets
	labelToStmt map[string]statementReference
	// for source access of any statement/expression within this function
	fileSet *token.FileSet
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

func (f FuncDecl) String() string {
	return fmt.Sprintf("FuncDecl(%s)", f.Name.Name)
}

type FuncType struct {
	*ast.FuncType
	TypeParams *FieldList
	Params     *FieldList
	Returns    *FieldList
}

func (t FuncType) String() string {
	return fmt.Sprintf("FuncType(%v,%v,%v)", t.TypeParams, t.Params, t.Returns)
}

func (t FuncType) Eval(vm *VM) {}

type Ellipsis struct {
	*ast.Ellipsis
	Elt Expr // ellipsis element type (parameter lists only); or nil
}

func (e Ellipsis) String() string {
	return fmt.Sprintf("Ellipsis(%v)", e.Elt)
}
func (e Ellipsis) Eval(vm *VM) {}
