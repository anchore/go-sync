package sync

import (
	"context"
)

type executorKey struct {
	name string
}

// HasExecutor returns true when the named executor is available in the context
func HasExecutor(ctx context.Context, name string) bool {
	return ctx.Value(executorKey{name: name}) != nil
}

// GetExecutor returns an executor in context with the given name, or a serial executor if none exists
// and replaces the context with one that contains a new executor which won't deadlock
func GetExecutor(ctx *context.Context, name string) Executor {
	if ctx == nil {
		return sequentialExecutor{}
	}
	executor := getExecutor(*ctx, name)
	*ctx = SetContextExecutor(*ctx, name, executor.ChildExecutor())
	return executor
}

// SetContextExecutor returns a context with the named provided executor for use with
// GetExecutor and ContextExecutor
func SetContextExecutor(ctx context.Context, name string, executor Executor) context.Context {
	return context.WithValue(ctx, executorKey{name: name}, executor)
}

// getExecutor only returns the executor in context
func getExecutor(ctx context.Context, name string) Executor {
	executor, ok := ctx.Value(executorKey{name: name}).(Executor)
	if !ok {
		return sequentialExecutor{}
	}
	return executor
}
