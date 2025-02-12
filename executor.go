package sync

import (
	"math"
	"sync"
	"sync/atomic"
)

// Executor the executor interface allows for different strategies to execute units of work and wait for all units
// of work to be completed
type Executor interface {
	// Execute adds a unit of work to be executed by the executor
	Execute(func())

	// Wait blocks and waits for all the executing functions to be completed before returning, if more functions are
	// added to be executed by this executor after the Wait call, these will also complete before Wait proceeds
	Wait()
}

// NewExecutor returns an Executor based on the desired concurrency:
//
//	< 0: unbounded, spawn a new goroutine for each Execute call
//	  0: serial, executes in the same thread/routine as the caller of Execute
//	> 0: a bounded executor with the maximum concurrency provided
func NewExecutor(maxConcurrency int) Executor {
	if maxConcurrency < 0 || maxConcurrency > math.MaxInt32 {
		return &unboundedExecutor{}
	}
	if maxConcurrency == 0 {
		return sequentialExecutor{}
	}
	return &boundedExecutor{
		maxConcurrent: int32(maxConcurrency),
	}
}

type boundedExecutor struct {
	maxConcurrent int32
	executing     atomic.Int32
	queue         List[*func()]
	wg            sync.WaitGroup
}

var _ Executor = (*boundedExecutor)(nil)

func (e *boundedExecutor) Execute(f func()) {
	e.wg.Add(1)
	fn := func() {
		defer e.wg.Done()
		f()
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
