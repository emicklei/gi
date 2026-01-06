package pkg

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
		val := vm.popOperand()
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
	return reflectNil
}

func (s TypeSpec) LiteralCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	if c, ok := s.Type.(CanCompose); ok {
		return c.LiteralCompose(vm, composite, values)
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

func (s StructType) Flow(g *graphBuilder) (head Step) {
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

func (s StructType) LiteralCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	i, ok := composite.Interface().(CanCompose)
	if !ok {
		expected(composite, "CanCompose")
	}
	return i.LiteralCompose(vm, composite, values)
}

func (s StructType) Make(vm *VM, size int, constructorArgs []reflect.Value) reflect.Value {
	return reflect.ValueOf(NewStructValue(vm, s))
}

func (s StructType) addMethod(decl *FuncDecl) {
	s.methods[decl.Name.Name] = decl
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
		keyType = structValueType
	}
	valueType := vm.localEnv().typeLookUp(valueTypeName)
	mapType := reflect.MapOf(keyType, valueType)
	return reflect.MakeMap(mapType)
}
func (m MapType) LiteralCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	for _, kv := range values {
		kv := kv.Interface().(keyValue)
		k := kv.Key
		// check for ident
		if ik, ok := k.Interface().(Ident); ok {
			// Ident.Eval
			k = vm.localEnv().valueLookUp(ik.Name)
		}
		v := kv.Value
		composite.SetMapIndex(k, v)
	}
	return composite
}

func (m MapType) Pos() token.Pos { return m.MapPos }

func (m MapType) String() string {
	return fmt.Sprintf("MapType(%v,%v)", m.Key, m.Value)
}
