package sync

import (
	"context"
	"sync"
	"sync/atomic"
)

type unboundedExecutor struct {
	canceled atomic.Bool
	wg       sync.WaitGroup
}

func (e *unboundedExecutor) Execute(f func()) {
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		if e.canceled.Load() {
			return
		}
		f()
	}()
}

func (e *unboundedExecutor) Wait(ctx context.Context) {
	e.canceled.Store(ctx.Err() != nil)

	done := make(chan struct{}, 1)
	go func() {
		e.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		e.canceled.Store(true)
	case <-done:
	}
}
