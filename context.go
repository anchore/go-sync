package sync

import "context"

type executorKey struct{}

func ContextExecutor(ctx context.Context) Executor {
	executor, ok := ctx.Value(executorKey{}).(Executor)
	if !ok {
		return sequentialExecutor{}
	}
	return executor
}

func SetContextExecutor(ctx context.Context, executor Executor) context.Context {
	return context.WithValue(ctx, executorKey{}, executor)
}
