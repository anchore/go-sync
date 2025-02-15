package sync

type sequentialExecutor struct{}

func (u sequentialExecutor) ChildExecutor() Executor {
	return u // safe for all children to use this executor
}

func (u sequentialExecutor) Execute(fn func()) {
	fn()
}

func (u sequentialExecutor) Wait() {}

var _ Executor = (*sequentialExecutor)(nil)
