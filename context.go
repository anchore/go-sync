package sync

import (
	"context"
)

type ExecutorName string

type executorKey struct {
	name ExecutorName
}

// ContextExecutor returns an executor and a child context which prevents deadlock
// in the case that child processes use the same bounded executor
func ContextExecutor(ctx context.Context, name ExecutorName) (context.Context, Executor) {
	executor := GetExecutor(ctx, name)
	child := executor.ChildExecutor()
	ctx = SetContextExecutor(ctx, name, child)
	return ctx, executor
}

// GetExecutor returns an executor in context with the given name, or a serial executor if none exists
func GetExecutor(ctx context.Context, name ExecutorName) Executor {
	executor, ok := ctx.Value(executorKey{name: name}).(Executor)
	if !ok {
		return sequentialExecutor{}
	}
	return executor
}

// SetContextExecutor returns a context with the named provided executor for use with
// GetExecutor and ContextExecutor
func SetContextExecutor(ctx context.Context, name ExecutorName, executor Executor) context.Context {
	return context.WithValue(ctx, executorKey{name: name}, executor)
}
