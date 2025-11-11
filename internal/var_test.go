package internal

import "testing"

func TestVar(t *testing.T) {
	testMain(t, `package main

var (
	a = 1
	s string
	b bool
)
func main() {
	print(a,s,b)
}`, "1false")
}

func TestConstScope(t *testing.T) {
	testMain(t, `package main

var b = a
func main() {
	var b = a
	const a = 2
	print(a, b)
}
const a = 1`, "21")
}

func TestMultiVar(t *testing.T) {
	testMain(t, `package main

func main() {	
	var b, c int = 1, 2
	print(b,c)
}`, "12")
}
