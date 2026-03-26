package pkg

import (
	"fmt"
	"go/token"
	"reflect"
)

var _ Expr = CallExpr{}

type CallExpr struct {
	lparenPos token.Pos // position of "("
	fun       Expr
	args      []Expr // function arguments; or nil
}

func (c CallExpr) eval(vm *VM) {
	// function fn is either a compiled or an interpreted one
	fn := vm.popOperand() // see Flow

	switch fn.Kind() {
	case reflect.Struct:
		switch f := fn.Interface().(type) {
		// order by frequency of use
		case builtinFunc:
			c.handleBuiltinFunc(vm, f)
		case builtinType:
			c.handleBuiltinType(vm, f)
		case ArrayType:
			c.handleArrayType(vm, f)
		case reflect.Method:
			c.handleReflectMethod(vm, f)
		case ExtendedType:
			c.handleExtendedType(vm, f)
		default:
			vm.fatalf("struct unexpected %s (%T)", stringOf(fn.Interface()), fn.Interface())
		}
	case reflect.Pointer:
		switch f := fn.Interface().(type) {
		case *FuncDecl:
			c.handleFuncDecl(vm, f)
		case *FuncLit:
			c.handleFuncLit(vm, f)
		default:
			vm.fatalf("pointer unexpected %T", fn.Interface())
		}
	case reflect.Func:
		c.handleFunc(vm, fn)
	default:
		vm.fatalf("struct unexpected %s (%T)", stringOf(fn.Interface()), fn.Interface())
	}
}

func (c CallExpr) handleFunc(vm *VM, fn reflect.Value) {
	args := make([]reflect.Value, len(c.args))
	// first to last, see Flow
	for i := range len(c.args) {
		var argType reflect.Type
		if fn.Type().IsVariadic() && i >= fn.Type().NumIn()-1 {
			// last arg is variadic
			argType = fn.Type().In(fn.Type().NumIn() - 1)
		} else {
			argType = fn.Type().In(i)
		}
		val := vm.popOperand()
		// TODO needed?
		if !val.IsValid() || val == untypedNil {
			args[i] = reflect.New(argType).Elem()
			continue
		}
		if hp, ok := asHeapPointer(val); ok {
			val := vm.heap.read(hp)
			if val.CanAddr() {
				// TODO
				args[i] = val.Addr()
			} else if sv, ok := val.Interface().(StructValue); ok {
				args[i] = reflect.ValueOf(&sv)
			} else {
				// TODO needed?
				// why not store newPtr in HeapPointer
				//newPtr := reflect.New(val.Type())
				//newPtr.Elem().Set(val)

				// tryout TODO
				newPtr := hp.ptrValue

				args[i] = newPtr
				// after the call use the value of newPtr to write back the heapointer backing value
				defer func() {
					// tryout TODO
					// hp.ptrValue.Elem().Set ...

					vm.heap.write(hp, newPtr.Elem())
				}()
			}
		} else {
			// need conversion?
			if argType.Kind() == reflect.Interface && val.Type().Implements(argType) {
				args[i] = val
				continue
			}
			// TestExtendedString
			if val.Type() == reflectExtendedType {
				etv := val.Interface().(ExtendedValue)
				// TODO for now pass the underlying value of ExtendedValue
				val = etv.val
			}
			// reflect convert?
			if val.IsValid() {
				if val.CanConvert(argType) {
					args[i] = val.Convert(argType)
					continue
				}
			} else {
				val = reflect.New(argType)
			}
			args[i] = val
		}
	}
	vals := fn.Call(args)
	vm.pushOperands(vals...)
}

func (c CallExpr) handleExtendedType(vm *VM, et ExtendedType) {
	arg := vm.popOperand()
	val := ExtendedValue{
		typ: et,
		val: arg,
	}
	vm.pushOperand(reflect.ValueOf(val))
}

func (c CallExpr) handleBuiltinType(vm *VM, blt builtinType) {
	arg := vm.popOperand()
	if arg == reflectNil {
		vm.pushOperands(blt.prtZeroValue)
		return
	}
	vals := blt.convertFunc.Call([]reflect.Value{arg})
	vm.pushOperands(vals...)
}

// pre: not a method from an interface type
func (c CallExpr) handleReflectMethod(vm *VM, rm reflect.Method) {
	// Get the receiver (the value the method is called on)
	receiver := vm.popOperand()

	// Prepare arguments: receiver + method args
	args := make([]reflect.Value, len(c.args)+1)
	args[0] = receiver // First arg is always the receiver

	for i := range c.args {
		val := vm.popOperand() // first to last, see Flow
		args[i+1] = val
	}

	// Call the method using rm.Func
	vals := rm.Func.Call(args)
	vm.pushOperands(vals...)
}

func (c CallExpr) handleArrayType(vm *VM, at ArrayType) {
	// do a conversion to array/slice
	toConvert := vm.popOperand()
	rt := makeType(vm, at.elt)
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
		if identArg, ok := c.args[0].(Ident); ok {
			vm.currentEnv().valueSet(identArg.name, cleared)
		} else {
			vm.fatalf("clear argument must be an identifier")
		}
	case "min":
		c.evalMin(vm)
	case "max":
		c.evalMax(vm)
	case "recover":
		c.evalRecover(vm)
	default:
		vm.fatalf("unknown builtin function: %s", bf.name)
	}
}

func (c CallExpr) handleFuncLit(vm *VM, fl *FuncLit) {
	// TODO deduplicate with handleFuncDecl
	// prepare arguments
	args := make([]reflect.Value, len(c.args))
	for i := range c.args {
		val := vm.popOperand() // first to last, see Flow
		args[i] = val
	}
	vm.pushNewFrame(fl)
	frame := vm.currentFrame

	setParametersForFrame(fl.Type, args, vm, frame)
	setZeroReturnsForFrame(fl.Type, vm, frame)

	vm.currentFrame.step = fl.callGraph
}

// TODO deduplicate with handleFuncLit
func (c CallExpr) handleFuncDecl(vm *VM, fd *FuncDecl) {
	// if method then take receiver from the stack
	var receiver reflect.Value
	if fd.recv != nil {
		receiver = vm.popOperand()
		// need to wait for a frame to set the receiver in env
	}

	// prepare arguments
	args := make([]reflect.Value, len(c.args))
	// first to last, see Flow
	for i := range c.args {
		val := vm.popOperand()
		expectedType := fieldTypeExpr(fd.typ.Params, i)
		if isEllipsis(expectedType) {
			// consume remaining as slice
			vals := make([]reflect.Value, len(c.args)-i)
			vals[0] = val
			for j := 1; j < len(vals); j++ {
				vals[j] = vm.popOperand()
			}
			elemType := makeType(vm, expectedType)
			sliceType := reflect.SliceOf(elemType)
			sliceVal := reflect.MakeSlice(sliceType, len(vals), len(vals))
			for k := range len(vals) {
				// TODO can also require dereferencing
				sliceVal.Index(k).Set(vals[k])
			}
			args[i] = sliceVal
			break
		} else {
			// check for value/pointer mismatch
			if val.Kind() == reflect.Pointer {
				if isStructValue(val) {
					if !isPointerExpr(expectedType) {
						// need to dereference
						clone := val.Interface().(StructValue).clone()
						val = reflect.ValueOf(clone)
					}
				}
			}
			if st, ok := val.Interface().(StructValue); ok {
				// need to clone to have copy semantics
				val = reflect.ValueOf(st.clone())
			}
			args[i] = val
		}
	}
	vm.pushNewFrame(fd)
	frame := vm.currentFrame

	// if method, set receiver in env
	if fd.recv != nil {
		recvName := fd.recv.List[0].names[0].name
		// check pointer receiver
		if isPointerExpr(fd.recv.List[0].typ) {
			frame.env.valueSet(recvName, receiver)
		} else {
			// put a copy of the value
			if sv, ok := receiver.Interface().(StructValue); ok {
				clone := sv.clone()
				frame.env.valueSet(recvName, reflect.ValueOf(clone))
			}
			if ev, ok := receiver.Interface().(ExtendedValue); ok {
				// no need to clone value
				frame.env.valueSet(recvName, reflect.ValueOf(ev))
			}
		}
	}

	setParametersForFrame(fd.typ, args, vm, frame)
	setZeroReturnsForFrame(fd.typ, vm, frame)

	vm.currentFrame.step = fd.callGraph
}

func setZeroReturnsForFrame(ft *FuncType, vm *VM, frame *stackFrame) {
	if ft.Results == nil {
		return
	}
	for _, field := range ft.Results.List {
		for _, name := range field.names {
			val := reflect.Zero(makeType(vm, field.typ)) // TODO put types from gopkg in Field?
			frame.env.valueSet(name.name, val)
		}
	}
}

// setParametersForFrame takes all parameters and put them in the env of the new frame
func setParametersForFrame(ft *FuncType, args []reflect.Value, vm *VM, frame *stackFrame) {
	if ft.Params == nil {
		return
	}
	p := 0
	for _, field := range ft.Params.List {
		for _, name := range field.names {
			val := args[p]
			if val.Interface() == untypedNil {
				// create a zero value of the expected type
				val = reflect.Zero(makeType(vm, field.typ)) // TODO put types from gopkg in Field?
			}
			frame.env.valueSet(name.name, val)
			p++
		}
	}
}

func (c CallExpr) flow(g *graphBuilder) (head Step) {
	// make sure first value is on top of the operand stack
	// so we can pop in the right order during Eval
	for i := len(c.args) - 1; i >= 0; i-- {
		argFlow := c.args[i].flow(g)
		if i == len(c.args)-1 {
			head = argFlow
		}
	}
	funFlow := c.fun.flow(g)
	if head == nil { // must be a function with no args
		head = funFlow
	}
	g.next(c)
	return head
}

// DeferCallExpr is a wrapper around CallExpr to distinguish defer calls from regular calls in the flow graph.
// For defer calls, the arguments are evaluated at the time of the defer statement creation.
type DeferCallExpr struct {
	CallExpr
}

func (c DeferCallExpr) flow(g *graphBuilder) (head Step) {
	head = c.fun.flow(g)
	g.next(c)
	return
}

func (c DeferCallExpr) String() string {
	return fmt.Sprintf("DeferCallExpr(%v, args=%d)", c.fun, len(c.args))
}

// Runs defers and pushes return values on the operand stack after a function call.
// Its pops the current frame and pushes the return values on the operand stack so they can be used by the caller.
// This is called for interpreted functions (FuncDecl,FuncLit) only and right after the return.
func postCallFunc(vm *VM) {
	frame := vm.currentFrame
	hasDefers := len(frame.defers) > 0
	hasResults := frame.callee != nil && frame.callee.results() != nil && len(frame.callee.results().List) != 0

	// check for quick return
	if !hasDefers && !hasResults {
		vm.popFrame()
		return
	}

	b := newGraphBuilder(vm.pkg.Package)
	var head Step
	if hasDefers {
		block := BlockStmt{}
		// reverse order
		for i := len(frame.defers) - 1; i >= 0; i-- {
			invocation := frame.defers[i]
			push := &pushArgumentsStmt{args: invocation.arguments, env: invocation.env} // Not sure if env is needed here TODO
			stmt := ExprStmt{x: DeferCallExpr{CallExpr: invocation.call.(CallExpr)}}
			block.list = append(block.list, push, stmt)
		}
		head = block.flow(b)
	}
	if hasResults {
		// add step to push results on parent frame operands
		popFrameAndPushResultsStep := newFuncStep(token.NoPos, "~pop frame and push results", popFrameAndPushResults)
		if head == nil {
			head = popFrameAndPushResultsStep
		}
		b.nextStep(popFrameAndPushResultsStep)
	}
	vm.currentFrame.step = head
}

// pre: frame.callee.results() != nil
func popFrameAndPushResults(vm *VM) {
	frame := vm.currentFrame
	// take values before popping frame
	vals := []reflect.Value{}
	for _, field := range frame.callee.results().List {
		for _, name := range field.names {
			val := frame.env.valueLookUp(name.name)
			vals = append(vals, val)
		}
	}
	vm.popFrame()
	vm.pushOperands(vals...)
}

func (c CallExpr) deferFlow(g *graphBuilder) (head Step) {
	head = c.fun.flow(g)
	g.next(c)
	return
}

func (c CallExpr) pos() token.Pos { return c.lparenPos }

func (c CallExpr) String() string {
	return fmt.Sprintf("CallExpr(%v, args=%d)", c.fun, len(c.args))
}
