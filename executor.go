package sync

import (
	"context"
	"math"
)

// Executor the executor interface allows for different strategies to execute units of work and wait for all units
// of work to be completed
type Executor interface {
	// Execute adds a unit of work to be executed by the executor
	Execute(func())

	// Wait blocks and waits for all the executing functions to be completed before returning, or the context is cancelled.
	// if more functions are added to be executed by this executor after the Wait call, these will also complete before Wait proceeds
	// If the context is canceled, any queued functions will not be executed
	Wait(context.Context)
}

// ChildExecutor interface, if implemented, will cause GetContextExecutor calls to replace the provided context
// with a child executor returned from this function. This is used when it is not safe to nest Execute calls
type ChildExecutor interface {
	ChildExecutor() Executor
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
		return serialExecutor{}
	}
	if useErrGroup {
		return newErrGroupExecutor(maxConcurrency)
	}
	e := &queuedExecutor{
		maxConcurrency: maxConcurrency,
	}
	return e
}

var useErrGroup = true
