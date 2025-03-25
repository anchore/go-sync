package sync

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_errGroupExecutorRepeated(t *testing.T) {
	// iterating these tests many times tends to make problems apparent much more quickly,
	// when they may succeed under certain conditions
	for i := 0; i < 1000; i++ {
		Test_errGroupExecutor(t)
	}
}

func Test_errGroupExecutor(t *testing.T) {
	// this test sets up specific wait groups to ensure that the maximum concurrency is honored
	// by stepping through and holding specific locks while conditions are verified
	e := errGroupExecutor{maxConcurrency: 2}
	e.g.SetLimit(2)

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
		wg3.Done()
	})

	e.Execute(func() {
		order.Append("pre wg2")
		wgReady.Done()
		wg2.Wait()
		order.Append("post wg2")
		executed += "2_"
	})

	wgReady.Wait()

	// errgroup execution is blocking, so the next e.Execute will block, so continue on the first before we deadlock
	wg1.Done()

	e.Execute(func() {
		order.Append("pre wg3")
		wg3.Wait()
		order.Append("post wg3")
		executed += "3_"
		wg2.Done()
	})

	e.Wait(context.Background())
	require.Equal(t, "1_3_2_", executed)
	require.True(t,
		order.indexOf("post wg1") < order.indexOf("post wg3") &&
			order.indexOf("post wg3") < order.indexOf("post wg2"),
	)
}

func Test_errGroupExecutorCancelRepeat(t *testing.T) {
	for i := 0; i < 100; i++ {
		Test_errGroupExecutorCancel(t)
	}
}

func Test_errGroupExecutorCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	e := &errGroupExecutor{}
	e.g.SetLimit(2)

	wgs := [3]sync.WaitGroup{}
	wns := [3]sync.WaitGroup{}
	for i := range wgs {
		wns[i].Add(1)
		wgs[i].Add(1)
	}

	executed := [3]bool{}
	e.Execute(func() {
		t.Logf("waiting 0")
		wns[0].Done()
		wgs[0].Wait()
		t.Logf("done 0")
		executed[0] = true
	})
	e.Execute(func() {
		t.Logf("waiting 1")
		wns[1].Done()
		wgs[1].Wait()
		t.Logf("done 1")
		executed[1] = true
	})

	go func() {
		wns[0].Wait()
		wns[1].Wait()

		// 0 and 1 are currently executing, waiting
		cancel()

		wns[2].Wait()

		e.Execute(func() {
			t.Logf("waiting 2")
			wgs[2].Wait()
			t.Logf("done 2")
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

func Test_errGroupExecutorSubcontext(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx := context.TODO()
	ctx = SetContextExecutor(ctx, "", newErrGroupExecutor(1))
	ContextExecutor(&ctx, "").Execute(func() {
		// context should be replaced with a secondary executor
		ContextExecutor(&ctx, "").Execute(func() {
			// context should be replaced again with a tertiary executor
			ContextExecutor(&ctx, "").Execute(func() {
				wg.Done()
			})
		})
	})
	wg.Wait() // only done by sub-executor
}
