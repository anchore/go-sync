package sync

import (
	"sync"
)

type unboundedExecutor struct {
	wg sync.WaitGroup
}

func (u *unboundedExecutor) Execute(f func()) {
	u.wg.Add(1)
	go func() {
		defer u.wg.Done()
		f()
	}()
}

func (u *unboundedExecutor) Wait() {
	u.wg.Wait()
}

var _ Executor = (*unboundedExecutor)(nil)
