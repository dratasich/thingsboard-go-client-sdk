package datastructures

type Set[T comparable] struct {
	elements map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{elements: make(map[T]struct{})}
}

func (s *Set[T]) Add(element T) {
	s.elements[element] = struct{}{}
}

func (s *Set[T]) Remove(element T) {
	delete(s.elements, element)
}

func (s *Set[T]) Contains(element T) bool {
	_, exists := s.elements[element]
	return exists
}

func (s *Set[T]) Iterator() <-chan T {
	ch := make(chan T)
	go func() {
		defer close(ch)
		for elem := range s.elements {
			ch <- elem
		}
	}()
	return ch
}

func (s *Set[T]) Size() int {
	return len(s.elements)
}
