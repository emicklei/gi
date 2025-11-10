package internal

import (
	"fmt"
	"maps"
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
	newChild() Env
	depth() int
	getParent() Env
	addConstOrVar(cv ConstOrVar)
	rootPackageEnv() *PkgEnvironment
	markContainsHeapPointer()
}

type PkgEnvironment struct {
	Env
	declarations []CanDeclare
	inits        []FuncDecl
	packageTable map[string]*Package // path -> *Package
}

func newBuiltinsEnvironment(parent Env) Env {
	return &Environment{
		parent:     parent,
		valueTable: builtinsMap,
	}
}

func newPkgEnvironment(parent Env) *PkgEnvironment {
	return &PkgEnvironment{
		Env:          newEnvironment(parent),
		packageTable: map[string]*Package{},
	}
}
func (p *PkgEnvironment) addInit(f FuncDecl) {
	p.inits = append(p.inits, f)
}

func (p *PkgEnvironment) addConstOrVar(cv ConstOrVar) {
	p.declarations = append(p.declarations, CanDeclare(cv))
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
func (p *PkgEnvironment) markContainsHeapPointer() {}

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
	v, ok := e.valueTable[name]
	if !ok {
		if e.parent == nil {
			return reflect.Value{}
		}
		return e.parent.valueLookUp(name)
	}
	return v
}

func (e *Environment) typeLookUp(name string) reflect.Type {
	v, ok := builtinTypesMap[name]
	if !ok {
		return nil
	}
	return v
}

func (e *Environment) valueOwnerOf(name string) Env {
	_, ok := e.valueTable[name]
	if !ok {
		if e.parent == nil {
			return nil
		}
		return e.parent.valueOwnerOf(name)
	}
	return e
}

func (e *Environment) set(name string, value reflect.Value) {
	if name == "_" {
		return
	}
	e.valueTable[name] = value
	// trace after set
	if trace {
		fmt.Println(e, name, "=", value.Interface())
	}
}

func (e *Environment) addConstOrVar(cv ConstOrVar) {}

func (e *Environment) rootPackageEnv() *PkgEnvironment {
	if e.parent == nil {
		return nil
	}
	return e.parent.rootPackageEnv()
}

func (e *Environment) markContainsHeapPointer() {
	e.hasHeapPointer = true
}

// clone creates new environment sharing the parent and cloning the value table.
func (e *Environment) clone() *Environment {
	return &Environment{
		parent:     e.parent,
		valueTable: maps.Clone(e.valueTable),
	}
}
