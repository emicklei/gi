package pkg

import (
	"fmt"
	"go/token"
	"strings"
)

type Field struct {
	names []*Ident // field/method/(type) parameter names; or nil
	typ   Expr     // field/method/parameter type; or nil
	tag   *string  // field tag; or nil
}

func (l Field) Eval(vm *VM) {} // noop

func (l Field) Pos() token.Pos { return token.NoPos } // TODO

func (l Field) String() string {
	return fmt.Sprintf("Field(%v,%v)", l.names, l.typ)
}

type FieldList struct {
	OpeningPos token.Pos
	List       []*Field
}

func (l FieldList) ensureNamedFields(meaning string) {
	for i, field := range l.List {
		if len(field.names) == 0 {
			field.names = []*Ident{{namePos: l.Pos(), name: internalVarName(meaning, i)}}
		}
	}
}

func (l FieldList) Pos() token.Pos { return l.OpeningPos }

func (l FieldList) String() string {
	names := []string{}
	for _, each := range l.List {
		for _, other := range each.names {
			names = append(names, other.name)
		}
	}
	return fmt.Sprintf("FieldList(%s)", strings.Join(names, ","))
}
func (l FieldList) Eval(vm *VM) {}
