package pkg

import (
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/google/go-dap"
)

var trace = os.Getenv("GI_TRACE") != ""

type Env interface {
	// value access
	valueLookUp(name string) reflect.Value
	valueOwnerOf(name string) Env
	valueSet(name string, value reflect.Value)
	valueUnset(name string)
	typeLookUp(name string) reflect.Type

	// hierarchy
	newChild() Env
	depth() int
	parent() Env
	rootPackageEnv() *PkgEnvironment

	// others
	funcLookUp(name string) reflect.Value
	addCanDeclare(cv CanDeclare)
	addDeclaration(stmt Stmt)

	// if marked, this env has references that escape to the heap
	// or is used by funcInvocation in a defer statement
	markSharedReferenced()

	// collect for debugging
	appendScopes(scopes []dap.Scope) []dap.Scope
	appendVariables(scopes []dap.Variable) []dap.Variable
}

type PkgEnvironment struct {
	Env
	declarations  []CanDeclare
	declarations2 []Stmt
	inits         []*FuncDecl
	methods       []*FuncDecl
	packageTable  map[string]*Package // path -> *Package
}

func newBuiltinsEnvironment(parent Env) Env {
	return &Environment{
		parentEnv:  parent,
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

func (p *PkgEnvironment) addDeclaration(stmt Stmt) {
	p.declarations2 = append(p.declarations2, stmt)
}

// rootPackageEnv returns the top-level package environment.
func (p *PkgEnvironment) rootPackageEnv() *PkgEnvironment {
	if p.parent() == nil {
		return p
	}
	if _, ok := p.parent().(*Environment); ok {
		return p
	}
	return p.parent().rootPackageEnv()
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

func (p *PkgEnvironment) appendScopes(scopes []dap.Scope) []dap.Scope {
	return append(scopes, dap.Scope{
		Name:               "package",
		VariablesReference: p.depth(),
	})
}

var envPool = sync.Pool{
	New: func() any {
		return &Environment{
			valueTable: map[string]reflect.Value{},
		}
	},
}

type Environment struct {
	parentEnv      Env
	valueTable     map[string]reflect.Value // TODO rename to symbolTable
	hasHeapPointer bool
}

func newEnvironment(parentOrNil Env) Env {
	return &Environment{
		parentEnv:  parentOrNil,
		valueTable: map[string]reflect.Value{},
	}
}

func (e *Environment) parent() Env {
	return e.parentEnv
}

func (e *Environment) depth() int {
	if e.parentEnv == nil {
		return 0
	}
	return e.parentEnv.depth() + 1
}

func (e *Environment) String() string {
	if e == nil {
		return "Environment(<nil>)"
	}
	return fmt.Sprintf("-- env[depth=%d,len=%d]", e.depth(), len(e.valueTable))
}

func (e *Environment) appendScopes(scopes []dap.Scope) []dap.Scope {
	return append(scopes, dap.Scope{
		Name:               "locals",
		VariablesReference: e.depth(),
		NamedVariables:     len(e.valueTable),
	})
}

func (e *Environment) appendVariables(vars []dap.Variable) []dap.Variable {
	for k, v := range e.valueTable {
		vars = append(vars, dap.Variable{
			Name:               k,
			Value:              stringOf(v),
			Type:               typeNameOf(v),
			VariablesReference: e.depth(),
		})
	}
	return vars
}

func (e *Environment) newChild() Env {
	return newEnvironment(e)
}

func (e *Environment) funcLookUp(name string) reflect.Value {
	return e.valueLookUp(name)
}

func (e *Environment) valueLookUp(name string) reflect.Value {
	current := e
	for current != nil {
		v, ok := current.valueTable[name]
		if ok {
			return v
		}
		if current.parentEnv == nil {
			return reflectUndeclared
		}
		// Continue iteration if parent is also an *Environment
		if env, ok := current.parentEnv.(*Environment); ok {
			current = env
		} else {
			// Parent is a different Env implementation, delegate to it
			return current.parentEnv.valueLookUp(name)
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
		if current.parentEnv == nil {
			return nil
		}
		// Continue iteration if parent is also an *Environment
		if env, ok := current.parentEnv.(*Environment); ok {
			current = env
		} else {
			// Parent is a different Env implementation, delegate to it
			return current.parentEnv.valueOwnerOf(name)
		}
	}
	return nil
}

func (e *Environment) valueSet(name string, value reflect.Value) {
	if name == "_" {
		return
	}
	e.valueTable[name] = value
	// trace after set
	if trace {
		fmt.Println(e, name, "=", stringOf(value))
	}
}
func (e *Environment) valueUnset(name string) {
	delete(e.valueTable, name)
}

func (e *Environment) addCanDeclare(cv CanDeclare) {}
func (e *Environment) addDeclaration(stmt Stmt)    {}

func (e *Environment) rootPackageEnv() *PkgEnvironment {
	if e.parentEnv == nil {
		return nil
	}
	return e.parentEnv.rootPackageEnv()
}

func (e *Environment) markSharedReferenced() {
	e.hasHeapPointer = true
}
