package internal

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = CallExpr{}

type CallExpr struct {
	Lparen token.Pos // position of "("
	Fun    Expr
	Args   []Expr // function arguments; or nil
}

func (c CallExpr) Eval(vm *VM) {
	// function fn is either an external or an interpreted one
	fn := vm.callStack.top().pop() // see Flow

	switch fn.Kind() {
	case reflect.Struct:
		switch f := fn.Interface().(type) {
		// order by frequency of use
		case builtinFunc:
			c.handleBuiltinFunc(vm, f)
		case FuncDecl:
			c.handleFuncDecl(vm, f)
		case FuncLit:
			c.handleFuncLit(vm, f)
		case ArrayType:
			c.handleArrayType(vm, f)
		case TypeSpec:
			c.handleTypeSpec(vm, f)
		default:
			vm.fatal(fmt.Sprintf("expected FuncDecl,FuncLit or builtinFunc, got %T", fn.Interface()))
		}
	case reflect.Func:
		args := make([]reflect.Value, len(c.Args))
		// first to last, see Flow
		for i := range len(c.Args) {
			var argType reflect.Type
			if fn.Type().IsVariadic() {
				argType = fn.Type().In(0)
			} else {
				argType = fn.Type().In(i)
			}
			val := vm.callStack.top().pop()
			if !val.IsValid() || val == untypedNil {
				args[i] = reflect.New(argType).Elem()
				continue
			}
			if hp, ok := isHeapPointer(val); ok {
				hpv := vm.heap.read(hp)
				if hpv.CanAddr() {
					// TODO
					args[i] = hpv.Addr()
				} else {
					newPtr := reflect.New(hpv.Type())
					newPtr.Elem().Set(hpv)
					args[i] = newPtr
					// after the call use the value of newPtr to write back the heapointer backing value
					defer func() {
						vm.heap.write(hp, newPtr.Elem())
					}()
				}
			} else {
				// need conversion?
				if argType.Kind() == reflect.Interface && val.Type().Implements(argType) {
					args[i] = val
					continue
				}
				// reflect convert?
				if val.CanConvert(argType) {
					args[i] = val.Convert(argType)
					continue
				}
				args[i] = val
			}
		}
		vals := fn.Call(args)
		pushCallResults(vm, vals)
	default:
		vm.fatal(fmt.Sprintf("call to unknown function type: %v (%T)", fn.Interface(), fn.Interface()))
	}
}

func (c CallExpr) handleTypeSpec(vm *VM, ts TypeSpec) {
	// do a conversion to the specified type
	toConvert := vm.callStack.top().pop()
	rt := vm.returnsType(ts.Type)
	cv := toConvert.Convert(rt)
	vm.pushOperand(cv)
}

func (c CallExpr) handleArrayType(vm *VM, at ArrayType) {
	// do a conversion to array/slice
	toConvert := vm.callStack.top().pop()
	rt := vm.returnsType(at.Elt)
	length := toConvert.Len()
	capacity := toConvert.Len()
	st := reflect.SliceOf(rt)
	sl := reflect.MakeSlice(st, length, capacity)
	reflect.Copy(sl, toConvert)
	vm.pushOperand(sl)
}

func (c CallExpr) handleBuiltinFunc(vm *VM, bf builtinFunc) {
	switch bf.name {
	case "new":
		c.evalNew(vm)
	case "append":
		c.evalAppend(vm)
	case "make":
		c.evalMake(vm)
	case "delete":
		c.evalDelete(vm)
	case "copy":
		c.evalCopy(vm)
	case "clear":
		cleared := c.evalClear(vm)
		// the argument of clear needs to be replaced
		if identArg, ok := c.Args[0].(Ident); ok {
			vm.callStack.top().env.set(identArg.Name, cleared)
		} else {
			vm.fatal("clear argument must be an identifier")
		}
	case "min":
		c.evalMin(vm)
	case "max":
		c.evalMax(vm)
	default:
		vm.fatal("unknown builtin function: " + bf.name)
	}
}

func (c CallExpr) handleFuncLit(vm *VM, fl FuncLit) {
	// TODO deduplicate with handleFuncDecl
	// prepare arguments
	args := make([]reflect.Value, len(c.Args))
	for i := range c.Args {
		val := vm.callStack.top().pop() // first to last, see Flow
		args[i] = val
	}
	vm.pushNewFrame(fl)
	frame := vm.callStack.top()

	setParametersToFrame(fl.Type, args, vm, frame)
	setZeroReturnsToFrame(fl.Type, vm, frame)

	// we already have the call graph in FuncLit
	vm.takeAllStartingAt(fl.callGraph)

	// run defer list
	for i := len(frame.deferList) - 1; i >= 0; i-- {
		invocation := frame.deferList[i]
		// put env in place
		frame.env = invocation.env
		vm.takeAllStartingAt(invocation.flow)
	}

	// take values before popping frame
	vals := frame.returnValues
	vm.popFrame()
	pushCallResults(vm, vals)
}

func setZeroReturnsToFrame(ft *FuncType, vm *VM, frame *stackFrame) {
	if ft.Results == nil {
		return
	}
	r := 0
	for _, field := range ft.Results.List {
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
		val := vm.callStack.top().pop()
		args[i] = val
	}
	vm.pushNewFrame(fd)
	frame := vm.callStack.top()
	setParametersToFrame(fd.Type, args, vm, frame)
	setZeroReturnsToFrame(fd.Type, vm, frame)

	// when stepping we the call graph in FuncDecl
	vm.takeAllStartingAt(fd.callGraph)

	// run defer list
	for i := len(frame.deferList) - 1; i >= 0; i-- {
		invocation := frame.deferList[i]
		frame.env = invocation.env
		// put env in place
		vm.takeAllStartingAt(invocation.flow)
	}

	// take values before popping frame
	vals := frame.returnValues
	vm.popFrame()
	pushCallResults(vm, vals)
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

func (c CallExpr) Pos() token.Pos { return c.Lparen }

func (c CallExpr) String() string {
	return fmt.Sprintf("CallExpr(%v, args=%d)", c.Fun, len(c.Args))
}
