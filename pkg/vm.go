package pkg

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

// framePool is a pool of stackFrame values for reuse.
var framePool = sync.Pool{
	New: func() any {
		return &stackFrame{
			operands: make([]reflect.Value, 0, 8),
		}
	},
}

// stackFrame represents a single frame in the VM's function call stack.
type stackFrame struct {
	creator  Func // typically a *FuncDecl or *FuncLit
	env      Env  // current environment with name->value mapping
	operands []reflect.Value
	defers   []funcInvocation
}

// push adds a value onto the operand stack.
func (f *stackFrame) push(v reflect.Value) {
	f.operands = append(f.operands, v)
}

// pop removes and returns the top value from the operand list.
func (f *stackFrame) pop() reflect.Value {
	v := f.operands[len(f.operands)-1]
	f.operands = f.operands[:len(f.operands)-1]
	return v
}

// pushEnv creates and activates a new child environment for the stack frame.
func (f *stackFrame) pushEnv() {
	child := envPool.Get().(*Environment)
	child.parent = f.env // can be nil
	f.env = child
}

// popEnv reverts to the parent environment for the stack frame.
func (f *stackFrame) popEnv() {
	child := f.env.(*Environment)
	f.env = child.getParent() // can become nil
	// return child to pool
	child.parent = nil
	clear(child.valueTable)
	envPool.Put(child)
}

func (f *stackFrame) takeDeferList(vm *VM) {
	for i := len(f.defers) - 1; i >= 0; i-- {
		invocation := f.defers[i]
		// push all argument values as operands on the stack
		// make sure first value is on top of the operand stack
		for i := len(invocation.arguments) - 1; i >= 0; i-- {
			vm.pushOperand(invocation.arguments[i])
		}
		vm.takeAllStartingAt(invocation.flow)
	}
}

func (f *stackFrame) String() string {
	if f == nil {
		return "stackFrame(<nil>)"
	}
	buf := strings.Builder{}
	if f.creator != nil {
		fmt.Fprintf(&buf, "%v ", f.creator)
	} else {
		fmt.Fprintf(&buf, "? ")
	}
	fmt.Fprintf(&buf, "%v ", f.env)
	fmt.Fprintf(&buf, "ops=%v ", f.operands)
	return buf.String()
}

// Runtime represents a virtual machine that can execute Go code.
type VM struct {
	callStack    stack[*stackFrame]
	currentFrame *stackFrame // optimization
	heap         *Heap
	output       *bytes.Buffer  // for testing only
	fileSet      *token.FileSet // optional file set for position info
}

func NewVM(env Env) *VM {
	if os.Getenv("GI_IGNORE_EXIT") != "" {
		OnOsExit(func(code int) {
			fmt.Fprintf(os.Stderr, "[gi] os.Exit called with code %d\n", code)
		})
	}
	if os.Getenv("GI_IGNORE_PANIC") != "" {
		OnPanic(func(why any) {
			fmt.Fprintf(os.Stderr, "[gi] panic called with %v\n", why)
		})
	}
	vm := &VM{
		output:    new(bytes.Buffer),
		callStack: make(stack[*stackFrame], 0, 16),
		heap:      newHeap()}
	frame := framePool.Get().(*stackFrame)
	frame.env = env
	vm.callStack.push(frame)
	vm.currentFrame = frame
	return vm
}

// OnPanic sets the function to be called when panic is invoked in the interpreted code.
// The Go SDK panic is not called.
func OnPanic(f func(any)) {
	// TODO make this thread-safe
	builtinsMap["panic"] = reflect.ValueOf(f)
}

// OnOsExit sets the function to be called when os.Exit is invoked in the interpreted code.
// The Go SDK os.Exit is not called.
func OnOsExit(f func(int)) {
	// TODO make this thread-safe
	stdfuncs["os"]["Exit"] = reflect.ValueOf(f)
}

func (vm *VM) setFileSet(fs *token.FileSet) {
	vm.fileSet = fs
}

// localEnv returns the current environment from the top stack frame.
func (vm *VM) localEnv() Env {
	return vm.currentFrame.env
}

// returnsEval evaluates the argument and returns the popped value that was pushed onto the operand stack.
func (vm *VM) returnsEval(e Evaluable) reflect.Value {
	vm.eval(e)
	return vm.popOperand()
}

func (vm *VM) proxyType(e Expr) CanMake {
	if id, ok := e.(Ident); ok {
		typ, ok := builtins[id.Name]
		if ok {
			gt := SDKType{typ: typ.Interface().(builtinType).typ}
			return gt
		}
		typ = vm.localEnv().valueLookUp(id.Name)
		// interpreted?
		if cm, ok := typ.Interface().(CanMake); ok {
			return cm
		}
		vm.fatal(fmt.Sprintf("unhandled proxyType for Ident %v (%T)", e, e))
	}

	if sel, ok := e.(SelectorExpr); ok {
		typ := vm.localEnv().valueLookUp(sel.X.(Ident).Name)
		val := typ.Interface()
		if canSelect, ok := val.(CanSelect); ok {
			selVal := canSelect.selectFieldOrMethod(sel.Sel.Name)
			return SDKType{typ: reflect.TypeOf(selVal.Interface())}
		}
		pkgType := stdtypes[sel.X.(Ident).Name][sel.Sel.Name]
		return SDKType{typ: reflect.TypeOf(pkgType.Interface())}
	}

	if star, ok := e.(StarExpr); ok {
		nonStarType := vm.proxyType(star.X)
		return nonStarType // .pointerType(). // TODO
	}

	if ar, ok := e.(ArrayType); ok {
		elemType := vm.makeType(ar.Elt)
		if ar.Len == nil {
			return SDKType{typ: reflect.SliceOf(elemType)}
		} else {
			lenVal := vm.returnsEval(ar.Len)
			size := int(lenVal.Int())
			return SDKType{typ: reflect.ArrayOf(size, elemType)}
		}
	}

	if _, ok := e.(FuncType); ok {
		// any function type will do; we just need its reflect.Type
		// TODO
		fn := func() {}
		return SDKType{typ: reflect.TypeOf(fn)}
	}

	if e, ok := e.(Ellipsis); ok {
		return vm.proxyType(e.Elt)
	}

	vm.fatal(fmt.Sprintf("unhandled proxyType for %v (%T)", e, e))
	return nil
}

// TODO only return a CanMake
func (vm *VM) makeType(e Evaluable) reflect.Type {
	if id, ok := e.(Ident); ok {
		typ, ok := builtins[id.Name]
		if ok {
			return typ.Interface().(builtinType).typ
		}
		return structValueType
	}
	if star, ok := e.(StarExpr); ok {
		nonStarType := vm.makeType(star.X)
		return reflect.PointerTo(nonStarType)
	}
	if sel, ok := e.(SelectorExpr); ok {
		typ := vm.localEnv().valueLookUp(sel.X.(Ident).Name)
		val := typ.Interface()
		if canSelect, ok := val.(CanSelect); ok {
			selVal := canSelect.selectFieldOrMethod(sel.Sel.Name)
			return reflect.TypeOf(selVal.Interface())
		}
		pkgType := stdtypes[sel.X.(Ident).Name][sel.Sel.Name]
		return reflect.TypeOf(pkgType.Interface())
	}
	if ar, ok := e.(ArrayType); ok {
		elemType := vm.makeType(ar.Elt)
		if ar.Len == nil {
			return reflect.SliceOf(elemType)
		} else {
			lenVal := vm.returnsEval(ar.Len)
			size := int(lenVal.Int())
			return reflect.ArrayOf(size, elemType)
		}
	}
	if _, ok := e.(FuncType); ok {
		// any function type will do; we just need its reflect.Type
		fn := func() {}
		return reflect.TypeOf(fn)
	}
	if e, ok := e.(Ellipsis); ok {
		return vm.makeType(e.Elt)
	}
	vm.fatal(fmt.Sprintf("unhandled makeType for %v (%T)", e, e))
	return nil
}

// pushOperand pushes a value onto the operand stack as the result of an evaluation.
func (vm *VM) pushOperand(v reflect.Value) {
	if trace {
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("vm.push: untyped nil\n")
			} else if isUndeclared(v) {
				fmt.Printf("vm.push: %v (undeclared)\n", v)
			} else {
				fmt.Printf("vm.push: %v (%T)\n", v.Interface(), v.Interface())
			}
		} else {
			fmt.Printf("vm.push: %v\n", v)
		}
	}
	vm.currentFrame.push(v)
}

// popOperand pops a value from the operand stack.
func (vm *VM) popOperand() reflect.Value {
	return vm.currentFrame.pop()
}

func (vm *VM) pushNewFrame(f Func) {
	if trace {
		fmt.Println("vm.pushNewFrame:", f)
	}
	frame := framePool.Get().(*stackFrame)
	frame.creator = f
	env := envPool.Get().(*Environment)
	env.parent = vm.localEnv()
	frame.env = env
	vm.callStack.push(frame)
	vm.currentFrame = frame
}

func (vm *VM) popFrame() {
	if trace {
		fmt.Println("vm.popFrame")
	}
	frame := vm.callStack.pop()
	if len(vm.callStack) > 0 {
		vm.currentFrame = vm.callStack.top()
	} else {
		vm.currentFrame = nil
	}

	// return env to pool
	env := frame.env.(*Environment)
	env.parent = nil
	// do not recycle environments that contain values referenced by a heap pointer
	if !env.hasHeapPointer {
		clear(env.valueTable)
		envPool.Put(env)
	}

	// reset references
	frame.operands = frame.operands[:0]
	frame.env = nil
	frame.creator = nil
	frame.defers = frame.defers[:0]
	framePool.Put(frame)
}

func (vm *VM) fatal(err any) {
	fmt.Fprintln(os.Stderr, "[gi] fatal error:", err)
	fmt.Fprintln(os.Stderr, "")
	vm.printStack()
	panic(err)
}

func (vm *VM) eval(e Evaluable) {
	if trace {
		fmt.Fprintln(os.Stderr, "vm.eval:", e)
	}
	e.Eval(vm)
}

func (vm *VM) takeAllStartingAt(head Step) {
	here := head
	for here != nil {
		if trace {
			fmt.Printf("%v @ %s\n", here, vm.sourceLocation(here))
		}
		here = here.take(vm)
	}
}

// sourceLocation returns a string representation of the source location of the given Evaluable (can be nil).
func (vm *VM) sourceLocation(e Evaluable) string {
	if e == nil {
		return "<no creator>"
	}
	if vm.fileSet == nil {
		return "<no file set>"
	}
	if f := vm.fileSet.File(e.Pos()); f != nil {
		nodir := filepath.Base(f.Name())
		return fmt.Sprintf("%s:%d", nodir, f.Line(e.Pos()))
	}
	return "<bad pos>"
}

func (vm *VM) printStack() {
	if len(vm.callStack) == 0 {
		fmt.Println("vm.ops: <empty>")
		return
	}
	frame := vm.currentFrame
	if env, ok := frame.env.(*PkgEnvironment); ok {
		for i, decl := range env.declarations {
			fmt.Printf("pkg.decl[%d]: %v\n", i, decl)
			if cd, ok := decl.(ConstDecl); ok {
				for s, spec := range cd.Specs {
					for n, idn := range spec.Names {
						fmt.Printf("  const.spec[%d][%d]: %v\n", s, n, idn.Name)
					}
				}
			}
		}
		for i, method := range env.inits {
			fmt.Printf("pkg.init[%d]: %v\n", i, method)
		}
		for i, method := range env.methods {
			fmt.Printf("pkg.method[%d]: %v\n", i, method)
		}
	}
	if env, ok := frame.env.(*Environment); ok {
		for k, v := range env.valueTable {
			if v.IsValid() && v.CanInterface() {
				if v == reflectNil {
					fmt.Printf("vm.env[%s]: untyped nil\n", k)
					continue
				}
				if isUndeclared(v) {
					fmt.Printf("vm.env[%s]: undeclared value\n", k)
					continue
				}
				fmt.Printf("vm.env.%s = %s (%T)\n", k, stringOf(v.Interface()), v.Interface())
			} else {
				fmt.Printf("vm.env.%s = %s\n", k, stringOf(v))
			}
		}
	}
	for i := 0; i < len(frame.operands); i++ {
		v := frame.operands[i]
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("vm.ops.%d: untyped nil\n", i)
				continue
			}
			if isUndeclared(v) {
				fmt.Printf("vm.ops.%d: undeclared value\n", i)
				continue
			}
			fmt.Printf("vm.ops.%d: %s (%T)\n", i, stringOf(v.Interface()), v.Interface())
		} else {
			fmt.Printf("vm.ops.%d: %s\n", i, stringOf(v))
		}
	}
}

func stringOf(v any) string {
	if v == nil {
		return "nil"
	}
	if ts, ok := v.(ToStringer); ok {
		return ts.toString()
	}
	if fs, ok := v.(fmt.Stringer); ok {
		return fs.String()
	}
	return fmt.Sprintf("%v", v)
}
