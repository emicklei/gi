package pkg

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

func TestMakeChanIntWriteRead(t *testing.T) {
	//	setAttr(t, "dot", true)
	testMain(t, `package main

	func main() {
		c := make(chan int, 1)
		c <- 42
		v := <-c
		print(v)
	}`, func(out string) bool { return out == "42" })
}

func TestMakeChanUserWriteRead(t *testing.T) {
	testMain(t, `package main
	type User struct{Name string}
	func main() {
		c := make(chan User, 1)
		c <- User{Name: "Alice"}
		v := <-c
		print(v.Name)
	}`, func(out string) bool { return out == "Alice" })
}

func TestSelect(t *testing.T) {
	t.Skip()
	testMain(t, `package main
func main() {
    c1 := make(chan string,1)
    c2 := make(chan string,1)
    c1 <- "one"
    c2 <- "two"
    for range 2 {
        select {
        case msg1 := <-c1:
            print(msg1)
        case msg2 := <-c2:
            print(msg2)
        }
    }
}`, "onetwo")
}
