package internal

import (
	"reflect"
	"testing"
)

func TestValueStack(t *testing.T) {
	s := valueStack{}
	if !s.isEmpty() {
		t.Error("new stack should be empty")
	}
	v1 := reflect.ValueOf(1)
	s = s.push(v1)
	if s.isEmpty() {
		t.Error("stack should not be empty after push")
	}
	if got := s.top(); got != v1 {
		t.Errorf("top() = %v, want %v", got, v1)
	}
	v2 := reflect.ValueOf("hello")
	s = s.push(v2)
	if got := s.top(); got != v2 {
		t.Errorf("top() = %v, want %v", got, v2)
	}
	p2, s := s.pop()
	if p2 != v2 {
		t.Errorf("pop() = %v, want %v", p2, v2)
	}
	if got := s.top(); got != v1 {
		t.Errorf("top() = %v, want %v", got, v1)
	}
	p1, s := s.pop()
	if p1 != v1 {
		t.Errorf("pop() = %v, want %v", p1, v1)
	}
	if !s.isEmpty() {
		t.Error("stack should be empty after popping all elements")
	}
}
