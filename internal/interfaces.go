package internal

import (
	"reflect"

	"github.com/emicklei/dot"
)

type Evaluable interface {
	Eval(vm *VM)
}

type CanAssign interface {
	// =
	Assign(vm *VM, value reflect.Value)
	// :=
	Define(vm *VM, value reflect.Value)
}

type CanDeclare interface {
	// Declare declares the variable in the current environment.
	// It returns true if the declaration set a valid reflect Value.
	Declare(vm *VM) bool
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

type HasZeroValue interface {
	ZeroValue(env Env) reflect.Value
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

type CanInstantiate interface {
	Instantiate(vm *VM) reflect.Value // , typeArgs []reflect.Type) reflect.Value
	LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value
}

type Decl interface {
	Flowable
	declStep() CanDeclare
}

type Step interface {
	Evaluable
	StepTaker
	Traverseable
	SetNext(s Step)
	Next() Step
	ID() int
	String() string
}

type StepTaker interface {
	Take(vm *VM) Step
}

type Traverseable interface {
	Traverse(g *dot.Graph, visited map[int]dot.Node) dot.Node
}
