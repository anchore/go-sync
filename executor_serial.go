package sync

import "context"

type serialExecutor struct{}

func (u serialExecutor) Execute(fn func()) {
	fn()
}

func (u serialExecutor) Wait(_ context.Context) {
}

var _ Executor = (*serialExecutor)(nil)
