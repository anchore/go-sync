package sync

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
)

// Executor the executor interface allows for different strategies to execute units of work and wait for all units
// of work to be completed
type Executor interface {
	// Name returns the executor's name, used to look it up in context
	Name() string

	// Execute adds a unit of work to be executed by the executor
	Execute(context.Context, func(context.Context))

	// Wait blocks and waits for all the executing functions to be completed before returning, if more functions are
	// added to be executed by this executor after the Wait call, these will also complete before Wait proceeds
	Wait()
}

// NewExecutor returns an Executor based on the desired concurrency:
//
//	< 0: unbounded, spawn a new goroutine for each Execute call
//	  0: serial, executes in the same thread/routine as the caller of Execute
//	> 0: a bounded executor with the maximum concurrency provided
func NewExecutor(name string, maxConcurrency int) Executor {
	if maxConcurrency < 0 || maxConcurrency > math.MaxInt32 {
		return &unboundedExecutor{
			name: name,
		}
	}
	if maxConcurrency == 0 {
		return sequentialExecutor{
			name: name,
		}
	}
	return &boundedExecutor{
		name:          name,
		maxConcurrent: int32(maxConcurrency),
	}
}

type boundedExecutor struct {
	name          string // name used in context
	maxConcurrent int32
	executing     atomic.Int32
	queue         List[*func()]
	wg            sync.WaitGroup
	childLock     sync.Mutex
	childExecutor *boundedExecutor
}

func (e *boundedExecutor) Name() string {
	return e.name
}

func (e *boundedExecutor) Execute(ctx context.Context, f func(context.Context)) {
	ctx = SetContextExecutor(ctx, e.getChildExecutor())
	e.wg.Add(1)
	fn := func() {
		defer e.wg.Done()
		f(ctx)
	}
	e.queue.Enqueue(&fn)
	if e.executing.Load() < e.maxConcurrent {
		e.executing.Add(1)
		go e.exec()
	}
}

func (e *boundedExecutor) Wait() {
	e.wg.Wait()
}

func (e *boundedExecutor) exec() {
	defer e.executing.Add(-1)
	if e.executing.Load() > e.maxConcurrent {
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

func (e *boundedExecutor) getChildExecutor() Executor {
	e.childLock.Lock()
	defer e.childLock.Unlock()
	// create a child executor with the same bound
	if e.childExecutor == nil {
		e.childExecutor = &boundedExecutor{
			name:          e.name,
			maxConcurrent: e.maxConcurrent,
		}
	}
	return e.childExecutor
}

var _ Executor = (*boundedExecutor)(nil)
