package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
)

var _ Expr = TypeAssertExpr{}

type TypeAssertExpr struct {
	X      Expr
	Type   Expr // asserted type; nil means type switch X.(type)
	Lparen token.Pos
}

func (e TypeAssertExpr) Eval(vm *VM) {
	if e.Type == nil {
		val := vm.callStack.top().pop()
		valType := val.Type()
		vm.pushOperand(reflect.ValueOf(valType.Name())) // we compare strings
		// need the value for the assignment
		vm.pushOperand(val)
	}
}

func (e TypeAssertExpr) Flow(g *graphBuilder) (head Step) {
	head = e.X.Flow(g)
	g.next(e)
	return
}

func (e TypeAssertExpr) String() string {
	return fmt.Sprintf("TypeAssertExpr(%v,%v)", e.X, e.Type)
}

func (e TypeAssertExpr) Pos() token.Pos { return e.Lparen }

var _ CanMake = TypeSpec{}
var _ Expr = TypeSpec{}

type TypeSpec struct {
	Name       *Ident
	TypeParams *FieldList
	Type       Expr
	AssignPos  token.Pos
}

func (s TypeSpec) Eval(vm *VM) {
	actualType := vm.returnsEval(s.Type)
	vm.localEnv().set(s.Name.Name, actualType) // use the spec itself as value
}

func (s TypeSpec) Flow(g *graphBuilder) (head Step) {
	g.next(s)
	return g.current
}

func (s TypeSpec) Make(vm *VM, _ int, constructorArgs []reflect.Value) reflect.Value {
	actualType := vm.returnsEval(s.Type).Interface()
	if i, ok := actualType.(CanMake); ok {
		structVal := i.Make(vm, 0, constructorArgs)
		return structVal
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

func (s TypeSpec) String() string {
	return fmt.Sprintf("TypeSpec(%v,%v,%v)", s.Name, s.TypeParams, s.Type)
}

func (s TypeSpec) Pos() token.Pos { return s.AssignPos }

var (
	_ Flowable = StructType{}
	_ Expr     = StructType{}
)

type StructType struct {
	*ast.StructType
	StructPos token.Pos
	Name      string
	Fields    *FieldList
	methods   map[string]FuncDecl
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
		methods: map[string]FuncDecl{},
	}
}

func (s StructType) tagForField(fieldName string) *ast.BasicLit {
	for _, field := range s.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				return field.Tag
			}
		}
	}
	return nil
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

func (s StructType) Make(vm *VM, size int, constructorArgs []reflect.Value) reflect.Value {
	return reflect.ValueOf(NewStructValue(vm, s))
}

var (
	_ Flowable = MapType{}
	_ Expr     = MapType{}
	_ CanMake  = MapType{}
)

type MapType struct {
	MapPos token.Pos
	Key    Expr
	Value  Expr
}

func (m MapType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(m))
}

func (m MapType) Flow(g *graphBuilder) (head Step) {
	g.next(m)
	return g.current
}

func (m MapType) Make(vm *VM, _ int, constructorArgs []reflect.Value) reflect.Value {
	keyTypeName := mustIdentName(m.Key)
	valueTypeName := mustIdentName(m.Value)
	// standard or importer types
	keyType := vm.localEnv().typeLookUp(keyTypeName)
	if keyType == nil {
		// TODO handl custom types as key types
		keyType = vm.localEnv().valueLookUp(keyTypeName).Type()
	}
	valueType := vm.localEnv().typeLookUp(valueTypeName)
	mapType := reflect.MapOf(keyType, valueType)
	return reflect.MakeMap(mapType)
}
func (m MapType) LiteralCompose(composite reflect.Value, values []reflect.Value) reflect.Value {
	if composite.Kind() != reflect.Map {
		expected(composite, "map") // TODO bug if reached here?
	}
	for _, kv := range values {
		kv := kv.Interface().(keyValue)
		composite.SetMapIndex(kv.Key, kv.Value)
	}
	return composite
}

func (m MapType) Pos() token.Pos { return m.MapPos }

func (m MapType) String() string {
	return fmt.Sprintf("MapType(%v,%v)", m.Key, m.Value)
}
