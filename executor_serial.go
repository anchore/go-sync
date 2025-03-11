package sync

import "context"

type sequentialExecutor struct{}

func (u sequentialExecutor) ChildExecutor() Executor {
	return u // safe for all children to use this executor
}

func (u sequentialExecutor) Execute(ctx context.Context, fn func(context.Context)) {
	fn(ctx)
}

func (u sequentialExecutor) Wait() {}

var _ Executor = (*sequentialExecutor)(nil)
