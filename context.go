package sync

import (
	"context"
)

type ExecutorName string

type executorKey struct {
	name ExecutorName
}

func ContextExecutor(ctx context.Context, name ExecutorName) Executor {
	executor, ok := ctx.Value(executorKey{name: name}).(Executor)
	if !ok {
		return sequentialExecutor{}
	}
	return executor
}

func SetContextExecutor(ctx context.Context, name ExecutorName, executor Executor) context.Context {
	return context.WithValue(ctx, executorKey{name: name}, executor)
}
