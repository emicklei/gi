package pkg

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestNil(t *testing.T) {
	z := reflect.ValueOf(nil)
	t.Log(z.IsValid()) // false
	// panics
	// t.Log(z.IsZero())
	// panics
	// t.Log(z.IsNil())

	var err error = nil
	e := reflect.ValueOf(err)
	t.Log(e.IsValid()) // false
	// panics
	// t.Log(e.IsZero())
	// panics
	// t.Log(e.IsNil())
}

func TestBasicLitValue(t *testing.T) {
	tests := []struct {
		name     string
		lit      *ast.BasicLit
		wantKind reflect.Kind
		want     interface{}
	}{
		{
			name:     "int",
			lit:      &ast.BasicLit{Kind: token.INT, Value: "42"},
			wantKind: reflect.Int,
			want:     42,
		},
		{
			name:     "octal",
			lit:      &ast.BasicLit{Kind: token.INT, Value: "0600"},
			wantKind: reflect.Int,
			want:     384,
		},
		{
			name:     "string",
			lit:      &ast.BasicLit{Kind: token.STRING, Value: "\"hello\""},
			wantKind: reflect.String,
			want:     "hello",
		},
		{
			name:     "string with slash n",
			lit:      &ast.BasicLit{Kind: token.STRING, Value: "\"hello\n\""},
			wantKind: reflect.String,
			want:     "hello\n",
		},
		{
			name:     "raw string",
			lit:      &ast.BasicLit{Kind: token.STRING, Value: "`hello`"},
			wantKind: reflect.String,
			want:     "hello",
		},
		{
			name:     "float",
			lit:      &ast.BasicLit{Kind: token.FLOAT, Value: "3.14"},
			wantKind: reflect.Float64,
			want:     3.14,
		},
		{
			name:     "char",
			lit:      &ast.BasicLit{Kind: token.CHAR, Value: "'a'"},
			wantKind: reflect.Int32,
			want:     int32(97),
		},
		{
			name:     "imag",
			lit:      &ast.BasicLit{Kind: token.IMAG, Value: "1i"},
			wantKind: reflect.Complex128,
			want:     complex(0, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := basicLitValue(tt.lit)
			if got.Kind() != tt.wantKind {
				t.Errorf("basicLitValue() kind = %v (%T), want %v (%T)", got.Kind(), got.Kind(), tt.wantKind, tt.wantKind)
			}
			if got.Interface() != tt.want {
				t.Errorf("basicLitValue() = %v (%T), want %v (%T)", got.Interface(), got.Interface(), tt.want, tt.want)
			}
		})
	}
}
