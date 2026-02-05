package pkg

import "testing"

func TestFor(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := 0; i < 10; i++ {
		print(i)
	}
	for i := 9; i > 0; i-- {
		print(i)
	}
}`, "0123456789987654321")
}

func TestForScope(t *testing.T) {
	testMain(t, `package main

func main() {
	j := 1
	for i := 0; i < 3; i++ {
		j = i
		print(i)
	}
	print(j)
}`, "0122")
}

func TestForScopeDefine(t *testing.T) {
	testMain(t, `package main

func main() {
	j := 1
	for i := 0; i < 3; i++ {
		j := i
		print(j)
	}
	print(j)
}`, "0121")
}
