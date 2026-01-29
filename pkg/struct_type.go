package pkg

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var (
	_ Flowable   = StructType{}
	_ Expr       = StructType{}
	_ HasMethods = StructType{}
)

// StructType represents a struct type definition that is interpreted (IType).
type StructType struct {
	StructPos token.Pos
	Name      string
	Fields    *FieldList
	methods   map[string]*FuncDecl
}

func (s StructType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(s))
}

func (s StructType) flow(g *graphBuilder) (head Step) {
	g.next(s)
	return g.current
}

func makeStructType(ast *ast.StructType) StructType {
	return StructType{
		StructPos: ast.Struct,
		methods:   map[string]*FuncDecl{},
	}
}

func (s StructType) tagForField(fieldName string) *string {
	for _, field := range s.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				return field.Tag
			}
		}
	}
	return nil
}

func (s StructType) Pos() token.Pos { return s.StructPos }

func (s StructType) String() string {
	n := s.Name
	if n == "" {
		n = "<anonymous>"
	}
	return fmt.Sprintf("StructType(%s,fields=%v,methods=%d)", n, s.Fields, len(s.methods))
}

// literalCompose initializes the composite StructType with the provided field values.
func (s StructType) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) (initializedComposite reflect.Value) {
	if len(values) == 0 {
		return composite
	}
	v := composite.Interface()
	i, ok := v.(CanCompose)
	if ok {
		return i.literalCompose(vm, composite, values)
	}
	// check for HeapPointer
	if hp, ok := asHeapPointer(composite); ok {
		v = vm.heap.read(hp).Interface()
		i, ok := v.(CanCompose)
		if ok {
			initializedComposite = i.literalCompose(vm, composite, values)
			vm.heap.write(hp, initializedComposite)
			return
		}
	}
	vm.fatal("expected a CanCompose value")
	return reflectNil // unreachable
}

func (s StructType) makeValue(vm *VM, size int, elements []reflect.Value) reflect.Value {
	return reflect.ValueOf(InstantiateStructValue(vm, s))
}

func (s StructType) addMethod(decl *FuncDecl) { // TODO inline
	s.methods[decl.Name.Name] = decl
}

func (s StructType) methodsMap() map[string]*FuncDecl {
	return s.methods
}
