package pkg

import (
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
		{token.INC, reflect.ValueOf(int8(0)), reflect.ValueOf(int8(1))},
		{token.DEC, reflect.ValueOf(int8(1)), reflect.ValueOf(int8(0))},
		{token.INC, reflect.ValueOf(int16(0)), reflect.ValueOf(int16(1))},
		{token.DEC, reflect.ValueOf(int16(1)), reflect.ValueOf(int16(0))},
		{token.INC, reflect.ValueOf(int32(0)), reflect.ValueOf(int32(1))},
		{token.DEC, reflect.ValueOf(int32(1)), reflect.ValueOf(int32(0))},
		{token.INC, reflect.ValueOf(int64(0)), reflect.ValueOf(int64(1))},
		{token.DEC, reflect.ValueOf(int64(1)), reflect.ValueOf(int64(0))},
		{token.INC, reflect.ValueOf(uint(0)), reflect.ValueOf(uint(1))},
		{token.DEC, reflect.ValueOf(uint(1)), reflect.ValueOf(uint(0))},
		{token.INC, reflect.ValueOf(uint8(0)), reflect.ValueOf(uint8(1))},
		{token.DEC, reflect.ValueOf(uint8(1)), reflect.ValueOf(uint8(0))},
		{token.INC, reflect.ValueOf(uint16(0)), reflect.ValueOf(uint16(1))},
		{token.DEC, reflect.ValueOf(uint16(1)), reflect.ValueOf(uint16(0))},
		{token.INC, reflect.ValueOf(uint32(0)), reflect.ValueOf(uint32(1))},
		{token.DEC, reflect.ValueOf(uint32(1)), reflect.ValueOf(uint32(0))},
		{token.INC, reflect.ValueOf(uint64(0)), reflect.ValueOf(uint64(1))},
		{token.DEC, reflect.ValueOf(uint64(1)), reflect.ValueOf(uint64(0))},
		{token.INC, reflect.ValueOf(float32(0)), reflect.ValueOf(float32(1))},
		{token.DEC, reflect.ValueOf(float32(1)), reflect.ValueOf(float32(0))},
		{token.INC, reflect.ValueOf(float64(0)), reflect.ValueOf(float64(1))},
		{token.DEC, reflect.ValueOf(float64(1)), reflect.ValueOf(float64(0))},
	}
	for _, tc := range cases {
		t.Run(tc.tok.String()+" "+tc.start.Kind().String(), func(t *testing.T) {
			p := &Package{env: newPkgEnvironment(nil)}
			vm := NewVM(p)
			frame := &stackFrame{env: p.env}
			vm.callStack.push(frame)
			vm.currentFrame = frame

			x := Ident{name: "x"}
			x.define(vm, tc.start)

			n := IncDecStmt{
				tok: tc.tok,
				x:   x,
			}
			vm.pushOperand(tc.start)
			n.eval(vm)

			v := vm.currentEnv().valueLookUp("x")
			if got, want := v.Interface(), tc.end.Interface(); got != want {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
