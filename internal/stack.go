package internal

type stack[T any] []T

func (s *stack[T]) push(f T) {
	*s = append(*s, f)
}

// pre: stack not empty
func (s *stack[T]) replaceTop(t T) {
	(*s)[len(*s)-1] = t
}

// pre: stack not empty
func (s *stack[T]) pop() T {
	f := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return f
}

// pre: stack not empty
func (s stack[T]) top() T {
	return s[len(s)-1]
}
func (s stack[T]) isEmpty() bool {
	return len(s) == 0
}
