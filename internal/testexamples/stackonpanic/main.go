package main

func one(n int) int {
	return two(n + 1)
}
func two(n int) int {
	m := 5
	panic("oh no")
	return m + n
}
func main() {
	print(one(1))
}
