package sync

import (
	"context"
	"sync"
	"sync/atomic"
)

type queuedExecutor struct {
	canceled       atomic.Bool
	maxConcurrency int
	executing      atomic.Int32
	queue          List[*func()]
	wg             sync.WaitGroup
	childLock      sync.RWMutex
	childExecutor  *errGroupExecutor
}

var _ Executor = (*queuedExecutor)(nil)

func (e *queuedExecutor) Execute(f func()) {
	if e.canceled.Load() {
		return
	}
	e.wg.Add(1)
	fn := func() {
		defer e.wg.Done()
		if e.canceled.Load() {
			return
		}
		f()
	}
	e.queue.Enqueue(&fn)
	if int(e.executing.Load()) < e.maxConcurrency {
		go e.exec()
	}
}

func (e *queuedExecutor) Wait(ctx context.Context) {
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

func (e *queuedExecutor) ChildExecutor() Executor {
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

func (e *queuedExecutor) exec() {
	e.executing.Add(1)
	defer e.executing.Add(-1)
	if int(e.executing.Load()) > e.maxConcurrency {
		return
	}
	for {
		f, ok := e.queue.Dequeue()
		if !ok {
			return
		}
		if f != nil {
			(*f)()
		}
	}
}
