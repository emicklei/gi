package pkg

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

func TestUnnamedReturn(t *testing.T) {
	testMain(t, `package main

func f() int { return 1 }
func main(){
	g := func() int { return 2 }
	print(f(),g())
}`, "12")
}

func TestNamedReturn(t *testing.T) {
	testMain(t, `package main

func f() (result int) {
	result = 1
	return
}
func main(){
	g := func() (result int) { result = 2 ; return  }
	print(f(),g())
}`, "12")
}

func TestNamedEllipsisReturn(t *testing.T) {
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
