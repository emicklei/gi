package pkg

import (
	"fmt"
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

func (e TypeAssertExpr) flow(g *graphBuilder) (head Step) {
	head = e.X.flow(g)
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

func (s TypeSpec) flow(g *graphBuilder) (head Step) {
	g.next(s)
	return g.current
}

func (s TypeSpec) makeValue(vm *VM, _ int, elements []reflect.Value) reflect.Value {
	actualType := vm.returnsEval(s.Type).Interface()
	if i, ok := actualType.(CanMake); ok {
		structVal := i.makeValue(vm, 0, elements)
		return structVal
	}
	vm.fatal(fmt.Sprintf("expected a CanMake value:%v", s.Type))
	return reflectNil
}

func (s TypeSpec) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	if c, ok := s.Type.(CanCompose); ok {
		return c.literalCompose(vm, composite, values)
	}
	return expected(s.Type, "a CanCompose value")
}

func (s TypeSpec) String() string {
	return fmt.Sprintf("TypeSpec(%v,%v,%v)", s.Name, s.TypeParams, s.Type)
}

func (s TypeSpec) Pos() token.Pos { return s.AssignPos }

var (
	_ Flowable = InterfaceType{}
	_ Expr     = InterfaceType{}
	_ CanMake  = InterfaceType{}
)

// InterfaceType represents an interface type definition.
// TODO needed?
type InterfaceType struct {
	InterfacePos token.Pos
	Methods      *FieldList
}

func (i InterfaceType) Eval(vm *VM) {
	vm.pushOperand(reflect.ValueOf(i))
}

func (i InterfaceType) flow(g *graphBuilder) (head Step) {
	g.next(i)
	return g.current
}

func (i InterfaceType) makeValue(vm *VM, size int, elements []reflect.Value) reflect.Value {
	if len(elements) > 0 {
		return elements[0]
	}
	return reflectNil
}

func (i InterfaceType) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	return reflectNil
}

func (i InterfaceType) Pos() token.Pos { return i.InterfacePos }

func (i InterfaceType) String() string {
	return fmt.Sprintf("InterfaceType(methods=%v)", i.Methods)
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

func (m MapType) flow(g *graphBuilder) (head Step) {
	g.next(m)
	return g.current
}

func (m MapType) makeValue(vm *VM, _ int, elements []reflect.Value) reflect.Value {
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
func (m MapType) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
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

func (m MapType) toString() string {
	return fmt.Sprintf("MapType(%v,%v)", m.Key, m.Value)
}

var reflectExtendedType = reflect.TypeFor[ExtendedValue]()

var _ HasMethods = ExtendedType{}

// type Count int
type ExtendedType struct {
	name    Ident
	methods map[string]*FuncDecl
}

func newExtendedType(name Ident) ExtendedType {
	return ExtendedType{
		name:    name,
		methods: map[string]*FuncDecl{},
	}
}
func (d ExtendedType) makeValue(vm *VM, size int, elements []reflect.Value) reflect.Value {
	// TODO rethink
	if len(d.methods) == 0 {
		// not extended after all
		return elements[0]
	}
	return reflect.ValueOf(ExtendedValue{
		typ: d,
		val: elements[0],
	})
}

func (d ExtendedType) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	return reflectNil
}

func (d ExtendedType) addMethod(decl *FuncDecl) { // TODO inline
	d.methods[decl.Name.Name] = decl
}

func (d ExtendedType) methodsMap() map[string]*FuncDecl {
	return d.methods
}

func (d ExtendedType) toString() string {
	return fmt.Sprintf("ExtendedType(%v,methods=%d)", d.name, len(d.methods))
}

// ExtendedValue represents a value of an ExtendedType.
type ExtendedValue struct {
	typ ExtendedType  // The typ is used for method resolution.
	val reflect.Value // The val field holds the actual reflect.Value.
}

func (e ExtendedValue) toString() string {
	return fmt.Sprintf("ExtendedValue(%v,%v)", e.typ.name.Name, e.val)
}

var _ CanMake = SDKType{}

type SDKType struct {
	typ reflect.Type // underlying Go type, builtin or struct
}

func (g SDKType) makeValue(vm *VM, size int, elements []reflect.Value) reflect.Value {
	pv := reflect.New(g.typ)
	if len(elements) == 1 {
		elm := elements[0]
		if elm.Kind() == reflect.Pointer {
			elm = elm.Elem()
		}
		pv.Elem().Set(elm.Convert(g.typ))
	}
	if g.typ.Kind() == reflect.Pointer {
		return pv
	} else {
		return pv.Elem()
	}
}

func (g SDKType) literalCompose(vm *VM, composite reflect.Value, values []reflect.Value) reflect.Value {
	return reflectNil
}
