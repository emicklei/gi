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

func (c CallExpr) Eval(vm *VM) {
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
			vm.fatalf("struct unexpected %T", fn.Interface())
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
				hpv := vm.heap.read(hp)
				if hpv.CanAddr() {
					// TODO
					args[i] = hpv.Addr()
				} else if sv, ok := hpv.Interface().(StructValue); ok {
					args[i] = reflect.ValueOf(&sv)
				} else {
					// TODO needed?
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
		pushCallResults(vm, vals)
	default:
		vm.fatal(fmt.Sprintf("call to unknown function type: %v (%T)", fn.Interface(), fn.Interface()))
	}
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
		pushCallResults(vm, []reflect.Value{blt.prtZeroValue})
		return
	}
	vals := blt.convertFunc.Call([]reflect.Value{arg})
	pushCallResults(vm, vals)
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
	pushCallResults(vm, vals)
}

func (c CallExpr) handleArrayType(vm *VM, at ArrayType) {
	// do a conversion to array/slice
	toConvert := vm.popOperand()
	rt := vm.makeType(at.elt)
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
			vm.localEnv().set(identArg.name, cleared)
		} else {
			vm.fatal("clear argument must be an identifier")
		}
	case "min":
		c.evalMin(vm)
	case "max":
		c.evalMax(vm)
	case "recover":
		c.evalRecover(vm)
	default:
		vm.fatal("unknown builtin function: " + bf.name)
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

	setParametersToFrame(fl.Type, args, vm, frame)
	setZeroReturnsToFrame(fl.Type, vm, frame)

	if fl.hasRecoverCall() {
		defer func() {
			r := recover()
			// temporary store it in the special variable in the parent env
			frame.env.getParent().set(internalVarName("recover", 0), reflect.ValueOf(r))
			frame.takeDeferList(vm)
		}()
	}

	// we already have the call graph in FuncLit
	vm.takeAllStartingAt(fl.callGraph)

	frame.takeDeferList(vm)

	// take values before popping frame
	vals := []reflect.Value{} // todo size it
	if fl.Type.Results != nil {
		for _, field := range fl.results().List {
			for _, name := range field.names {
				val := frame.env.valueLookUp(name.name)
				vals = append(vals, val)
			}
		}
	}
	vm.popFrame()
	pushCallResults(vm, vals)
}

func setZeroReturnsToFrame(ft *FuncType, vm *VM, frame *stackFrame) {
	if ft.Results == nil {
		return
	}
	for _, field := range ft.Results.List {
		for _, name := range field.names {
			val := reflect.Zero(vm.makeType(field.typ)) // TODO put types from gopkg in Field?
			frame.env.set(name.name, val)
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
		for _, name := range field.names {
			val := args[p]
			if val.Interface() == untypedNil {
				// create a zero value of the expected type
				val = reflect.Zero(vm.makeType(field.typ)) // TODO put types from gopkg in Field?
			}
			frame.env.set(name.name, val)
			p++
		}
	}
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
			elemType := vm.makeType(expectedType)
			sliceType := reflect.SliceOf(elemType)
			sliceVal := reflect.MakeSlice(sliceType, len(vals), len(vals))
			for k := 0; k < len(vals); k++ {
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
			frame.env.set(recvName, receiver)
		} else {
			// put a copy of the value
			if sv, ok := receiver.Interface().(StructValue); ok {
				clone := sv.clone()
				frame.env.set(recvName, reflect.ValueOf(clone))
			}
			if ev, ok := receiver.Interface().(ExtendedValue); ok {
				// no need to clone value
				frame.env.set(recvName, reflect.ValueOf(ev))
			}
		}
	}

	setParametersToFrame(fd.typ, args, vm, frame)
	setZeroReturnsToFrame(fd.typ, vm, frame)

	if fd.hasRecoverCall() {
		defer func() {
			if r := recover(); r != nil {
				// temporary store it in the special variable in the parent env
				frame.env.getParent().set(internalVarName("recover", 0), reflect.ValueOf(r))
				frame.takeDeferList(vm)
			}
		}()
	}

	// take all steps from the call graph in FuncDecl
	vm.takeAllStartingAt(fd.graph)

	frame.takeDeferList(vm)

	// take values before popping frame
	vals := []reflect.Value{} // todo size it
	if fd.typ.Results != nil {
		for _, field := range fd.results().List {
			for _, name := range field.names {
				val := frame.env.valueLookUp(name.name)
				vals = append(vals, val)
			}
		}
	}

	vm.popFrame()
	pushCallResults(vm, vals)
}

func isEllipsis(t Expr) bool {
	_, ok := t.(Ellipsis)
	return ok
}
func isStructValue(v reflect.Value) bool {
	_, ok := v.Interface().(StructValue)
	return ok
}

func isPointerExpr(e Expr) bool {
	_, ok := e.(StarExpr)
	return ok
}

func (c CallExpr) flow(g *graphBuilder) (head Step) {
	// make sure first value is on top of the operand stack
	// so we can pop in the right order during Eval
	for i := len(c.args) - 1; i >= 0; i-- {
		if i == len(c.args)-1 {
			head = c.args[i].flow(g)
			continue
		}
		c.args[i].flow(g)
	}
	funFlow := c.fun.flow(g)
	if head == nil { // must be a function with no args
		head = funFlow
	}
	g.next(c)
	return head
}

func (c CallExpr) deferFlow(g *graphBuilder) (head Step) {
	head = c.fun.flow(g)
	g.next(c)
	return
}

func (c CallExpr) Pos() token.Pos { return c.lparenPos }

func (c CallExpr) String() string {
	return fmt.Sprintf("CallExpr(%v, args=%d)", c.fun, len(c.args))
}
