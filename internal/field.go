package internal

import (
	"fmt"
	"go/token"
	"strings"
)

type Field struct {
	Names []*Ident // field/method/(type) parameter names; or nil
	Type  Expr     // field/method/parameter type; or nil
	Tag   *string  // field tag; or nil
}

func (l Field) Pos() token.Pos { return token.NoPos } // TODO

func (l Field) String() string {
	return fmt.Sprintf("Field(%v,%v)", l.Names, l.Type)
}
func (l Field) Eval(vm *VM) {}

type FieldList struct {
	OpeningPos token.Pos
	List       []*Field
}

func (l FieldList) ensureNamedFields(meaning string) {
	for i, field := range l.List {
		if len(field.Names) == 0 {
			field.Names = []*Ident{{NamePos: l.Pos(), Name: internalVarName(meaning, i)}}
		}
	}
}

func (l FieldList) Pos() token.Pos { return l.OpeningPos }

func (l FieldList) String() string {
	names := []string{}
	for _, each := range l.List {
		for _, other := range each.Names {
			names = append(names, other.Name)
		}
	}
	return fmt.Sprintf("FieldList(%s)", strings.Join(names, ","))
}
func (l FieldList) Eval(vm *VM) {}
