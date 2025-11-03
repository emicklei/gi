package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Expr = CallExpr{}

type CallExpr struct {
	*ast.CallExpr
	Fun  Expr
	Args []Expr
}

func (c CallExpr) Eval(vm *VM) {
	// function fn is either an external or an interpreted one
	fn := vm.frameStack.top().pop() // see Flow

	// TODO
	if !fn.IsValid() {
		vm.fatal("call to invalid function:" + fmt.Sprintf("%v", c.Fun))
	}

	switch fn.Kind() {
	case reflect.Func:
		args := make([]reflect.Value, len(c.Args))
		for i, arg := range c.Args {
			val := vm.frameStack.top().pop() // first to last, see Flow
			if !val.IsValid() {
				vm.fatal(fmt.Sprintf("call to function: %v with invalid argument %d: %v", c.Fun, i, arg))
			}
			args[i] = val
		}
		vals := fn.Call(args)
		vm.pushCallResults(vals)

	case reflect.Struct:
		fl, ok := fn.Interface().(FuncLit)
		if ok {
			c.handleFuncLit(vm, fl)
			return
		}
		lf, ok := fn.Interface().(FuncDecl)
		if ok {
			c.handleFuncDecl(vm, lf)
			return
		}
		bf, ok := fn.Interface().(builtinFunc)
		if ok {
			c.handleBuiltinFunc(vm, bf)
			return
		}
		vm.fatal(fmt.Sprintf("expected FuncDecl,FuncLit or builtinFunc, got %T", fn.Interface()))
	default:
		vm.fatal(fmt.Sprintf("call to unknown function type: %v (%T)", fn.Interface(), fn.Interface()))
	}
}

func (c CallExpr) handleBuiltinFunc(vm *VM, bf builtinFunc) {
	switch bf.name {
	case "delete":
		c.evalDelete(vm)
	case "append":
		c.evalAppend(vm)
	case "clear":
		cleared := c.evalClear(vm)
		// the argument of clear needs to be replaced
		if identArg, ok := c.Args[0].(Ident); ok {
			vm.frameStack.top().env.set(identArg.Name, cleared)
		} else {
			vm.fatal("clear argument must be an identifier")
		}
	case "min":
		c.evalMin(vm)
	case "max":
		c.evalMax(vm)
	case "make":
		c.evalMake(vm)
	case "new":
		c.evalNew(vm)
	default:
		vm.fatal("unknown builtin function: " + bf.name)
	}
}

func (c CallExpr) handleFuncLit(vm *VM, fl FuncLit) {
	// TODO deduplicate with handleFuncDecl
	// prepare arguments
	args := make([]reflect.Value, len(c.Args))
	for i := range c.Args {
		val := vm.frameStack.top().pop() // first to last, see Flow
		args[i] = val
	}
	vm.pushNewFrame(c)
	frame := vm.frameStack.top()

	setParametersToFrame(fl.Type, args, vm, frame)
	setZeroReturnsToFrame(fl.Type, vm, frame)

	// we already have the call graph in FuncLit
	vm.takeAllStartingAt(fl.callGraph)

	// take values before popping frame
	vals := vm.frameStack.top().returnValues
	vm.popFrame()
	vm.pushCallResults(vals)
}

func setZeroReturnsToFrame(ft *FuncType, vm *VM, frame *stackFrame) {
	if ft.Returns == nil {
		return
	}
	r := 0
	for _, field := range ft.Returns.List {
		for _, name := range field.Names {
			val := reflect.Zero(vm.returnsType(field.Type)) // TODO use gopkg?
			frame.env.set(name.Name, val)
			r++
		}
	}
}

// setParametersToFrame takes all parameters and put them in the env of the new frame
func setParametersToFrame(ft *FuncType, args []reflect.Value, vm *VM, frame *stackFrame) {
	if ft.Params == nil {
		return
	}
	p := 0
	for _, field := range ft.Params.List {
		for _, name := range field.Names {
			val := args[p]
			if val.Interface() == untypedNil {
				// create a zero value of the expected type
				val = reflect.Zero(vm.returnsType(field.Type)) // TODO use gopkg?
			}
			frame.env.set(name.Name, val)
			p++
		}
	}
}

func (c CallExpr) handleFuncDecl(vm *VM, fd FuncDecl) {
	// TODO deduplicate with handleFuncLit
	// prepare arguments
	args := make([]reflect.Value, len(c.Args))
	// first to last, see Flow
	for i := range c.Args {
		val := vm.frameStack.top().pop()
		args[i] = val
	}
	vm.pushNewFrame(c)
	frame := vm.frameStack.top()
	setParametersToFrame(fd.Type, args, vm, frame)
	setZeroReturnsToFrame(fd.Type, vm, frame)

	// when stepping we the call graph in FuncDecl
	vm.takeAllStartingAt(fd.callGraph)

	// take values before popping frame
	vals := frame.returnValues
	vm.popFrame()
	vm.pushCallResults(vals)
}

func (c CallExpr) Flow(g *graphBuilder) (head Step) {
	// make sure first value is on top of the operand stack
	// so we can pop in the right order during Eval
	for i := len(c.Args) - 1; i >= 0; i-- {
		if i == len(c.Args)-1 {
			head = c.Args[i].Flow(g)
			continue
		}
		c.Args[i].Flow(g)
	}
	funFlow := c.Fun.Flow(g)
	if head == nil { // could be only a function with no args
		head = funFlow
	}
	g.next(c)
	return head
}

func (c CallExpr) String() string {
	return fmt.Sprintf("CallExpr(%v, len=%d)", c.Fun, len(c.Args))
}
