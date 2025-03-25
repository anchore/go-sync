package sync

import (
	"context"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type errGroupExecutor struct {
	maxConcurrency int
	canceled       atomic.Bool
	g              errgroup.Group
	wg             sync.WaitGroup
	childLock      sync.RWMutex
	childExecutor  *errGroupExecutor
}

func newErrGroupExecutor(maxConcurrency int) *errGroupExecutor {
	e := &errGroupExecutor{
		maxConcurrency: maxConcurrency,
	}
	e.g.SetLimit(maxConcurrency)
	return e
}

func (e *errGroupExecutor) Execute(f func()) {
	e.wg.Add(1)
	fn := func() error {
		defer e.wg.Done()
		if e.canceled.Load() {
			return nil
		}
		f()
		return nil
	}
	e.g.Go(fn)
}

func (e *errGroupExecutor) Wait(ctx context.Context) {
	e.canceled.Store(ctx.Err() != nil)

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		e.canceled.Store(true)

	case <-done:
	}
}

func (e *errGroupExecutor) ChildExecutor() Executor {
	e.childLock.RLock()
	defer e.childLock.RUnlock()
	if e.childExecutor == nil {
		e.childLock.RUnlock()
		e.childLock.Lock() // exclusive lock so we only create one child executor
		if e.childExecutor == nil {
			// create a child executor with the same bound
			e.childExecutor = newErrGroupExecutor(e.maxConcurrency)
		}
		e.childLock.Unlock()
		e.childLock.RLock() // needed for defer to unlock
	}
	return e.childExecutor
}

var _ Executor = (*errGroupExecutor)(nil)
