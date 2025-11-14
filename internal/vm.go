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

func newVM(env Env) *VM {
	if os.Getenv("GI_IGNORE_EXIT") != "" {
		patchReflectRegistries()
	}
	vm := &VM{
		output:     new(bytes.Buffer),
		frameStack: make(stack[*stackFrame], 0, 16),
		heap:       newHeap()}
	frame := framePool.Get().(*stackFrame)
	frame.env = env
	vm.frameStack.push(frame)
	return vm
}

var patchOnce sync.Once

func patchReflectRegistries() {
	patchOnce.Do(func() {
		stdfuncs["os"]["Exit"] = reflect.ValueOf(func(code int) {
			fmt.Fprintf(os.Stderr, "[gi] os.Exit called with code %d\n", code)
		})
	})
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
		return reflect.SliceOf(elemType)
	}
	vm.fatal(fmt.Sprintf("unhandled returnsType for %v (%T)", e, e))
	return nil
}

// pushOperand pushes a value onto the operand stack as the result of an evaluation.
func (vm *VM) pushOperand(v reflect.Value) {
	if trace {
		if v.IsValid() && v.CanInterface() {
			fmt.Printf("vm.push: %v (%T)\n", v.Interface(), v.Interface())
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
		fmt.Fprintln(os.Stderr, "[gi]", frame)
	}
	s := structexplorer.NewService("vm", vm)
	for i, each := range vm.frameStack {
		s.Explore(fmt.Sprintf("vm.frameStack.%d", i), each, structexplorer.Column(0))
		s.Explore(fmt.Sprintf("vm.frameStack.%d.env", i), each.env, structexplorer.Column(1))
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
			if vm.fileSet != nil {
				f := vm.fileSet.File(here.Pos())
				if f != nil {
					nodir := filepath.Base(f.Name())
					fmt.Print(" @ ", nodir, ":", f.Line(here.Pos()))
				} else {
					fmt.Print(" @ bad file info")
				}
			}
			fmt.Println()
		}
		here = here.Take(vm)
	}
}
