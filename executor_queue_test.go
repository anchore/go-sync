package sync

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-sync/internal/atomic"
	"github.com/anchore/go-sync/internal/stats"
)

func Test_queuedExecutor(t *testing.T) {
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

	e := &queuedExecutor{maxConcurrency: concurrency}

	for i := 0; i < count; i++ {
		expected = append(expected, i)
		e.Go(makeFunc(i))
	}

	go func() {
		for i := 0; i < count; i++ {
			wgs[i].Done()
		}
	}()

	e.Wait(context.Background())
	require.LessOrEqual(t, concurrent.Max(), int64(concurrency))
	require.Equal(t, total.Load(), uint64(count))
}

func Test_queuedExecutorSmall(t *testing.T) {
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

	e := &queuedExecutor{maxConcurrency: concurrency}
	for i := 0; i < count; i++ {
		e.Go(makeFunc(i))
	}

	time.Sleep(10 * time.Millisecond)
	waiting.Wait()
	require.Equal(t, 2, concurrent.Val())

	go func() {
		for i := 0; i < count; i++ {
			wgs[i].Done()
		}
	}()

	e.Wait(context.Background())
	require.LessOrEqual(t, concurrent.Max(), concurrency)
	require.Equal(t, total.Load(), uint64(count))
}

func Test_explicitExecutorLimiting(t *testing.T) {
	// this test sets up specific wait groups to ensure that the maximum concurrency is honored
	// by stepping through and holding specific locks while conditions are verified
	e := queuedExecutor{maxConcurrency: 2}

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

	e.Go(func() {
		order.Append("pre wg1")
		wgReady.Done()
		wg1.Wait()
		order.Append("post wg1")
		executed += "1_"
	})

	e.Go(func() {
		order.Append("pre wg2")
		wgReady.Done()
		wg2.Wait()
		order.Append("post wg2")
		executed += "2_"
		wg3.Done()
	})

	wgReady.Wait()

	e.Go(func() {
		order.Append("pre wg3")
		wg3.Wait()
		order.Append("post wg3")
		executed += "3_"
		wg1.Done()
	})

	wg2.Done()

	e.Wait(context.Background())
	require.Equal(t, "2_3_1_", executed)
	require.True(t,
		order.indexOf("post wg2") < order.indexOf("post wg3") &&
			order.indexOf("post wg3") < order.indexOf("post wg1"),
	)
}

func Test_queuedExecutorCancelRepeat(t *testing.T) {
	// iterating these tests many times tends to make problems apparent much more quickly,
	// when they may succeed under certain conditions
	for i := 0; i < 1000; i++ {
		Test_queuedExecutorCancel(t)
	}
}

func Test_queuedExecutorCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	e := &queuedExecutor{maxConcurrency: 2}

	wgs := [3]sync.WaitGroup{}
	wns := [3]sync.WaitGroup{}
	for i := range wgs {
		wns[i].Add(1)
		wgs[i].Add(1)
	}

	executed := [3]bool{}
	e.Go(func() {
		wns[0].Done()
		wgs[0].Wait()
		executed[0] = true
	})
	e.Go(func() {
		wns[1].Done()
		wgs[1].Wait()
		executed[1] = true
	})

	go func() {
		wns[0].Wait()
		wns[1].Wait()

		// 0 and 1 are currently executing, waiting
		cancel()

		wns[2].Wait()

		e.Go(func() {
			wgs[2].Wait()
			executed[2] = true
		})

		for i := range wgs {
			wgs[i].Done()
		}

	}()

	// should be waiting in 0, 1 not executed 2
	e.Wait(ctx)

	wns[2].Done()

	// should not have executed 2
	require.False(t, executed[2])
}

func Test_queuedExecutorSubcontext(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx := context.TODO()
	ctx = SetContextExecutor(ctx, "", &queuedExecutor{maxConcurrency: 1})
	ContextExecutor(&ctx, "").Go(func() {
		// context should be replaced with a secondary executor
		ContextExecutor(&ctx, "").Go(func() {
			// context should be replaced again with a tertiary executor
			ContextExecutor(&ctx, "").Go(func() {
				wg.Done()
			})
		})
	})
	wg.Wait() // only done by sub-executor
}
