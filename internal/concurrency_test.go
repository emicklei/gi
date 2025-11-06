package internal

import (
	"strings"
	"testing"
)

func TestMakeChanInt(t *testing.T) {
	testMain(t, `package main

	func main() {
		c := make(chan int, 100)
		print(c)
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
