package internal

import (
	"fmt"
	"go/ast"
	"go/token"
)

type activeFuncDecl struct {
	FuncDecl  FuncDecl
	deferList []Expr
}

func (af *activeFuncDecl) addDefer(call Expr) {
	af.deferList = append(af.deferList, call)
}

type statementReference struct {
	step  Step
	index int
}

type FuncDecl struct {
	*ast.FuncDecl
	Name *Ident
	Recv *FieldList
	Body *BlockStmt
	Type *FuncType
	// control flow graph
	callGraph Step
	// goto targets
	labelToStmt map[string]statementReference
	// for source access of any statement/expression within this function
	fileSet *token.FileSet
}

func (f FuncDecl) Eval(vm *VM) {
	// if f.Body != nil {
	// 	af := &activeFuncDecl{FuncDecl: f, bodyListIndex: -1}
	// 	vm.activeFuncStack.push(af)
	// 	// execute statements
	// 	for af.hasNextStmt() {
	// 		stmt := af.nextStmt()
	// 		if trace {
	// 			vm.traceEval(stmt.stmtStep())
	// 		} else {
	// 			stmt.stmtStep().Eval(vm)
	// 		}
	// 	}
	// 	// run defer statements
	// 	for i := len(af.deferList) - 1; i >= 0; i-- {
	// 		deferCall := af.deferList[i]
	// 		if trace {
	// 			vm.traceEval(deferCall)
	// 		} else {
	// 			deferCall.Eval(vm)
	// 		}
	// 	}
	// 	vm.activeFuncStack.pop()
	// }
}

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
