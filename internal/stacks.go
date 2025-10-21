package internal

import "reflect"

type valueStack struct {
	elements []reflect.Value
}

// returns a new stack with the value pushed
func (s valueStack) pushed(v reflect.Value) valueStack {
	s.elements = append(s.elements, v)
	return s
}

// returns the top value and a new stack with the top value popped
func (s valueStack) popped() (reflect.Value, valueStack) {
	v := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return v, s
}
func (s valueStack) top() reflect.Value {
	return s.elements[len(s.elements)-1]
}
func (s valueStack) isEmpty() bool {
	return len(s.elements) == 0
}
func (s valueStack) peek(offsetFromTop int) reflect.Value {
	i := len(s.elements) - 1 - offsetFromTop
	if i < 0 || i >= len(s.elements) {
		return reflect.Value{}
	}
	return s.elements[i]
}

type frameStack struct {
	elements []stackFrame
}

func (s *frameStack) push(v stackFrame) {
	s.elements = append(s.elements, v)
}
func (s *frameStack) pop() stackFrame {
	v := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return v
}
func (s *frameStack) top() stackFrame {
	return s.elements[len(s.elements)-1]
}
func (s *frameStack) isEmpty() bool {
	return len(s.elements) == 0
}
func (s *frameStack) replaceTop(v stackFrame) {
	s.elements[len(s.elements)-1] = v
}
