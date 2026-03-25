package pkg

import "testing"

// This file contains a collection of quiz programs
// collected by Erik Dubbelboer and are used for testing purposes.

func TestQuiz1(t *testing.T) {
	testMain(t, `package main

func main() {
	s := []int{1,2,3}
	var p *int
	for _, i := range s {
	        if i == 2 {
	                p = &i
	        }
	}
	print(*p)
}`, "2")
}

func TestQuiz1Slim(t *testing.T) {
	testMain(t, `package main

func main() {
	i := 2
	var p *int
	p = &i
	print(*p)
}`, "2")
}
