package pkg

import "testing"

func TestBlankIdentifier(t *testing.T) {
	testMain(t, `package main

func main() {
	_, h, _ := "gi", "flow", "!"
	print(h)
}`, "flow")
}
