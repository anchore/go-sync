package sync

// Provider is the interface to use when returning a single item
type Provider[T any] interface {
	Get() T
}

// Iterator is used when there are multiple values to be operated on
type Iterator[T any] interface {
	Each(fn func(value T))
}

// Collection is a generic collection of values, which can be added to, removed from, and provide a length
type Collection[T any] interface {
	Iterator[T]
	Add(value T)
	Remove(value T)
	Contains(value T) bool
	Len() int
}

// Queue is a generic queue interface
type Queue[T any] interface {
	Enqueue(value T)
	Dequeue() (value T, ok bool)
}

// Stack is a generic stack interface
type Stack[T any] interface {
	Push(value T)
	Pop() (value T, ok bool)
	Peek() (value T, ok bool)
}
