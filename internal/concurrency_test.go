package internal

import (
	"strings"
	"testing"
)

func TestMakeChanInt(t *testing.T) {
	testMain(t, `package main

	func main() {
		c := make(chan int)
		bc := make(chan int, 100)
		print(c,bc)
	}`, func(out string) bool { return strings.HasPrefix(out, "0x") })
}
func TestMakeChanTime(t *testing.T) {

	testMain(t, `package main
	import "time"
	func main() {
		c := make(chan time.Time, 100)
		print(c)
	}`, func(out string) bool { return strings.HasPrefix(out, "0x") })
}
func TestMakeChanUserType(t *testing.T) {

	testMain(t, `package main
	type User struct{}
	func main() {
		c := make(chan User, 100)
		print(c)
	}`, func(out string) bool { return strings.HasPrefix(out, "0x") })
}

func TestMakeChanWriteRead(t *testing.T) {
	//t.Skip()
	testMain(t, `package main

	func main() {
		c := make(chan int, 1)
		c <- 42
		v := <-c
		print(v)
	}`, func(out string) bool { return out == "42" })
}
