package sync

import (
	"sync"
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

func Test_Executor2(t *testing.T) {
	concurrency := 25
	count := 1000

	wgs := make([]sync.WaitGroup, count)

	concurrent := stats.Tracked[int64]{}
	total := atomic.Uint64{}

	makeFunc := func(idx int) func() {
		wgs[idx].Add(1)
		return func() {
			defer total.Add(1)
			concurrent.Incr()
			wgs[idx].Wait()
			concurrent.Decr()
		}
	}

	var expected []int

	e := NewExecutor(concurrency)
	for i := 0; i < count; i++ {
		expected = append(expected, i)
		e.Execute(makeFunc(i))
	}

	go func() {
		for i := 0; i < count; i++ {
			wgs[i].Done()
		}
	}()

	e.Wait()
	require.LessOrEqual(t, concurrent.Max(), int64(concurrency))
	require.Equal(t, total.Load(), uint64(count))
}

func Test_executorSmall(t *testing.T) {
	concurrency := 2
	count := 4

	wgs := make([]sync.WaitGroup, count)
	waiting := sync.WaitGroup{}
	waiting.Add(int(concurrency))

	concurrent := stats.Tracked[int]{}
	total := atomic.Uint64{}

	makeFunc := func(idx int) func() {
		wgs[idx].Add(1)
		return func() {
			if idx < int(concurrency) {
				waiting.Done()
			}
			defer total.Add(1)
			concurrent.Incr()
			wgs[idx].Wait()
			concurrent.Decr()
		}
	}

	e := NewExecutor(concurrency)
	for i := 0; i < count; i++ {
		e.Execute(makeFunc(i))
	}

	time.Sleep(10 * time.Millisecond)
	waiting.Wait()
	require.Equal(t, 2, concurrent.Val())

	go func() {
		for i := 0; i < count; i++ {
			wgs[i].Done()
		}
	}()

	e.Wait()
	require.LessOrEqual(t, concurrent.Max(), concurrency)
	require.Equal(t, total.Load(), uint64(count))
}

func Test_explicitExecutorLimiting(t *testing.T) {
	// this test sets up specific wait groups to ensure that the maximum concurrency is honored
	// by stepping through and holding specific locks while conditions are verified
	e := NewExecutor(2)

	wg1 := &sync.WaitGroup{}
	wg1.Add(1)
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	wg3 := &sync.WaitGroup{}
	wg3.Add(1)

	order := List[string]{}
	executed := ""

	wgReady := &sync.WaitGroup{}
	wgReady.Add(2)

	e.Execute(func() {
		order.Append("pre wg1")
		wgReady.Done()
		wg1.Wait()
		order.Append("post wg1")
		executed += "1_"
	})

	e.Execute(func() {
		order.Append("pre wg2")
		wgReady.Done()
		wg2.Wait()
		order.Append("post wg2")
		executed += "2_"
		wg3.Done()
	})

	wgReady.Wait()

	e.Execute(func() {
		order.Append("pre wg3")
		wg3.Wait()
		order.Append("post wg3")
		executed += "3_"
		wg1.Done()
	})

	wg2.Done()

	e.Wait()
	require.Equal(t, "2_3_1_", executed)
	require.True(t,
		order.indexOf("post wg2") < order.indexOf("post wg3") &&
			order.indexOf("post wg3") < order.indexOf("post wg1"),
	)
}
