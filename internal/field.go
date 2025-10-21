package internal

import (
	"fmt"
	"go/ast"
	"strings"
)

type Field struct {
	*ast.Field
	Names []*Ident
	Type  Expr
}

func (l Field) String() string {
	return fmt.Sprintf("Field(%v,%v)", l.Names, l.Type)
}
func (l Field) Eval(vm *VM) {}

type FieldList struct {
	*ast.FieldList
	List []*Field
}

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
