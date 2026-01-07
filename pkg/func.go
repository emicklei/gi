package pkg

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

type funcDeclPair struct {
	fn       Func
	bodyList []ast.Stmt // need this to find index of each block stmt
}

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
	graph Step
	// goto targets
	labelToStmt map[string]statementReference
	// for source access of any statement/expression within this function
	fileSet      *token.FileSet
	callsRecover bool
}

func (f *FuncDecl) Eval(vm *VM) {} // noop

func (f *FuncDecl) flow(g *graphBuilder) (head Step) {
	head = g.current
	if f.Body != nil {
		g.funcStack.push(f)
		head = f.Body.flow(g)
		g.funcStack.pop()
	}
	return
}

func (f *FuncDecl) setHasRecoverCall(bool) { f.callsRecover = true }
func (f *FuncDecl) hasRecoverCall() bool   { return f.callsRecover }
func (f *FuncDecl) putGotoReference(label string, ref statementReference) {
	if f.labelToStmt == nil {
		f.labelToStmt = make(map[string]statementReference)
	}
	f.labelToStmt[label] = ref
}
func (f *FuncDecl) gotoReference(label string) statementReference {
	return f.labelToStmt[label]
}
func (f *FuncDecl) results() *FieldList {
	return f.Type.Results
}
func (f *FuncDecl) params() *FieldList {
	return f.Type.Params
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

func (t FuncType) flow(g *graphBuilder) (head Step) {
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

func (e Ellipsis) flow(g *graphBuilder) (head Step) {
	if e.Elt != nil {
		head = e.Elt.flow(g)
	} else {
		g.next(e)
		return g.current
	}
	return
}

// funcInvocation represents a function call instance with its own environment.
// this used to handle defer statements properly.
type funcInvocation struct {
	flow      Step
	env       Env
	arguments []reflect.Value
}

func isRecoverCall(expr Expr) bool {
	if ident, ok := expr.(Ident); ok {
		return ident.Name == "recover"
	}
	return false
}

var _ Expr = &FuncLit{}

type FuncLit struct {
	Type      *FuncType
	Body      *BlockStmt // TODO not sure what to do when Body and/or Type is nil
	callGraph Step
	// goto targets
	labelToStmt  map[string]statementReference // TODO lazy initialization
	callsRecover bool
}

func (f *FuncLit) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(f))
}

func (f *FuncLit) flow(g *graphBuilder) (head Step) {
	g.next(f)
	return g.current
}

func (f *FuncLit) Pos() token.Pos { return f.Type.Pos() }

func (f *FuncLit) setHasRecoverCall(bool) { f.callsRecover = true }
func (f *FuncLit) hasRecoverCall() bool   { return f.callsRecover }
func (f *FuncLit) putGotoReference(label string, ref statementReference) {
	if f.labelToStmt == nil {
		f.labelToStmt = make(map[string]statementReference)
	}
	f.labelToStmt[label] = ref
}
func (f *FuncLit) gotoReference(label string) statementReference {
	return f.labelToStmt[label]
}
func (f *FuncLit) results() *FieldList {
	return f.Type.Results
}
func (f *FuncLit) params() *FieldList {
	return f.Type.Params
}
func (f *FuncLit) String() string {
	return fmt.Sprintf("FuncLit(%v,%v)", f.Type, f.Body)
}
