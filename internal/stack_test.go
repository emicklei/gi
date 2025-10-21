package internal

import (
	"testing"
)

func TestStack(t *testing.T) {
	var s stack[string]
	if !s.isEmpty() {
		t.Error("new stack should be empty")
	}
	s.push("a")
	if s.isEmpty() {
		t.Error("stack should not be empty after push")
	}
	if got, want := s.top(), "a"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	s.push("b")
	if got, want := s.top(), "b"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	s.replaceTop("c")
	if got, want := s.top(), "c"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := s.pop(), "c"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := s.pop(), "a"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if !s.isEmpty() {
		t.Error("stack should be empty after popping all elements")
	}
}
