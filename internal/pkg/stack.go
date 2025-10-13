package pkg

import (
	"errors"
	"slices"
)

var ErrEmptyStack = errors.New("empty")

type Stack[T any] struct {
	stack []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		stack: make([]T, 0, 4),
	}
}

func (s *Stack[T]) Push(value T) {
	s.stack = append(s.stack, value)
}

func (s *Stack[T]) Pop() (T, error) {
	if len(s.stack) == 0 {
		var zero T
		return zero, ErrEmptyStack
	}
	value := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return value, nil
}

func (s *Stack[T]) Length() int {
	return len(s.stack)
}

func (s *Stack[T]) Shot() *Stack[T] {
	return &Stack[T]{
		stack: s.stack,
	}
}

func (s *Stack[T]) Clone() *Stack[T] {
	return &Stack[T]{
		stack: slices.Clone(s.stack),
	}
}
