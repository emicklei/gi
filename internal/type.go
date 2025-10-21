package internal

import (
	"fmt"
	"go/ast"
	"reflect"
)

type TypeSpec struct {
	*ast.TypeSpec
	Name       *Ident
	TypeParams *FieldList
	Type       Expr
}

func (s TypeSpec) String() string {
	return fmt.Sprintf("TypeSpec(%v,%v,%v)", s.Name, s.TypeParams, s.Type)
}

func (s TypeSpec) Eval(vm *VM) {
	if s.Name == nil {
		return // TODO ?
	}
	actualType := vm.returnsEval(s.Type)
	vm.localEnv().set(s.Name.Name, actualType) // use the spec itself as value
}

func (s TypeSpec) Instantiate(vm *VM) reflect.Value {
	actualType := vm.returnsEval(s.Type).Interface()
	// fmt.Println(actualType)
	if i, ok := actualType.(CanInstantiate); ok {
		instance := i.Instantiate(vm)
		// fmt.Println(instance)
		return instance
	}
	vm.fatal(fmt.Sprintf("expected a CanInstantiate value:%v", s.Type))
	return reflect.Value{}
}

func (s TypeSpec) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	if c, ok := s.Type.(CanCompose); ok {
		return c.LiteralCompose(composite, values)
	}
	return expected(s.Type, "a CanCompose value")
}

var (
	_ Flowable = StructType{}
	_ Expr     = StructType{}
)

type StructType struct {
	*ast.StructType
	Fields  *FieldList
	Methods map[string]FuncDecl
}

func (s StructType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(s))
}

func (s StructType) Flow(g *graphBuilder) (head Step) {
	g.next(s)
	return g.current
}

func makeStructType(ast *ast.StructType) StructType {
	return StructType{StructType: ast,
		Methods: map[string]FuncDecl{},
	}
}

func (s StructType) String() string {
	return fmt.Sprintf("StructType(%v)", s.Fields)
}

func (s StructType) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	i, ok := composite.Interface().(CanCompose)
	if !ok {
		expected(composite, "CanCompose")
	}
	return i.LiteralCompose(composite, values)
}

func (s StructType) Instantiate(vm *VM) reflect.Value {
	return reflect.ValueOf(NewInstance(vm, s))
}

var (
	_ Flowable = MapType{}
	_ Expr     = MapType{}
)

type MapType struct {
	*ast.MapType
	Key   Expr
	Value Expr
}

func (m MapType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(m))
}

func (m MapType) Flow(g *graphBuilder) (head Step) {
	g.next(m)
	return g.current
}

func (m MapType) Instantiate(vm *VM) reflect.Value {
	keyTypeName := mustIdentName(m.Key)
	valueTypeName := mustIdentName(m.Value)
	keyType := vm.localEnv().typeLookUp(keyTypeName)
	valueType := vm.localEnv().typeLookUp(valueTypeName)
	mapType := reflect.MapOf(keyType, valueType)
	return reflect.MakeMap(mapType)
}
func (m MapType) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	if composite.Kind() != reflect.Map {
		expected(composite, "map")
	}
	for _, kv := range values {
		kv := kv.Interface().(KeyValue)
		composite.SetMapIndex(kv.Key, kv.Value)
	}
	return composite
}

func (m MapType) String() string {
	return fmt.Sprintf("MapType(%v,%v)", m.Key, m.Value)
}
