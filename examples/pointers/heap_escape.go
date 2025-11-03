package main

import (
	"fmt"
)

func main() {
	a := 1
	b := &a
	fmt.Println(a, b)
	a = 2
	fmt.Println(a, b)
}
