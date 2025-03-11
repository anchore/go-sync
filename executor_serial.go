package sync

import "context"

type sequentialExecutor struct {
	name string
}

func (u sequentialExecutor) Name() string {
	return u.name
}

func (u sequentialExecutor) Execute(ctx context.Context, fn func(context.Context)) {
	fn(ctx)
}

func (u sequentialExecutor) Wait() {}

var _ Executor = (*sequentialExecutor)(nil)
