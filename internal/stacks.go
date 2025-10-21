package internal

import "reflect"

type valueStack struct {
	elements []reflect.Value
}

func (s valueStack) push(v reflect.Value) valueStack {
	s.elements = append(s.elements, v)
	return s
}
func (s valueStack) pop() (reflect.Value, valueStack) {
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

type frameStack struct {
	elements []stackFrame
}

func (s frameStack) push(v stackFrame) frameStack {
	s.elements = append(s.elements, v)
	return s
}
func (s frameStack) pop() (stackFrame, frameStack) {
	v := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return v, s
}
func (s frameStack) top() stackFrame {
	return s.elements[len(s.elements)-1]
}
func (s frameStack) isEmpty() bool {
	return len(s.elements) == 0
}
