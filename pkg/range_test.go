package pkg

import (
	"regexp"
	"testing"
)

func TestRangeOfChan(t *testing.T) {
	t.Skip()
	setAttr(t, "dot", true)
	testMain(t, `package main

func main() {
		io := make(chan int, 10)
		for i := 0; i < 10; i++ {
			io <- i
		}
		for {
			j := <-io
			print(j)
			if j == 4 {
				break
			}
		}
		for j := range io {
			print(j)
			if j == 9 {
				break
			}
		}
}`, "0123456789")
}

func TestRangeOfStrings(t *testing.T) {
	testMain(t, `package main

func main() {
	strings := []string{"hello", "world"}
	for i,s := range strings {
		print(i,s)
	}
}`, "0hello1world")
}

func TestRangeOfString(t *testing.T) {
	testMain(t, `package main

func main() {
	for _, s := range "gi" {
		print(s)
	}
}`, "103105")
}

func TestRangeOfStringsNoValue(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := range [2]string{} {
		print(i)
	}
}`, "01")
}

func TestRangeOfIntNoKey(t *testing.T) {
	testMain(t, `package main

func main() {
	for range 2 {
		print("a")
	}
}`, "aa")
}

func TestRangeOfIntWithKey(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := range 2 {
		print(i)
	}
}`, "01")
}

func TestRangeOfMap(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"a":1, "b":2}
	for k,v := range m {
		print(k,v)
	}
}`, func(out string) bool { return out == "a1b2" || out == "b2a1" })
}

func TestRangeNested(t *testing.T) {
	testMain(t, `package main

func main() {
	m := map[string]int{"a": 1, "b": 2}
	for j := range []int{0, 1} {
		for range j {
			for i := range 2 {
				for k, v := range m {
					print(i)
					print(k)
					print(v)
				}
			}
		}
	}
}`, func(out string) bool {
		// because map iteration is random we need to match all possibilities
		ok, _ := regexp.MatchString("^(?:0a10b2|0b20a1)(?:1a11b2|1b21a1)$", out)
		return ok
	})
}
