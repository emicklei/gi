package internal

import (
	"fmt"
	"go/ast"
)

type activeFuncDecl struct {
	FuncDecl      FuncDecl
	bodyListIndex int
	deferList     []Expr
}

func (af *activeFuncDecl) hasNextStmt() bool {
	return af.bodyListIndex+1 < len(af.FuncDecl.Body.List)
}

func (af *activeFuncDecl) nextStmt() Stmt {
	af.bodyListIndex++
	return af.FuncDecl.Body.List[af.bodyListIndex]
}

func (af *activeFuncDecl) setNextStmtIndex(index int) {
	af.bodyListIndex = index - 1
}

func (af *activeFuncDecl) setDone() {
	af.bodyListIndex = len(af.FuncDecl.Body.List)
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
	Name        *Ident
	Recv        *FieldList
	Body        *BlockStmt
	Type        *FuncType
	callGraph   Step
	labelToStmt map[string]statementReference
}

func (f FuncDecl) Eval(vm *VM) {
	if f.Body != nil {
		af := &activeFuncDecl{FuncDecl: f, bodyListIndex: -1}
		vm.activeFuncStack.push(af)
		// execute statements
		for af.hasNextStmt() {
			stmt := af.nextStmt()
			if trace {
				vm.traceEval(stmt.stmtStep())
			} else {
				stmt.stmtStep().Eval(vm)
			}
		}
		// run defer statements
		for i := len(af.deferList) - 1; i >= 0; i-- {
			deferCall := af.deferList[i]
			if trace {
				vm.traceEval(deferCall)
			} else {
				deferCall.Eval(vm)
			}
		}
		vm.activeFuncStack.pop()
	}
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
