package index

import (
	"iter"

	"github.com/anchore/go-sync"
)

// Set implements an optimized, non-thread-safe collection of unordered values
type Set[T comparable] map[T]struct{}

var _ sync.Collection[int] = (*Set[int])(nil)

func NewSet[T comparable](values ...T) *Set[T] {
	out := &Set[T]{}
	out.AppendAll(sync.ToSeq(values))
	return out
}

func (s Set[T]) AppendAll(values iter.Seq[T]) {
	for value := range values {
		s[value] = struct{}{}
	}
}

func (s Set[T]) Seq(fn func(value T) bool) {
	for s := range s {
		if !fn(s) {
			return
		}
	}
}

func (s Set[T]) Append(value T) {
	s[value] = struct{}{}
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) Contains(value T) bool {
	_, ok := s[value]
	return ok
}

func (s Set[T]) Remove(value T) {
	delete(s, value)
}

func (s Set[T]) RemoveAll(values iter.Seq[T]) {
	for value := range values {
		delete(s, value)
	}
}

func (s Set[T]) RemoveFunc(remove SetFilterFunc[T]) {
	for v := range s {
		if !remove(v) {
			continue
		}
		delete(s, v)
	}
}

func (s Set[T]) KeepAll(values sync.Collection[T]) {
existingValue:
	for v := range s {
		if values.Contains(v) {
			continue existingValue
		}
		delete(s, v)
	}
}

func (s Set[T]) KeepFunc(keep SetFilterFunc[T]) {
	for v := range s {
		if keep(v) {
			continue
		}
		delete(s, v)
	}
}

type SetFilterFunc[T comparable] func(T) bool

func SetFilter[T comparable](set []T) SetFilterFunc[T] {
	return func(t T) bool {
		for _, remove := range set {
			if t == remove {
				return true
			}
		}
		return false
	}
}
