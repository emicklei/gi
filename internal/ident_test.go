package internal

import (
	"go/ast"
	"testing"
)

func TestZeroValueOfType(t *testing.T) {
	env := newEnvironment(nil)
	i := Ident{Ident: &ast.Ident{Name: "string"}}
	v := i.ZeroValue(env)
	if v.Interface() != "" {
		t.Fail()
	}
}
