package sync

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-sync/internal/atomic"
	"github.com/anchore/go-sync/internal/stats"
)

func Test_callstackAwareExecutor(t *testing.T) {
	const sleepTime = 1 * time.Millisecond
	const count = 100

	tests := []struct {
		name           string
		maxConcurrency int
	}{
		{
			name:           "serial execution",
			maxConcurrency: 0,
		},
		{
			name:           "single execution",
			maxConcurrency: 1,
		},
		{
			name:           "dual execution",
			maxConcurrency: 2,
		},
		{
			name:           "ten-x execution",
			maxConcurrency: 10,
		},
		{
			name:           "unbounded concurrency",
			maxConcurrency: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewExecutor(test.maxConcurrency)
			e = NewTrackedExecutor(e, func(f func()) { f() })

			executed := atomic.Int32{}
			concurrency := stats.Tracked[int]{}

			for i := 0; i < count; i++ {
				e.Execute(func() {
					defer concurrency.Incr()()
					executed.Add(1)
					executedSerially := false
					time.Sleep(sleepTime)
					e.Execute(func() {
						time.Sleep(sleepTime)
						// this sub-call should be handled inline
						executedSerially = true
					})
					require.True(t, executedSerially)
				})
			}

			e.Wait()

			require.Equal(t, count, int(executed.Load()))
			if test.maxConcurrency > 0 {
				require.LessOrEqual(t, concurrency.Max(), test.maxConcurrency)
			} else {
				require.GreaterOrEqual(t, concurrency.Max(), 1)
			}
		})
	}
}

func Test_TrackedExecutor(t *testing.T) {
	e := NewExecutor(1)
	e1 := NewTrackedExecutor(e, func(f func()) { f() })

	wg := sync.WaitGroup{}
	wg.Add(1)
	e1.Execute(func() {
		innerExecuted := false
		// would cause deadlock if not executed serially
		e1.Execute(func() {
			innerExecuted = true
		})
		require.True(t, innerExecuted)
		wg.Done()
	})
	wg.Wait()
}

func Test_uniqueTrackedExecutors(t *testing.T) {
	e := NewExecutor(1)
	e1 := NewTrackedExecutor(e, func(f func()) { f() })
	e2 := NewTrackedExecutor(e, func(f func()) { f() })

	wg := sync.WaitGroup{}
	wg.Add(1)
	e1.Execute(func() {
		e2executed := false
		e2.Execute(func() {
			e2executed = true
		})
		require.False(t, e2executed)
		wg.Done()
	})
	wg.Wait()
}
