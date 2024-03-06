package sync

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-sync/internal/stats"
)

func Test_Collectors(t *testing.T) {
	for i := 0; i < 100; i++ {
		Test_Collector(t)
	}
}

func Test_Collector(t *testing.T) {
	const count = 20
	const maxConcurrency = 5
	c := NewCollector[int](NewExecutor(maxConcurrency))

	concurrency := stats.Tracked[int]{}

	for i := 0; i < count; i++ {
		i := i
		c.Provide(func() int {
			defer concurrency.Incr()()

			time.Sleep(1 * time.Millisecond)

			return i
		})
	}

	values := c.Collect()
	require.Len(t, values, count)
	for i := 0; i < count; i++ {
		require.Contains(t, values, i)
	}

	require.LessOrEqual(t, concurrency.Max(), maxConcurrency)
}

func Test_lotsaLotsaCollector(t *testing.T) {
	concurrency := 100
	executors := 1000

	concurrent := stats.Tracked[int64]{}

	c := NewCollector[int](NewExecutor(concurrency))
	for i := 0; i < executors; i++ {
		i := i
		c.Provide(func() int {
			defer concurrent.Incr()()
			Test_collector(t)
			return i
		})
	}

	got := c.Collect()
	require.Len(t, got, executors)
	t.Logf("max concurrent: %d", concurrent.Max())
}

func Test_lotsaCollector(t *testing.T) {
	for i := 0; i < 1000; i++ {
		Test_collector(t)
	}
}

func Test_collector(t *testing.T) {
	concurrency := (25)
	count := 1000

	wgs := make([]sync.WaitGroup, count)

	concurrent := stats.Tracked[int64]{}
	total := atomic.Uint64{}

	makeFunc := func(idx int) func() int {
		wgs[idx].Add(1)
		return func() int {
			defer total.Add(1)
			concurrent.Incr()
			wgs[idx].Wait()
			concurrent.Decr()
			return idx
		}
	}

	var expected []int

	c := NewCollector[int](NewExecutor(concurrency))
	for i := 0; i < count; i++ {
		expected = append(expected, i)
		c.Provide(makeFunc(i))
	}

	go func() {
		for i := 0; i < count; i++ {
			wgs[i].Done()
		}
	}()

	got := c.Collect()
	require.ElementsMatch(t, expected, got)

	require.LessOrEqual(t, concurrent.Max(), int64(concurrency))
	require.Equal(t, total.Load(), uint64(count))
}

func Test_collectorSmall(t *testing.T) {
	concurrency := (2)
	count := 4

	wgs := make([]sync.WaitGroup, count)
	waiting := sync.WaitGroup{}
	waiting.Add(int(concurrency))

	concurrent := stats.Tracked[int]{}
	total := atomic.Uint64{}

	makeFunc := func(idx int) func() int {
		wgs[idx].Add(1)
		return func() int {
			if idx < int(concurrency) {
				waiting.Done()
			}
			defer total.Add(1)
			concurrent.Incr()
			wgs[idx].Wait()
			concurrent.Decr()
			return idx
		}
	}

	var expected []int

	c := NewCollector[int](NewExecutor(concurrency))
	for i := 0; i < count; i++ {
		expected = append(expected, i)
		c.Provide(makeFunc(i))
	}

	time.Sleep(10 * time.Millisecond)
	waiting.Wait()
	require.Equal(t, 2, concurrent.Val())

	go func() {
		for i := 0; i < count; i++ {
			wgs[i].Done()
		}
	}()

	got := c.Collect()
	require.ElementsMatch(t, expected, got)

	require.LessOrEqual(t, concurrent.Max(), concurrency)
	require.Equal(t, total.Load(), uint64(count))
}

func Test_collectorLimiting(t *testing.T) {
	c := NewCollector[string](NewExecutor(2))

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

	c.Provide(func() string {
		order.Add("pre wg1")
		wgReady.Done()
		wg1.Wait()
		order.Add("post wg1")
		executed += "1_"
		return "1"
	})

	c.Provide(func() string {
		order.Add("pre wg2")
		wgReady.Done()
		wg2.Wait()
		order.Add("post wg2")
		executed += "2_"
		wg3.Done()
		return "2"
	})

	wgReady.Wait()

	c.Provide(func() string {
		order.Add("pre wg3")
		wg3.Wait()
		order.Add("post wg3")
		executed += "3_"
		wg1.Done()
		return "3"
	})

	wg2.Done()

	got := c.Collect()
	require.Equal(t, "2_3_1_", executed)
	require.ElementsMatch(t, []string{"2", "3", "1"}, got)
	require.True(t,
		order.indexOf("post wg2") < order.indexOf("post wg3") &&
			order.indexOf("post wg3") < order.indexOf("post wg1"),
	)
}
