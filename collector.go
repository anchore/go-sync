package sync

import (
	"sync"
)

type ProviderFunc[T any] func() T

// Collector describes an executor that runs providers in parallel and returns the results from all providers
// after they have completed by using the Collect function
type Collector[T any] interface {
	// Provide is used to add providers to this collector
	Provide(provider ProviderFunc[T])

	// Collect waits for all providers to complete and returns the results from all executions
	// after an item has been returned by Collect, it will be guaranteed not to be returned again
	Collect() (everything []T)
}

func NewCollector[T any](executor Executor) Collector[T] {
	return &collector[T]{
		executor: executor,
	}
}

type collector[T any] struct {
	executor Executor
	out      []T
	mu       sync.Mutex
	wg       sync.WaitGroup
}

func (c *collector[T]) Provide(provider ProviderFunc[T]) {
	c.wg.Add(1)
	c.executor.Execute(func() {
		defer c.wg.Done()
		values := provider()
		c.mu.Lock()
		c.out = append(c.out, values)
		c.mu.Unlock()
	})
}

func (c *collector[T]) Collect() (everything []T) {
	c.wg.Wait()
	c.mu.Lock()
	everything = c.out
	c.out = nil
	c.mu.Unlock()
	return
}

var _ Collector[int] = (*collector[int])(nil)
