package internal

import (
	"fmt"
	"go/ast"
)

type activeFuncDecl struct {
	FuncDecl      FuncDecl
	bodyListIndex int
}

func (af *activeFuncDecl) hasNext() bool {
	return af.bodyListIndex+1 < len(af.FuncDecl.Body.List)
}

func (af *activeFuncDecl) next() Stmt {
	af.bodyListIndex++
	return af.FuncDecl.Body.List[af.bodyListIndex]
}

func (af *activeFuncDecl) setNextIndex(index int) {
	af.bodyListIndex = index - 1
}

func (af *activeFuncDecl) setDone() {
	af.bodyListIndex = len(af.FuncDecl.Body.List)
}

type FuncDecl struct {
	*ast.FuncDecl
	Name      *Ident
	Recv      *FieldList
	Body      *BlockStmt
	Type      *FuncType
	callGraph Step
	// for labelled statements
	labelToListIndex map[string]int
}

func (f FuncDecl) Eval(vm *VM) {
	if f.Body != nil {
		af := &activeFuncDecl{FuncDecl: f, bodyListIndex: -1}
		vm.funcStack.push(af)
		for af.hasNext() {
			stmt := af.next()
			if trace {
				vm.traceEval(stmt.stmtStep())
			} else {
				stmt.stmtStep().Eval(vm)
			}
		}
		vm.funcStack.pop()
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
