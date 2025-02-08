package sync

import (
	"runtime"
)

type callstackAwareExecutor struct {
	buffers  List[*[]uintptr]
	maxStack int
	marker   func(func())
	markerPC uintptr
	executor Executor
}

// NewTrackedExecutor returns an executor that is aware of _itself_, such that
// if it is already executing something in parallel, it will execute subsequent calls in the same
// callstack serially.
func NewTrackedExecutor(executor Executor, maxStack int, markFunc func(func())) Executor {
	e := &callstackAwareExecutor{
		maxStack: maxStack,
		marker:   markFunc,
		executor: executor,
	}

	// figure out the program counter of the provided markFunc
	var pc uintptr
	markFunc(func() {
		buf := make([]uintptr, e.maxStack)
		defer e.buffers.Push(&buf)
		_ = runtime.Callers(2, buf)
		frames := runtime.CallersFrames(buf[:1])
		frame, _ := frames.Next()
		pc = frame.PC
	})
	e.markerPC = pc
	return e
}

var _ Executor = (*callstackAwareExecutor)(nil)

func (e *callstackAwareExecutor) Execute(fn func()) {
	pcPtr, ok := e.buffers.Pop()
	if !ok {
		buf := make([]uintptr, e.maxStack)
		pcPtr = &buf
	}
	defer e.buffers.Push(pcPtr)
	callers := *pcPtr

	inExecutor := false
	// 3 here means: skip runtime.Callers, this invocation, and prior caller which won't be the mark
	count := runtime.Callers(3, callers)
	frames := runtime.CallersFrames(callers[:count])

	for {
		frame, more := frames.Next()
		if frame.PC == e.markerPC {
			inExecutor = true
			break
		}
		if !more {
			break
		}
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
