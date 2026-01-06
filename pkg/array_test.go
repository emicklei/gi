package pkg

import (
	"testing"
)

func TestSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	print([]int{1, 2})
}`, "[1 2]")
}

func TestMakeSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	s1 := make([]int, 1)
	s2 := make([]int, 2)
	print(len(s1), len(s2))
}`, "12")
}

func TestSliceLen(t *testing.T) {
	testMain(t, `package main

func main() {
	print(len([]int{1}))
}`, "1")
}

func TestSliceCap(t *testing.T) {
	testMain(t, `package main

func main() {
	print(cap([]int{1}))
}`, "1")
}

func TestArray(t *testing.T) {
	testMain(t, `package main

func main() {
	print([2]string{"A", "B"})
}`, "[A B]")
}

func TestArrayLen(t *testing.T) {
	testMain(t, `package main

func main() {
	print(len([2]string{"A", "B"}))
}`, "2")
}

func TestArrayCap(t *testing.T) {
	testMain(t, `package main

func main() {
	print(cap([2]string{"A", "B"}))
}`, "2")
}

func TestSliceClear(t *testing.T) {
	testMain(t, `package main

func main() {
	s := []int{1,2,3}
	clear(s)
	print(len(s))
}`, "3")
}

func TestSliceAppendAndIndex(t *testing.T) {
	testMain(t, `package main

func main() {
	list := []int{}
	list = append(list, 1, 2)
	print(list[0], list[1])
}`, "12")
}

func TestSubSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	list := []int{1,2,3}
	print(list[1:2])
}`, "[2]")
}

func TestReSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	a := [5]int{1, 2, 3, 4, 5}
	t := a[1:3:5]
	print(t)
}`, "[2 3]")
}

func TestEllipsisArray(t *testing.T) {
	testMain(t, `package main

func main() {
	arr := [...]int{1,2,3}	
	print(arr[0], arr[1], arr[2])
}
`, "123")
}

func TestCopy(t *testing.T) {
	testMain(t, `package main

func main() {
	src := []int{1,2,3}
	dest := make([]int, 3)
	n := copy(dest, src)
	print(n, dest[0], dest[1], dest[2])
}`, "3123")
}

func TestAppend(t *testing.T) {
	testMain(t, `package main

func main() {
	list := []int{}
	print(list)
	list = append(list, 4, 5)
	print(list)
}`, "[][4 5]")
}

func TestAppendStringToByteSlice(t *testing.T) {
	testMain(t, `package main

func main() {
	var b []byte
	b = append(b, "bar"...)
	print(string(b))
}`, "bar")
}

// https://go.dev/ref/spec#Appending_and_copying_slices
func TestCopySliceVariations(t *testing.T) {
	//t.Skip()
	testMain(t, `package main

func main() {	
	var a = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	var s = make([]int, 6)
	var b = make([]byte, 5)
	n1 := copy(s, a[0:])            // n1 == 6, s is []int{0, 1, 2, 3, 4, 5}
	n2 := copy(s, s[2:])            // n2 == 4, s is []int{2, 3, 4, 5, 4, 5}
	n3 := copy(b, "Hello, World!")  // n3 == 5, b is []byte("Hello")
	print(n1, n2, n3)
}`, "645")
}

func TestTwoDimArray(t *testing.T) {
	testMain(t, `package main

func main() {
	var a [2][3]string
	a[0] = [3]string{"foo", "bar", "baz"}
	a[1][0] = "foo"
	print(a[0][1])
	print(a[1][0])
}`, "barfoo")
}

func TestTwoDimIntArray(t *testing.T) {
	testMain(t, `package main

func main() {
	twoD := [2][3]int{
		{1, 2, 3},
		{1, 2, 3},
	}
	print(twoD[0][1])
}`, "2")
}
