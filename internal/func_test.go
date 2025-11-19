package internal

import "testing"

func TestFunc(t *testing.T) {
	testMain(t, `package main

func plus(a int, b int) int {
	return a + b
}
func main() {
	result := plus(2, 3)
	print(result)
}`, "5")
}

func TestFuncMultiReturn(t *testing.T) {
	testMain(t, `package main

func ab(a int, b int) (int,int) {
	return a,b
}
func main() {
	a,b := ab(2, 3)
	print(a,b)
}`, "23")
}

func TestNamedReturn(t *testing.T) {
	testMain(t, `package main
		
func f() (result int) {
	return 1 
}
func main(){
	print(f())
}`, "1")
}

func TestVariadicFunction(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func main() {
	print(sum(1, 2, 3))
}`, "6")
}

func TestFunctionLiteral(t *testing.T) {
	testMain(t, `package main

func main() {
	f := func(a int) int { return a }
	print(f(1))
}`, "1")
}

func TestFuncAsPackageVar(t *testing.T) {
	testMain(t, `package main

const h = "1"
var f = func() string { return h }

func main() {
	print(f())
}`, "1")
}
