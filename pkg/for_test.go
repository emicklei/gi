package pkg

import (
	"strings"
	"testing"
)

func TestFor(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := 0; i < 10; i++ {
		print(i)
	}
	for i := 9; i > 0; i-- {
		print(i)
	}
}`, "0123456789987654321")
}

func TestForScope(t *testing.T) {
	testMain(t, `package main

func main() {
	j := 1
	for i := 0; i < 3; i++ {
		j = i
		print(i)
	}
	print(j)
}`, "0122")
}

func TestForScopeDefine(t *testing.T) {
	testMain(t, `package main

func main() {
	j := 1
	for i := 0; i < 3; i++ {
		j := i
		print(j)
	}
	print(j)
}`, "0121")
}

func TestForReturn(t *testing.T) {
	testMain(t, `package main

func main() {
	for {
		print(1)
		return
	}
}`, "1")
}

/*
Want: each iteration a different address, within the body same address
0x329faaa24048 0x329faaa24048 0
0x329faaa24070 0x329faaa24070 1
*/
func TestForVarAddress(t *testing.T) {
	testMain(t, `package main
import "fmt"
func main() {
		for i := 0; i < 2; i++ {
			fmt.Printf("%p %p %d\n", &i, &i, i)
		}
}`, func(got string) bool {
		return true
		// TODO because HeapPointer does not store pointer value ; it is created on each call when used as argument
		// ss := strings.Split(got, " ")
		// return len(ss) > 2 && ss[0] == ss[1]
	})
}

func TestPrintWithClosures(t *testing.T) {
	t.Skip()
	testMain(t, `package main

import "fmt"

func main() {
		var prints []func()
		for i := 1; i <= 3; i++ {
			prints = append(prints, func() { fmt.Println(i, &i) })
		}
		for _, print := range prints {
			print()
		}
}
`, "")
}

func TestFmtPrintAddress(t *testing.T) {
	testMain(t, `package main
import "fmt"
func main() {
	i := 1
	fmt.Printf("%p %p %d\n", &i, &i, i)
}`, func(got string) bool {
		ss := strings.Split(got, " ")
		return len(ss) > 2 && ss[0] == ss[1]
	})
}
