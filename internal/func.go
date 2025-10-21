package internal

import (
	"fmt"
	"go/ast"
)

type FuncDecl struct {
	*ast.FuncDecl
	Name      *Ident
	Recv      *FieldList
	Body      *BlockStmt
	Type      *FuncType
	callGraph Step
}

func (f FuncDecl) Eval(vm *VM) {
	if f.Body != nil {
		vm.eval(f.Body)
	}
}

func (f FuncDecl) Flow(g *graphBuilder) (head Step) {
	head = g.current
	if f.Body != nil {
		head = f.Body.Flow(g)
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
