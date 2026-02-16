package main

func foo(a, b int) int {
	return a + b
}

func main() {
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			if i == j {
				foo(i, j)
			} else if i < j {
				foo(j, i)
			} else {
				foo(i, j)
			}
		}
	}
}
