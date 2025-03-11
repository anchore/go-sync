package sync

import (
	"context"
	"errors"
	"iter"
	"sync"
)

// Collect iterates over the provided values, executing the processor in parallel to map each incoming value to a result.
// The collector is used to apply the results, with an exclusive lock; collector will never execute in parallel.
// All errors returned from processor functions will be joined with errors.Join as the returned error.
func Collect[From, To any](ctx context.Context, executorName string, values iter.Seq[From], collector func(context.Context, From, To), processor func(context.Context, From) (To, error)) error {
	var errs []error
	var lock sync.Mutex
	var wg sync.WaitGroup
	executor := GetExecutor(ctx, executorName)
	for value := range values {
		wg.Add(1)
		executor.Execute(ctx, func(ctx context.Context) {
			defer wg.Done()
			result, err := processor(ctx, value)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				errs = append(errs, err)
			}
			if collector != nil {
				collector(ctx, value, result)
			}
		})
	}
	wg.Wait()
	return errors.Join(errs...)
}

// CollectSlice is a specialized Collect call which appends results to a slice
func CollectSlice[From, To any](ctx context.Context, executorName string, values iter.Seq[From], slice *[]To, processor func(context.Context, From) (To, error)) error {
	return Collect(ctx, executorName, values, func(_ context.Context, _ From, value To) {
		*slice = append(*slice, value)
	}, processor)
}

// CollectMap is a specialized Collect call which fills a map using the incoming value as a key, mapped to the result
func CollectMap[From comparable, To any](ctx context.Context, executorName string, values iter.Seq[From], result map[From]To, processor func(context.Context, From) (To, error)) error {
	return Collect(ctx, executorName, values, func(_ context.Context, key From, value To) {
		result[key] = value
	}, processor)
}

// ToSeq converts a slice to an iter.Seq
func ToSeq[T any](values []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, value := range values {
			if !yield(value) {
				return
			}
		}
	}
}

// ToSlice takes an iter.Seq and returns a slice of the values returned
func ToSlice[T any](values iter.Seq[T]) (everything []T) {
	for v := range values {
		everything = append(everything, v)
	}
	return everything
}
