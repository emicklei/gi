package pkg

import (
	"fmt"
	"testing"
)

func TestMapClear(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"A":1, "B":2}
	clear(m)
	print(len(m))
}`, "0")
}

func TestRecover(t *testing.T) {
	testMain(t, `package main

func main() {
	defer func() {
		print(recover())
	}()
	panic("0")
}`, "0")
}

func TestNoPanicRecover(t *testing.T) {
	testMain(t, `package main

func main() {
	defer func() {
		print(recover())
	}()
}`, "[]")
}

func TestNestedRecover(t *testing.T) {
	testMain(t, `package main

func catchthrow() {
	defer func() {
		r := recover().(string)
		print(r)
		panic(r + "-caught")
	}()
	panic("hi")
}

func main() {
	defer func() {
		r := recover()
		print(r)
	}()
	catchthrow()
}`, "hihi-caught")
}

func TestMinMax(t *testing.T) {
	testMain(t, `package main

func main() {
	print(min(1,2), max(1,2))
}`, "12")
}
func TestMaxAtLeast(t *testing.T) {
	testMain(t, `package main

func main() {
	print(max(1,2,10))
	print(max(1,5,3))
}`, "105")
}

func TestMinAtMost(t *testing.T) {
	testMain(t, `package main

func main() {
	print(min(3,2,1))
	print(min(4,2,3))
}`, "12")
}

func TestMaxString(t *testing.T) {
	testMain(t, `package main

func main() {
	print(max("", "foo", "bar"))
}`, "foo")
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if got, want := fmt.Sprint(r), "oops"; got != want {
				t.Errorf("got [%[1]v:%[1]T] want [%[2]v:%[2]T]", got, want)
			}
		}
	}()
	testMain(t, `package main

func main() {
	panic("oops")
}`, "")
}
