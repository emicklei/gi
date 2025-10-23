package internal

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestIncDec(t *testing.T) {
	cases := []struct {
		tok   token.Token
		start reflect.Value
		end   reflect.Value
	}{
		{token.INC, reflect.ValueOf(int(0)), reflect.ValueOf(int(1))},
		{token.DEC, reflect.ValueOf(int(1)), reflect.ValueOf(int(0))},
		{token.INC, reflect.ValueOf(int32(0)), reflect.ValueOf(int32(1))},
		{token.DEC, reflect.ValueOf(int32(1)), reflect.ValueOf(int32(0))},
		{token.INC, reflect.ValueOf(int64(0)), reflect.ValueOf(int64(1))},
		{token.DEC, reflect.ValueOf(int64(1)), reflect.ValueOf(int64(0))},
		{token.INC, reflect.ValueOf(float32(0)), reflect.ValueOf(float32(1))},
		{token.DEC, reflect.ValueOf(float32(1)), reflect.ValueOf(float32(0))},
		{token.INC, reflect.ValueOf(float64(0)), reflect.ValueOf(float64(1))},
		{token.DEC, reflect.ValueOf(float64(1)), reflect.ValueOf(float64(0))},
	}
	for _, tc := range cases {
		t.Run(tc.tok.String()+" "+tc.start.Kind().String(), func(t *testing.T) {
			vm := newVM(newEnvironment(nil))
			vm.localEnv().set("x", tc.start)
			n := IncDecStmt{
				IncDecStmt: &ast.IncDecStmt{Tok: tc.tok},
				X:          Ident{Ident: &ast.Ident{Name: "x"}},
			}
			if trace {
				vm.traceEval(n)
			} else {
				n.Eval(vm)
			}
			v := vm.localEnv().valueLookUp("x")
			if got, want := v.Interface(), tc.end.Interface(); got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
