package pkg

import (
	"fmt"
	"testing"
)

func TestUnaries(t *testing.T) {
	tests := []struct {
		src  string
		op   string
		want string
	}{
		{"true", "!", "false"},
		{"int(1)", "^", "-2"},
		{"int8(1)", "^", "-2"},
		{"int16(1)", "^", "-2"},
		{"int32(1)", "^", "-2"},
		{"int64(1)", "^", "-2"},
		{"uint64(1)", "^", "18446744073709551614"},
		{"uint32(1)", "^", "4294967294"},
		{"uint16(1)", "^", "65534"},
		{"uint8(1)", "^", "254"},
		{"uint(1)", "^", "18446744073709551614"},
		{"int(1)", "+", "1"},
		{"int8(1)", "+", "1"},
		{"int16(1)", "+", "1"},
		{"int32(1)", "+", "1"},
		{"int64(1)", "+", "1"},
		{"uint64(1)", "+", "1"},
		{"uint32(1)", "+", "1"},
		{"uint16(1)", "+", "1"},
		{"uint8(1)", "+", "1"},
		{"uint(1)", "+", "1"},
		{"int(1)", "+", "1"},
		{"int8(1)", "-", "-1"},
		{"int16(1)", "-", "-1"},
		{"int32(1)", "-", "-1"},
		{"int64(1)", "-", "-1"},
		{"uint64(1)", "-", "18446744073709551615"},
		{"uint32(1)", "-", "4294967295"},
		{"uint16(1)", "-", "65535"},
		{"uint8(1)", "-", "255"},
		{"uint(1)", "-", "18446744073709551615"},
		{"float32(1.0)", "-", "-1"},
		{"float32(1.0)", "+", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			t.Parallel()
			src := fmt.Sprintf(`
			package main

			func main() {
				v := %s
				print(%sv)
			}`, tt.src, tt.op)
			out := parseAndWalk(t, src)
			if got, want := out, tt.want; got != want {
				t.Errorf("%s got [%[1]v:%[1]T] want [%[2]v:%[2]T]", tt.src, got, want)
			}
		})
	}
}

func TestNumbers(t *testing.T) {
	testMain(t, `package main

	func main() {
		print(-1,+3.14,0.1e10)
	}`, "-13.141e+09")
}

func TestUnaryComplex(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"complex64(1+2i)", "(1+2i)"},
		{"+complex64(1+2i)", "(1+2i)"},
		{"-complex64(1+2i)", "(-1-2i)"},
		{"complex128(1+2i)", "(1+2i)"},
		{"+complex128(1+2i)", "(1+2i)"},
		{"-complex128(1+2i)", "(-1-2i)"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			src := `
			package main
			func main() {
				print(` + tt.src + `)
			}`
			out := parseAndWalk(t, src)
			if got, want := out, tt.want; got != want {
				t.Errorf("%s got %s want %s", tt.src, got, want)
			}
		})
	}
}
