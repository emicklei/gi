package pkg

import "testing"

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
