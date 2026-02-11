package main

func one(n int) int {
	return two(n + 1)
}
func two(n int) int {
	panic("oh no")
}
func main() {
	print(one(1))
}
