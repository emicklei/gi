package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

var _ Flowable = CallExpr{}
var _ Expr = CallExpr{}

type CallExpr struct {
	*ast.CallExpr
	Fun  Expr
	Args []Expr
}

func (c CallExpr) Eval(vm *VM) {
	// function fn is either an external or an interpreted one
	var fn reflect.Value
	if vm.isStepping {
		fn = vm.callStack.top().pop() // see Flow
	} else {
		fn = vm.returnsEval(c.Fun)
	}

	// TODO
	if !fn.IsValid() {
		vm.fatal("call to invalid function:" + fmt.Sprintf("%v", c.Fun))
	}

	switch fn.Kind() {
	case reflect.Func:
		args := make([]reflect.Value, len(c.Args))
		for i, arg := range c.Args {
			var val reflect.Value
			if vm.isStepping {
				val = vm.callStack.top().pop() // first to last, see Flow
			} else {
				val = vm.returnsEval(arg)
			}
			if !val.IsValid() {
				vm.fatal("call to function with invalid argument:" + fmt.Sprintf("%d=%v", i, arg))
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
		vm.fatal("expected FuncDecl or FuncLit, got " + fmt.Sprintf("%T", fn.Interface()))
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
			vm.callStack.top().env.set(identArg.Name, cleared)
		}
	case "min":
		c.evalMin(vm)
	case "max":
		c.evalMax(vm)
	case "make":
		c.evalMake(vm)
	default:
		vm.fatal("unknown builtin function: " + bf.name)
	}
}

func (c CallExpr) handleFuncLit(vm *VM, fl FuncLit) {
	// TODO deduplicate with handleFuncDecl
	// prepare arguments
	args := make([]reflect.Value, len(c.Args))
	for i, arg := range c.Args {
		var val reflect.Value
		if vm.isStepping {
			val = vm.callStack.top().pop() // first to last, see Flow
		} else {
			val = vm.returnsEval(arg)
		}
		args[i] = val
	}
	frame := vm.pushNewFrame()
	// take all parameters and put them in the env of the new frame
	p := 0
	for _, field := range fl.Type.Params.List {
		for _, name := range field.Names {
			frame.env.set(name.Name, args[p])
			p++
		}
	}
	if vm.isStepping {
		// when stepping we already have the call graph in FuncLit
		vm.takeAll(fl.callGraph)
	} else {
		vm.eval(fl.Body)
	}
	top := vm.popFrame()
	vm.pushCallResults(top.returnValues)
}
func (c CallExpr) handleFuncDecl(vm *VM, fd FuncDecl) {
	// TODO deduplicate with handleFuncLit
	// prepare arguments
	args := make([]reflect.Value, len(c.Args))
	for i, arg := range c.Args {
		var val reflect.Value
		if vm.isStepping {
			val = vm.callStack.top().pop() // first to last, see Flow
		} else {
			val = vm.returnsEval(arg)
		}
		args[i] = val
	}
	frame := vm.pushNewFrame()
	// take all parameters and put them in the env of the new frame
	p := 0
	for _, field := range fd.Type.Params.List {
		for _, name := range field.Names {
			frame.env.set(name.Name, args[p])
			p++
		}
	}
	if vm.isStepping {
		// when stepping we already have the call graph in FuncDecl
		vm.takeAll(fd.callGraph)
	} else {
		vm.eval(fd.Body)
	}
	top := vm.popFrame()
	vm.pushCallResults(top.returnValues)
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
	c.Fun.Flow(g)
	if head == nil { // could be only a function with no args
		head = g.current
	}
	g.next(c)
	return head
}

func (c CallExpr) String() string {
	return fmt.Sprintf("CallExpr(%v, len=%d)", c.Fun, len(c.Args))
}
