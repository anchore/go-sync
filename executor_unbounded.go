package sync

import (
	"context"
	"sync"
)

type unboundedExecutor struct {
	name string
	wg   sync.WaitGroup
}

func (u *unboundedExecutor) Name() string {
	return u.name
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
