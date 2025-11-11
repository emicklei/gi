package internal

import (
	"maps"
	"testing"
)

func TestMapKeysGo(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	seq := maps.Keys(m)
	for v := range seq {
		print(v)
	}
}
func TestMapKeys(t *testing.T) {
	t.Skip()
	testMain(t, `package main
import "maps"
func main() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	seq := maps.Keys(m)
	for v := range seq {
		print(v)
	}
}`, "123")
}
