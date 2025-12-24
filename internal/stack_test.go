package internal

import (
	"testing"
)

func TestStack(t *testing.T) {
	var s stack[string]
	s.push("a")
	if got, want := s.top(), "a"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	s.push("b")
	if got, want := s.top(), "b"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := s.underTop(), "a"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := s.top(), "c"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := s.pop(), "c"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := s.pop(), "a"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}
