package sync

import (
	"context"
)

type executorKey struct {
	name string
}

// HasContextExecutor returns true when the named executor is available in the context
func HasContextExecutor(ctx context.Context, name string) bool {
	return ctx.Value(executorKey{name: name}) != nil
}

// ContextExecutor returns an executor in context with the given name, or a serial executor if none exists
// and replaces the context with one that contains a new executor which won't deadlock if the context
func ContextExecutor(ctx *context.Context, name string) Executor {
	if ctx == nil {
		return serialExecutor{}
	}
	executor, ok := (*ctx).Value(executorKey{name: name}).(Executor)
	if !ok || executor == nil {
		return serialExecutor{}
	}
	if e, _ := executor.(ChildExecutor); e != nil {
		*ctx = SetContextExecutor(*ctx, name, e.ChildExecutor())
	}
	return executor
}

// SetContextExecutor returns a context with the named executor for use with GetExecutor
func SetContextExecutor(ctx context.Context, name string, executor Executor) context.Context {
	return context.WithValue(ctx, executorKey{name: name}, executor)
}
