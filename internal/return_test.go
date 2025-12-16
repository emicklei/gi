package internal

import "testing"

func TestEarlyReturn(t *testing.T) {
	testMain(t, `package main

func main() {
	if true {
		print("2")
		return
	}
	print("0")
	return
}`, "2")
}

func TestFuncMultiReturn(t *testing.T) {
	testMain(t, `package main

func ab(a int, b int) (int,int) {
	print(a,b)
	// stack order: b, a
	return a,b
}
func main() {
	a,b := ab(2, 3)
	print(a,b)
}`, "2323")
}

func TestNamedReturn(t *testing.T) {
	t.Skip()
	testMain(t, `package main
		
func f() (result int) {
	result = 1
	return
}
func main(){
	print(f())
}`, "1")
}

func TestNamedEllipsisReturn(t *testing.T) {
	t.Skip()
	testMain(t, `package main
		
func f() (result1, result2 int) {
	result1 = 1
	result2 = 2
	return
}
func main(){
	r1,r2 := f()
	print(r1,r2)
}`, "12")
}

// https://go.dev/ref/spec#Defer_statements
func TestDeferReturnUpdateTestNestedLoop(t *testing.T) {
	// currently the returns puts all values on the top stackframe
	// but the defer can change the value from the environment
	// so we need to adjust the return value accordingly somehow
	t.Skip()
	testMain(t, `package main

func f() (result int) {
	defer func() {
		// result is accessed after it was set to 6 by the return statement
		result *= 7
	}()
	return 6
}
func main(){
	print(f())
}`, "42")
}
