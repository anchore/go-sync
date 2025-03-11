package sync

import (
	"context"
	"sync"
)

type unboundedExecutor struct {
	wg sync.WaitGroup
}

func (u *unboundedExecutor) ChildExecutor() Executor {
	return u // safe for all children to use this executor
}

func (u *unboundedExecutor) Execute(ctx context.Context, f func(context.Context)) {
	u.wg.Add(1)
	go func() {
		defer u.wg.Done()
		f(ctx)
	}()
}

func (u *unboundedExecutor) Wait() {
	u.wg.Wait()
}

var _ Executor = (*unboundedExecutor)(nil)
