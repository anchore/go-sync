package sync

import (
	"runtime"
)

type callstackAwareExecutor struct {
	marker   func(func())
	markerPC uintptr
	executor Executor
}

// NewTrackedExecutor returns an executor that is aware of _itself_, such that
// if it is already executing something in parallel, it will execute subsequent calls in the same
// callstack serially.
func NewTrackedExecutor(executor Executor, markFunc func(func())) Executor {
	e := &callstackAwareExecutor{
		marker:   markFunc,
		executor: executor,
	}

	// figure out the program counter of the provided markFunc
	var pc uintptr
	markFunc(func() {
		arr := [1]uintptr{}
		buf := arr[:]
		_ = runtime.Callers(2, buf)
		frames := runtime.CallersFrames(buf)
		frame, _ := frames.Next()
		pc = frame.PC
	})
	e.markerPC = pc
	return e
}

var _ Executor = (*callstackAwareExecutor)(nil)

func (e *callstackAwareExecutor) Execute(fn func()) {
	const bufSize = 8
	arr := [bufSize]uintptr{}
	callers := arr[:]

	inExecutor := false

	// starting with 3 here means: skip runtime.Callers, this invocation, and prior caller which won't be the mark
	start := 3
check:
	for {
		count := runtime.Callers(start, callers)
		if count == 0 {
			break
		}
		frames := runtime.CallersFrames(callers[:count])
		for {
			frame, more := frames.Next()
			if frame.PC == e.markerPC {
				inExecutor = true
				break check
			}
			if !more {
				break
			}
		}
		if count < bufSize {
			break
		}
		start += bufSize
	}
	if inExecutor {
		fn()
	} else {
		e.executor.Execute(func() {
			e.marker(fn)
		})
	}
}

func (e *callstackAwareExecutor) Wait() {
	e.executor.Wait()
}
