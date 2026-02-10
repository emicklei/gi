package pkg

import (
	"fmt"
	"strings"
	"testing"
)

func assertPanicContains(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic containing %q, got none", want)
		} else {
			got := fmt.Sprint(r)
			if !strings.Contains(got, want) {
				t.Fatalf("expected panic containing %q, got %q", want, got)
			}
		}
	}()
	fn()
}

func TestToString(t *testing.T) {
	if got := toString("hello"); got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}
	if got := toString([]byte("world")); got != "world" {
		t.Fatalf("expected world, got %q", got)
	}
	assertPanicContains(t, "string convert undefined", func() {
		toString(42)
	})
}

func TestToFloat32(t *testing.T) {
	var want float32 = 12.5
	if got := toFloat32(float32(12.5)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := toFloat32(float64(12.5)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	assertPanicContains(t, "float32 convert undefined", func() {
		toFloat32("nope")
	})
}

func TestToFloat64(t *testing.T) {
	var want float64 = 19.75
	if got := toFloat64(float32(19.75)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := toFloat64(float64(19.75)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	assertPanicContains(t, "float64 convert undefined", func() {
		toFloat64("nope")
	})
}

func TestToComplex64(t *testing.T) {
	want := complex64(3 + 4i)
	if got := toComplex64(complex64(3 + 4i)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := toComplex64(complex128(3 + 4i)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	assertPanicContains(t, "complex64 convert undefined", func() {
		toComplex64(3.14)
	})
}

func TestToComplex128(t *testing.T) {
	want := complex128(5 + 6i)
	if got := toComplex128(complex64(5 + 6i)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := toComplex128(complex128(5 + 6i)); got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
	assertPanicContains(t, "complex128 convert undefined", func() {
		toComplex128(1.23)
	})
}

func TestToInt(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  int
	}{
		{"int", int(42), 42},
		{"int8", int8(42), 42},
		{"int16", int16(42), 42},
		{"int32", int32(42), 42},
		{"int64", int64(42), 42},
		{"float64", float64(42.99), 42},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toInt(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "int convert undefined", func() {
		toInt("nope")
	})
}

func TestToInt8(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  int8
	}{
		{"int", int(7), 7},
		{"int8", int8(8), 8},
		{"int16", int16(9), 9},
		{"int32", int32(10), 10},
		{"int64", int64(11), 11},
		{"float64", float64(12.7), 12},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toInt8(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "int8 convert undefined", func() {
		toInt8("nope")
	})
}

func TestToInt16(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  int16
	}{
		{"int", int(13), 13},
		{"int8", int8(14), 14},
		{"int16", int16(15), 15},
		{"int32", int32(16), 16},
		{"int64", int64(17), 17},
		{"float64", float64(18.2), 18},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toInt16(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "int16 convert undefined", func() {
		toInt16("nope")
	})
}

func TestToInt32(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  int32
	}{
		{"int", int(19), 19},
		{"int8", int8(20), 20},
		{"int16", int16(21), 21},
		{"int32", int32(22), 22},
		{"int64", int64(23), 23},
		{"float64", float64(24.9), 24},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toInt32(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "int32 convert undefined", func() {
		toInt32("nope")
	})
}

func TestToInt64(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  int64
	}{
		{"int", int(25), 25},
		{"int8", int8(26), 26},
		{"int16", int16(27), 27},
		{"int32", int32(28), 28},
		{"int64", int64(29), 29},
		{"float64", float64(30.5), 30},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toInt64(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "int64 convert undefined", func() {
		toInt64("nope")
	})
}

func TestToUint(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  uint
	}{
		{"int", int(31), 31},
		{"uint", uint(32), 32},
		{"uint8", uint8(33), 33},
		{"uint16", uint16(34), 34},
		{"uint32", uint32(35), 35},
		{"uint64", uint64(36), 36},
		{"float64", float64(37.8), 37},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toUint(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "uint convert undefined", func() {
		toUint("nope")
	})
}

func TestToUint8(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  uint8
	}{
		{"int", int(38), 38},
		{"uint", uint(39), 39},
		{"uint8", uint8(40), 40},
		{"uint16", uint16(41), 41},
		{"uint32", uint32(42), 42},
		{"uint64", uint64(43), 43},
		{"float64", float64(44.3), 44},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toUint8(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "uint8 convert undefined", func() {
		toUint8("nope")
	})
}

func TestToUint16(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  uint16
	}{
		{"int", int(45), 45},
		{"uint", uint(46), 46},
		{"uint8", uint8(47), 47},
		{"uint16", uint16(48), 48},
		{"uint32", uint32(49), 49},
		{"uint64", uint64(50), 50},
		{"float64", float64(51.6), 51},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toUint16(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "uint16 convert undefined", func() {
		toUint16("nope")
	})
}

func TestToUint32(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  uint32
	}{
		{"int", int(52), 52},
		{"uint", uint(53), 53},
		{"uint8", uint8(54), 54},
		{"uint16", uint16(55), 55},
		{"uint32", uint32(56), 56},
		{"uint64", uint64(57), 57},
		{"float64", float64(58.4), 58},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toUint32(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "uint32 convert undefined", func() {
		toUint32("nope")
	})
}

func TestToUint64(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  uint64
	}{
		{"int", int(59), 59},
		{"uint", uint(60), 60},
		{"uint8", uint8(61), 61},
		{"uint16", uint16(62), 62},
		{"uint32", uint32(63), 63},
		{"uint64", uint64(64), 64},
		{"float64", float64(65.9), 65},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := toUint64(tc.input); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
	assertPanicContains(t, "uint64 convert undefined", func() {
		toUint64("nope")
	})
}

func TestToBool(t *testing.T) {
	if !toBool(true) {
		t.Fatalf("expected true")
	}
	assertPanicContains(t, "bool convert error", func() {
		toBool("nope")
	})
}
