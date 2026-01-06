package pkg

import "testing"

func TestSimpleAssign(t *testing.T) {
	testMain(t, `package main	
func main() {
	s := "gi"
	print(s)
}`, "gi")
}

func TestMultiAssign(t *testing.T) {
	testMain(t, `package main	
func main() {
	in1, in2 := "gi", "flow"
	print(in1, in2)
}`, "giflow")
}

func TestIfMultiAssign(t *testing.T) {
	testMain(t, `package main

func main() {
	if got, want := min(1,2), 1; got == want {
		print("min")
	}
}`, "min")
}

func TestArrayFuncIndexAssign(t *testing.T) {
	// t.Skip()
	testMain(t, `package main	
func one() int { return 1 }
func main() {
	a := [2]int{0,1}
	a[one()] = 10
	print(a[1])
}`, "10")
}

func TestAssignToStructField(t *testing.T) {
	testMain(t, `package main

type Point struct {
	X int
	Y int
}
func main() {
	x := 5
	p := Point{X: 10}
	p.X = x
	p.Y = 20
	print(p.X, p.Y)
}`, "520")
}

func TestMapOk(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{}
	m["a"] = 1
	a, ok := m["a"]
	print(a,ok)
}`, "1true")
}
