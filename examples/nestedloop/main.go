package main

import (
	"fmt"
)

func squared(n int) int {
	return n * n
}

func main() {
	n := []int{42}
	i, j := 0, 2
	for k := i; k < j; k++ {
		for s := -1; s < k+1; s++ {
			n = append(n, squared(s))
			fmt.Println(i, j, k, s)
		}
	}
	fmt.Println(n)
}
