package sync

type sequentialExecutor struct{}

func (u sequentialExecutor) Execute(fn func()) {
	fn()
}

func (u sequentialExecutor) Wait() {}

var _ Executor = (*sequentialExecutor)(nil)
