package pkg

import (
	"fmt"
	"os"
	"reflect"
	"sync"
)

var trace = os.Getenv("GI_TRACE") != ""

type Env interface {
	valueLookUp(name string) reflect.Value
	typeLookUp(name string) reflect.Type
	valueOwnerOf(name string) Env
	set(name string, value reflect.Value)
	unset(name string)
	newChild() Env
	depth() int
	getParent() Env
	addCanDeclare(cv CanDeclare)
	rootPackageEnv() *PkgEnvironment
	// if marked, this env has references that escape to the heap
	// or is used by funcInvocation in a defer statement
	markSharedReferenced()
}

type PkgEnvironment struct {
	Env
	declarations []CanDeclare
	inits        []*FuncDecl
	methods      []*FuncDecl
	packageTable map[string]*Package // path -> *Package
}

func newBuiltinsEnvironment(parent Env) Env {
	return &Environment{
		parent:     parent,
		valueTable: builtins,
	}
}

func newPkgEnvironment(parent Env) *PkgEnvironment {
	return &PkgEnvironment{
		Env:          newEnvironment(parent),
		packageTable: map[string]*Package{},
	}
}
func (p *PkgEnvironment) addInit(f *FuncDecl) {
	p.inits = append(p.inits, f)
}
func (p *PkgEnvironment) addMethod(f *FuncDecl) {
	p.methods = append(p.methods, f)
}

func (p *PkgEnvironment) addCanDeclare(cv CanDeclare) {
	p.declarations = append(p.declarations, cv)
}

// rootPackageEnv returns the top-level package environment.
func (p *PkgEnvironment) rootPackageEnv() *PkgEnvironment {
	if p.getParent() == nil {
		return p
	}
	if _, ok := p.getParent().(*Environment); ok {
		return p
	}
	return p.getParent().rootPackageEnv()
}

func (p *PkgEnvironment) String() string {
	if p == nil || p.packageTable == nil {
		return "PkgEnvironment(<nil>)"
	}
	return fmt.Sprintf("PkgEnvironment(pkgs=%d)", len(p.packageTable))
}

func (p *PkgEnvironment) newChild() Env {
	return newEnvironment(p)
}
func (p *PkgEnvironment) markSharedReferenced() {}

var envPool = sync.Pool{
	New: func() any {
		return &Environment{
			valueTable: map[string]reflect.Value{},
		}
	},
}

type Environment struct {
	parent         Env
	valueTable     map[string]reflect.Value
	hasHeapPointer bool
}

func newEnvironment(parentOrNil Env) Env {
	return &Environment{
		parent:     parentOrNil,
		valueTable: map[string]reflect.Value{},
	}
}

func (e *Environment) getParent() Env {
	return e.parent
}

func (e *Environment) depth() int {
	if e.parent == nil {
		return 0
	}
	return e.parent.depth() + 1
}

func (e *Environment) String() string {
	if e == nil {
		return "Environment(<nil>)"
	}
	return fmt.Sprintf("-- env[depth=%d,len=%d]", e.depth(), len(e.valueTable))
}

func (e *Environment) newChild() Env {
	return newEnvironment(e)
}
func (e *Environment) valueLookUp(name string) reflect.Value {
	current := e
	for current != nil {
		v, ok := current.valueTable[name]
		if ok {
			return v
		}
		if current.parent == nil {
			return reflectUndeclared
		}
		// Continue iteration if parent is also an *Environment
		if env, ok := current.parent.(*Environment); ok {
			current = env
		} else {
			// Parent is a different Env implementation, delegate to it
			return current.parent.valueLookUp(name)
		}
	}
	return reflectUndeclared
}

func (e *Environment) typeLookUp(name string) reflect.Type {
	v, ok := builtins[name]
	if !ok {
		return nil
	}
	return v.Interface().(builtinType).typ
}

func (e *Environment) valueOwnerOf(name string) Env {
	current := e
	for current != nil {
		if _, ok := current.valueTable[name]; ok {
			return current
		}
		if current.parent == nil {
			return nil
		}
		// Continue iteration if parent is also an *Environment
		if env, ok := current.parent.(*Environment); ok {
			current = env
		} else {
			// Parent is a different Env implementation, delegate to it
			return current.parent.valueOwnerOf(name)
		}
	}
	return nil
}

func (e *Environment) set(name string, value reflect.Value) {
	if name == "_" {
		return
	}
	e.valueTable[name] = value
	// trace after set
	if trace {
		fmt.Println(e, name, "=", stringOf(value))
	}
}
func (e *Environment) unset(name string) {
	delete(e.valueTable, name)
}

func (e *Environment) addCanDeclare(cv CanDeclare) {}

func (e *Environment) rootPackageEnv() *PkgEnvironment {
	if e.parent == nil {
		return nil
	}
	return e.parent.rootPackageEnv()
}

func (e *Environment) markSharedReferenced() {
	e.hasHeapPointer = true
}
