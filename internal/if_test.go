package internal

import "testing"

func TestIfElseIfElse(t *testing.T) {
	testMain(t, `package main

func main() {
	if 1 == 2 {
		print("unreachable 1")
	} else if 2 == 2 {
		print("gi")
	} else {
		print("unreachable 2")
	}
}`, "gi")
}

func TestIfIf(t *testing.T) {
	testMain(t, `package main

func main() {
	if 1 == 2 {
		print("unreachable")
	} 
	if 2 == 2 {
		print("gi")
	}
}`, "gi")
}
