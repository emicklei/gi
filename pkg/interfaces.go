package pkg

import (
	"fmt"
	"go/token"
	"reflect"

	"github.com/emicklei/dot"
)

type Evaluable interface {
	Pos() token.Pos
	Eval(vm *VM)
}

type CanAssign interface {
	// =
	assign(vm *VM, value reflect.Value)
	// :=
	define(vm *VM, value reflect.Value)
}

// TODO only ValueSpec implements CanDeclare
type CanDeclare interface {
	// Declare declares the variable in the current environment.
	// It returns true if the declaration set a valid reflect Value.
	declare(vm *VM) bool
	callGraph() Step
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
	declStep() CanDeclare
}

type Step interface {
	fmt.Stringer
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
	take(vm *VM) Step
}

type Traverseable interface {
	traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node
}

// FuncDecl and FuncLit implement this
type Func interface {
	setHasRecoverCall(bool)
	hasRecoverCall() bool
	putGotoReference(label string, ref statementReference)
	gotoReference(label string) statementReference
	results() *FieldList
	params() *FieldList
}
