package internal

import (
	"bytes"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/emicklei/structexplorer"
)

// framePool is a pool of stackFrame instances for reuse.
var framePool = sync.Pool{
	New: func() any {
		return &stackFrame{
			operandStack: make([]reflect.Value, 0, 8),
			returnValues: make([]reflect.Value, 0, 2),
		}
	},
}

// stackFrame represents a single frame in the VM's function call stack.
type stackFrame struct {
	creator      Evaluable // typically a FuncDecl or FuncLit
	env          Env       // current environment with name->value mapping
	operandStack []reflect.Value
	returnValues []reflect.Value
	deferList    []funcInvocation
}

// push adds a value onto the operand stack.
func (f *stackFrame) push(v reflect.Value) {
	f.operandStack = append(f.operandStack, v)
}

// pop removes and returns the top value from the operand stack.
func (f *stackFrame) pop() reflect.Value {
	v := f.operandStack[len(f.operandStack)-1]
	f.operandStack = f.operandStack[:len(f.operandStack)-1]
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
	fmt.Fprintf(&buf, "ops=%v ", f.operandStack)
	fmt.Fprintf(&buf, "rets=%v", f.returnValues)
	return buf.String()
}

// Runtime represents a virtual machine that can execute Go code.
type VM struct {
	frameStack stack[*stackFrame]
	heap       *Heap
	output     *bytes.Buffer  // for testing only
	fileSet    *token.FileSet // optional file set for position info
}

var panicOnce sync.Once

func newVM(env Env) *VM {
	// TODO not sure when/if useful outside treerunner
	if os.Getenv("GI_IGNORE_EXIT") != "" {
		stdfuncs["os"]["Exit"] = reflect.ValueOf(func(code int) {
			fmt.Fprintf(os.Stderr, "[gi] os.Exit called with code %d\n", code)
		})
	}
	if os.Getenv("GI_IGNORE_PANIC") != "" {
		builtinsMap["panic"] = reflect.ValueOf(func(why any) {
			fmt.Fprintf(os.Stderr, "[gi] panic called with %v\n", why)
		})
	}
	vm := &VM{
		output:     new(bytes.Buffer),
		frameStack: make(stack[*stackFrame], 0, 16),
		heap:       newHeap()}
	frame := framePool.Get().(*stackFrame)
	frame.env = env
	vm.frameStack.push(frame)

	// TODO
	if os.Getenv("GI_IGNORE_PANIC") == "" {
		panicOnce.Do(func() {
			builtinsMap["panic"] = reflect.ValueOf(vm.fatal)
		})
	}

	return vm
}

func (vm *VM) setFileSet(fs *token.FileSet) {
	vm.fileSet = fs
}

// localEnv returns the current environment from the top stack frame.
func (vm *VM) localEnv() Env {
	return vm.frameStack.top().env
}

// returnsEval evaluates the argument and returns the popped value that was pushed onto the operand stack.
func (vm *VM) returnsEval(e Evaluable) reflect.Value {
	vm.eval(e)
	return vm.frameStack.top().pop()
}

func (vm *VM) returnsType(e Evaluable) reflect.Type {
	if id, ok := e.(Ident); ok {
		typ, ok := builtinTypesMap[id.Name]
		if ok {
			return typ
		}
		// non-imported user defined type
		return instanceType
	}
	if star, ok := e.(StarExpr); ok {
		nonStarType := vm.returnsType(star.X)
		return reflect.PointerTo(nonStarType)
	}
	if sel, ok := e.(SelectorExpr); ok {
		typ := vm.localEnv().valueLookUp(sel.X.(Ident).Name)
		val := typ.Interface()
		if canSelect, ok := val.(FieldSelectable); ok {
			selVal := canSelect.Select(sel.Sel.Name)
			return reflect.TypeOf(selVal.Interface())
		}
		pkgType := stdtypes[sel.X.(Ident).Name][sel.Sel.Name]
		return reflect.TypeOf(pkgType.Interface())
	}
	if ar, ok := e.(ArrayType); ok {
		elemType := vm.returnsType(ar.Elt)
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
	vm.fatal(fmt.Sprintf("unhandled returnsType for %v (%T)", e, e))
	return nil
}

// pushOperand pushes a value onto the operand stack as the result of an evaluation.
func (vm *VM) pushOperand(v reflect.Value) {
	if trace {
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("vm.push: untyped nil\n")
			} else if v == reflectUndeclared {
				fmt.Printf("vm.push: undeclared\n")
			} else {
				fmt.Printf("vm.push: %v (%T)\n", v.Interface(), v.Interface())
			}
		} else {
			fmt.Printf("vm.push: %v\n", v)
		}
	}
	vm.frameStack.top().push(v)
}

func (vm *VM) pushNewFrame(e Evaluable) { // typically a FuncDecl or FuncLit
	if trace {
		fmt.Println("vm.pushNewFrame:", e)
	}
	frame := framePool.Get().(*stackFrame)
	frame.creator = e
	env := envPool.Get().(*Environment)
	env.parent = vm.localEnv()
	frame.env = env
	vm.frameStack.push(frame)
}

func (vm *VM) popFrame() {
	if trace {
		fmt.Println("vm.popFrame")
	}
	frame := vm.frameStack.pop()

	// return env to pool
	env := frame.env.(*Environment)
	env.parent = nil
	// do not recycle environments that contain values referenced by a heap pointer
	if !env.hasHeapPointer {
		clear(env.valueTable)
		envPool.Put(env)
	}

	// reset references
	frame.operandStack = frame.operandStack[:0]
	frame.returnValues = frame.returnValues[:0]
	frame.env = nil
	frame.creator = nil
	frame.deferList = frame.deferList[:0]
	framePool.Put(frame)
}

func (vm *VM) fatal(err any) {
	fmt.Fprintln(os.Stderr, "[gi] fatal error:", err)
	fmt.Fprintln(os.Stderr, "")
	// dump the callstack
	for i := len(vm.frameStack) - 1; i >= 0; i-- {
		frame := vm.frameStack[i]
		fmt.Fprintln(os.Stderr, "[gi]", vm.sourceLocation(frame.creator), frame)
	}
	s := structexplorer.NewService("vm", vm)
	for i, each := range vm.frameStack {
		s.Explore(fmt.Sprintf("vm.frameStack.%d", i), each, structexplorer.Column(0))
		s.Explore(fmt.Sprintf("vm.frameStack.%d.env", i), each.env, structexplorer.Column(1))
		if tableHolder, ok := each.env.(*Environment); ok {
			s.Explore(fmt.Sprintf("vm.frameStack.%d.env.valueTable", i), tableHolder.valueTable, structexplorer.Column(2))
		}
		s.Explore(fmt.Sprintf("vm.frameStack.%d.operandStack", i), each.operandStack, structexplorer.Column(1))
		s.Explore(fmt.Sprintf("vm.frameStack.%d.returnValues", i), each.returnValues, structexplorer.Column(1))
	}
	s.Dump("gi-vm-panic.html")
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
			fmt.Print(here)
			fmt.Printf(" @ %s\n", vm.sourceLocation(here))
		}
		here = here.Take(vm)
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
	if len(vm.frameStack) == 0 {
		fmt.Println("vm.ops: <empty>")
		return
	}
	frame := vm.frameStack.top()
	for k, v := range frame.env.(*Environment).valueTable { // TODO hack
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("vm.env[%s]: untyped nil\n", k)
				continue
			}
			if v == reflectUndeclared {
				fmt.Printf("vm.env[%s]: undeclared value\n", k)
				continue
			}
			fmt.Printf("vm.env[%s]: %v (%T)\n", k, v.Interface(), v.Interface())
		} else {
			fmt.Printf("vm.env[%s]: %v\n", k, v)
		}
	}
	for i := 0; i < len(frame.operandStack); i++ {
		v := frame.operandStack[i]
		if v.IsValid() && v.CanInterface() {
			if v == reflectNil {
				fmt.Printf("vm.ops[%d]: untyped nil\n", i)
				continue
			}
			if v == reflectUndeclared {
				fmt.Printf("vm.ops[%d]: undeclared value\n", i)
				continue
			}
			fmt.Printf("vm.ops[%d]: %v (%T)\n", i, v.Interface(), v.Interface())
		} else {
			fmt.Printf("vm.ops[%d]: %v\n", i, v)
		}
	}
}
