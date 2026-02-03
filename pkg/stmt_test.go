package pkg

import "testing"

func TestForBreak(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	for i := 0; i < 5; i++ {
		if i == 3 {
			break
		}
		print(i)
	}
}
`, "012")
}
func TestForContinue(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	for i := 0; i < 5; i++ {
		if i < 3 {
			continue
		}
		print(i)
	}
}
`, "34")
}

func TestGoto(t *testing.T) {
	testMain(t, `package main

func main() {
	s := 1
one:
	print(s)
	s++
	if s == 4 {
		return
	} else {
		goto two
	}
	print("unreachable")
two:
	print(s)
	s++
	goto one
}
`, "123")
}

func TestGotoInFunctionLiteral(t *testing.T) {
	testMain(t, `package main

func main() {
	f := func() {
		a := 1
	label:
		a++
		if a < 3 {
			goto label
		}
		print(a)
	}
	f()
}
`, "3")
}
