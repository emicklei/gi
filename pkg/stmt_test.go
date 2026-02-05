package pkg

import "testing"

func TestForBreak(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := 0; i < 5; i++ {
		if i == 3 {
			break
		}
		print(i)
	}
	print("-")
}
`, "012-")
}

func TestBreakOutOfEmptyFor(t *testing.T) {
	testMain(t, `package main

func main() {
	i := 0
	for {
		i++
		if i == 2 {
			break
		}
		print(i)
	}
}
`, "1")
}

func TestForBreakInner(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := 0; i < 5; i++ {
		for j := 1; j < 3 ; j++ {
			if j == 2 {
				break
			}
			print(j)
		}
		print(i)
	}
}
`, "1011121314")
}
func TestForNoContinue(t *testing.T) {
	testMain(t, `package main

func main() {
	for i := 0; i < 5; i++ {
		print(i)
	}
}
`, "01234")
}

func TestForContinue(t *testing.T) {
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

func TestContinueAndBreakInEmptyFor(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	i := 0
	for {
		i++
		if i == 2 {
			continue
		}
		print(i)
		if i == 4 {
			break
		}
	}
}
`, "134")
}

func TestForContinueInRange(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	for i := range 5 {
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

func TestDeferScope(t *testing.T) {
	testMain(t, `package main

func main() {
	a := 1
	defer func(b int) {
		print(a)
		print(b)
	}(a)
	a++
}
`, "21")
}

func TestDefer(t *testing.T) {
	testMain(t, `package main

func main() {
	a := 1
	defer print(a)
	a++
	defer print(a)
}`, "21")
}

func TestDeferFuncLiteral(t *testing.T) {
	testMain(t, `package main

func main() {
	f := func() {
		defer print(1)
	}
	f()
}`, "1")
}

func TestDeferInLoop(t *testing.T) {
	// i must be captured by value in the defer
	testMain(t, `package main	

func main(){
	for i := 0; i <= 3; i++ {
		defer print(i)
	}
}`, "3210")
}

func TestDeferInLoopInFuncLiteral(t *testing.T) {
	testMain(t, `package main

func main(){
	f := func() {
		for i := 0; i <= 3; i++ {
			defer print(i)
		}
	}
	f()
}`, "3210")
}
