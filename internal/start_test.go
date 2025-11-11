package internal

import "testing"

func TestPointerBasic(t *testing.T) {
	testMain(t, `package main

func main() {
	x := 42
	p := &x
	print(*p)
}`, "42")
}

func TestPointerAssignment(t *testing.T) {
	testMain(t, `package main

func main() {
	x := 10
	p := &x
	*p = 20
	print(x)
}`, "20")
}

func TestPointerTypedNil(t *testing.T) {
	testMain(t, `package main

func main() {
	nv4 := (*int64)(nil)
	print(nv4)
	v4 := int64(42)
	nv4 = &v4
	print(*nv4)
}`, "<nil>42")
}

func TestPointerMultipleAssignments(t *testing.T) {
	testMain(t, `package main

func main() {
	x := 1
	y := 2
	px := &x
	py := &y
	*px = 100
	*py = 200
	print(*px, *py)
}`, "100200")
}

func TestPointerToString(t *testing.T) {
	testMain(t, `package main

func main() {
	s := "hello"
	p := &s
	*p = "world"
	print(*p)
}`, "world")
}

func TestPointerSwap(t *testing.T) {
	testMain(t, `package main

func swap(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}
func main() {
	x := 5
	y := 10
	swap(&x, &y)
	print(x, y)
}`, "105")
}

func TestPointerToLocalVariable(t *testing.T) {
	testMain(t, `package main
func main() {
    // Test 1: Pointer reflects variable changes
    a := 1
    b := &a
    print(*b)  // Should be 1
    print(" ")
    a = 2
    print(*b)  // Should be 2 (pointer reflects the change)
    print(" ")
    // Test 2: Writing through pointer updates the variable
    *b = 42
    print(a)   // Should be 42
    print(" ")
    print(*b)  // Should be 42
    print(" ")
    // Test 3: Multiple pointers to same variable
    c := &a
    *c = 100
    print(a)   // Should be 100
    print(" ")
    print(*b)  // Should be 100 (b also points to a)
    print(" ")
    print(*c)  // Should be 100
}`, "1 2 42 42 100 100 100")
}

func TestPointerEscapeFromFunction(t *testing.T) {
	testMain(t, `package main
func ReturnPointer(val int) *int {
	return &val
}
func ModifyThroughPointer(p *int) {
	*p = 999
}
func main() {
	// Test 1: Pointer escapes function scope
	ptr := ReturnPointer(42)
	print(*ptr)  // Should be 42
	print(" ")
	
	// Test 2: Modify through returned pointer
	*ptr = 100
	print(*ptr)  // Should be 100
	print(" ")
	
	// Test 3: Pass pointer to function
	x := 5
	ModifyThroughPointer(&x)
	print(x)  // Should be 999
	print(" ")
	
	// Test 4: String pointers
	s := "hello"
	ps := &s
	print(*ps)  // Should be "hello"
	print(" ")
	s = "world"
	print(*ps)  // Should be "world" (pointer reflects change)
}`, "42 100 999 hello world")
}

func TestComplexPointerScenarios(t *testing.T) {
	testMain(t, `package main
type Point struct {
	X int
	Y int
}
func main() {
	// Test 1: Pointer to struct
	p := Point{X: 10, Y: 20}
	ptr := &p
	print(ptr.X)  // Should be 10
	print(" ")
	p.X = 30
	print(ptr.X)  // Should be 30 (pointer reflects change)
	print(" ")

	// Test 2: Modify struct through pointer
	ptr.Y = 99
	print(p.Y)  // Should be 99
	print(" ")
	
	// Test 3: Pointer arithmetic with different types
	i := 42
	pi := &i
	*pi = *pi + 8
	print(i)  // Should be 50
	print(" ")
	print(*pi)  // Should be 50
}`, "10 30 99 50 50")
}
