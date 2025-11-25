package internal

import "testing"

func TestSwitchOnBool(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int = 1
	switch {
	case a == 1:
		print(a)
	}
}`, "1")
}
func TestSwitchNoExpresssion(t *testing.T) {
	testMain(t, `package main

func main() {
	switch a := 1;{
	case a < 2:
		print(a)
	}
}`, "1")
}

func TestSwitchOnLiteral(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int
	switch a = 1; a {
	case 1:
		print(a)
	}
}`, "1")
}

func TestSwitchDefault(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int
	switch a {
	case 2:
	default:
		print(3)
	}
}`, "3")
}

func TestSwitch(t *testing.T) {
	testMain(t, `package main

func main() {
	var a int
	switch a = 1; a {
	case 1:
		print(a)
	}
	switch a {
	case 2:
	default:
		print(3)
	}
}`, "13")
}

/**
a = 1
if a == 1 {
	print(a)
	goto end
}
if a == 2 {
	print(a)
	goto end
}
print(2)
end:
**/

func TestSwitchTypeAssign(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	var v any
	v = "gi"
	switch w := v.(type) {
	case int, int8:
		print("int:", w)
	case string:
		print("string:", w)
	default:
		print("unknown:", w)
	}
}`, "string:gi")
}
func TestSwitchTypeNoAssign(t *testing.T) {
	t.Skip()
	testMain(t, `package main

func main() {
	var i any
	i = 1
	switch i.(type) {
	case int:
		print("int:", i) 
	}
}`, "int:1")
}
