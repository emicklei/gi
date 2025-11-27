package internal

import (
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
	Assign(vm *VM, value reflect.Value)
	// :=
	Define(vm *VM, value reflect.Value)
}

// TODO only ValueSpec implements CanDeclare
type CanDeclare interface {
	// Declare declares the variable in the current environment.
	// It returns true if the declaration set a valid reflect Value.
	Declare(vm *VM) bool
	CallGraph() Step
}

type Expr interface {
	Flowable
	Evaluable
}

type Flowable interface {
	// Flow builds the control flow graph using the provided grapher.
	// Head is the entry point to that call flow graph.
	Flow(g *graphBuilder) (head Step)
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Flowable
	stmtStep() Evaluable
}

type CanCompose interface {
	LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value
}

type FieldSelectable interface {
	Select(name string) reflect.Value
}

type FieldAssignable interface {
	Assign(name string, val reflect.Value)
}

type CanInstantiate interface {
	// size can be 0 if not applicable
	// constructorArgs can be nil if not applicable
	Instantiate(vm *VM, size int, constructorArgs []reflect.Value) reflect.Value
	CanCompose
}

type Decl interface {
	Flowable
	declStep() CanDeclare
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

	String() string
}

type StepTaker interface {
	Take(vm *VM) Step
}

type Traverseable interface {
	Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node
}
