package main

import "fmt"

func main() {
	// Test basic switch with tag
	x := 2
	switch x {
	case 1:
		fmt.Println("one")
	case 2:
		fmt.Println("two")
	case 3:
		fmt.Println("three")
	default:
		fmt.Println("other")
	}

	// Test switch with initialization
	switch y := 5; y {
	case 5:
		fmt.Println("five")
	case 10:
		fmt.Println("ten")
	}

	// Test switch without tag (bool expressions)
	z := 15
	switch {
	case z < 10:
		fmt.Println("less than 10")
	case z < 20:
		fmt.Println("less than 20")
	default:
		fmt.Println("20 or more")
	}
}
