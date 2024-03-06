package sync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-sync/internal/atomic"
	"github.com/anchore/go-sync/internal/stats"
)

func Test_Executors(t *testing.T) {
	for i := 0; i < 100; i++ {
		Test_Executor(t)
	}
}

func Test_Executor(t *testing.T) {
	const count = 1000

	tests := []struct {
		name           string
		maxConcurrency int
	}{
		{
			name:           "unbounded concurrency",
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := NewExecutor(test.maxConcurrency)

			executed := atomic.Int32{}
			concurrency := stats.Tracked[int]{}

			for i := 0; i < count; i++ {
				e.Execute(func() {
					defer concurrency.Incr()()
					executed.Add(1)
					time.Sleep(10 * time.Nanosecond)
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
