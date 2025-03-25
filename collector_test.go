package sync

import (
	"context"
	"iter"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-sync/internal/stats"
)

func Test_CollectCancelRepeat(t *testing.T) {
	// iterating these tests many times tends to make problems apparent much more quickly,
	// when they may succeed under certain conditions
	for i := 0; i < 1000; i++ {
		Test_CollectCancel(t)
	}
}

func Test_CollectCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	e := &errGroupExecutor{} // use errgroup executor as it will block before executing 3
	e.g.SetLimit(2)
	ctx = SetContextExecutor(ctx, "", e)

	executed3 := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	err := Collect(&ctx, "", ToSeq([]int{1, 2, 3}), func(i int) (string, error) {
		switch i {
		case 1:
			// cancel
			cancel()
			// ensure 2 doesn't block
			wg.Done()
		case 2:
			// ensure only 1 and 2 execute by waiting here
			wg.Wait()
		case 3:
			executed3 = true
		}
		return "", nil
	}, func(i int, s string) {})

	// should not have an error, even though context was canceled
	require.NoError(t, err)

	// should not have executed 3
	require.False(t, executed3)
}

func Test_CollectSlice(t *testing.T) {
	const count = 1000
	const maxConcurrency = 5

	concurrency := stats.Tracked[int]{}

	var values []int
	ctx := SetContextExecutor(context.Background(), "", NewExecutor(maxConcurrency))
	err := CollectSlice(&ctx, "", countIter(count), func(i int) (int, error) {
		defer concurrency.Incr()()

		time.Sleep(1 * time.Millisecond)

		return i * 10, nil
	}, &values)
	require.NoError(t, err)

	require.Len(t, values, count)
	for i := 0; i < count; i++ {
		require.Contains(t, values, i*10)
	}

	require.LessOrEqual(t, concurrency.Max(), maxConcurrency)
}

func Test_CollectMap(t *testing.T) {
	const count = 1000
	const maxConcurrency = 5

	concurrency := stats.Tracked[int]{}

	values := map[int]int{}
	ctx := SetContextExecutor(context.Background(), "", NewExecutor(maxConcurrency))
	err := CollectMap(&ctx, "", countIter(count), func(i int) (int, error) {
		defer concurrency.Incr()()

		time.Sleep(1 * time.Millisecond)

		return i * 10, nil
	}, values)
	require.NoError(t, err)

	require.Len(t, values, count)
	for i := 0; i < count; i++ {
		require.Equal(t, values[i], i*10)
	}

	require.LessOrEqual(t, concurrency.Max(), maxConcurrency)
}

func Test_ToSeqToSlice(t *testing.T) {
	expected := []int{0, 1, 2, 3, 4}

	seq := ToSeq(expected)

	got := ToSlice(seq)

	require.EqualValues(t, expected, got)
}

func countIter(count int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < count; i++ {
			if !yield(i) {
				return
			}
		}
	}
}
