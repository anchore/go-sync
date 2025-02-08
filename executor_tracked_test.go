package sync

import (
	"github.com/anchore/go-sync/internal/atomic"
	"github.com/anchore/go-sync/internal/stats"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
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
			e = NewTrackedExecutor(e, 10, func(f func()) { f() })

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

func Test_concurrentCallstackAwareExecutors(t *testing.T) {
	e := NewExecutor(1)

	e1 := NewTrackedExecutor(e, 10, func(f func()) { f() })

	wg := sync.WaitGroup{}

	wg.Add(1)
	e1.Execute(func() {
		e1.Execute(func() {
			wg.Done()
		})
	})

	wg.Add(1)
	e2 := NewTrackedExecutor(e, 10, func(f func()) { f() })
	e1.Execute(func() {
		e2.Execute(func() {
			wg.Done()
		})
	})

	wg.Wait()
}
