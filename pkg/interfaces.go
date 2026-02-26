package pkg

import (
	"go/token"
	"reflect"

	"github.com/emicklei/dot"
)

type Evaluable interface {
	pos() token.Pos
	eval(vm *VM)
}

type CanAssign interface {
	// =
	assign(vm *VM, value reflect.Value)
	// :=
	define(vm *VM, value reflect.Value)
}

type Expr interface {
	Flowable
	Evaluable
}

type Flowable interface {
	// flow builds the control flow graph using the provided grapher.
	// head is the entry point to that call flow graph.
	flow(g *graphBuilder) (head Step)
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Flowable
	stmtStep() Evaluable
}

type CanCompose interface {
	literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value
}

// CanSelect is implemented by types that support selection of fields or methods by name.
type CanSelect interface {
	selectFieldOrMethod(name string) reflect.Value
}

type FieldAssignable interface {
	fieldAssign(name string, val reflect.Value)
}

type CanMake interface {
	// size can be 0 if not applicable
	// elements can be nil if not applicable
	makeValue(vm *VM, size int, elements []reflect.Value) reflect.Value
	CanCompose
}

type Decl interface {
	Flowable
	Evaluable
}

type Step interface {
	Evaluable
	StepTaker
	Traverseable

	// implemented by step
	Next() Step
	SetNext(s Step)
	ID() int
	SetID(id int)
}

type StepTaker interface {
	take(vm *VM)
}

type Traverseable interface {
	traverse(g *dot.Graph, fs *token.FileSet) dot.Node
}

// FuncDecl and FuncLit implement this
type Func interface {
	setHasRecoverCall(bool)
	hasRecoverCall() bool
	putGotoReference(label string, ref stmtReference)
	gotoReference(label string) stmtReference
	results() *FieldList
	pos() token.Pos
}

// for gi internal use
type ToStringer interface {
	toString() string
}

type HasMethods interface {
	methodsMap() map[string]*FuncDecl
}
