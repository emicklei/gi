package main

import "fmt"

// swap swaps the values of two integers using pointers
func swap(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}

// increment increments a value through a pointer
func increment(p *int) {
	*p = *p + 1
}

func main() {
	// Basic pointer usage
	x := 42
	p := &x
	fmt.Println("Value of x:", x)
	fmt.Println("Value through pointer:", *p)

	// Modifying through pointer
	*p = 100
	fmt.Println("After *p = 100, x is:", x)

	// Pointer to string
	s := "hello"
	ps := &s
	*ps = "world"
	fmt.Println("String:", s)

	// Using swap function
	a := 5
	b := 10
	fmt.Println("Before swap: a =", a, ", b =", b)
	swap(&a, &b)
	fmt.Println("After swap: a =", a, ", b =", b)

	// Using increment function
	counter := 0
	for i := 0; i < 5; i++ {
		increment(&counter)
	}
	fmt.Println("Counter after 5 increments:", counter)
}
