package sync

import (
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-sync/internal/stats"
)

func Test_ReduceCollectSlice(t *testing.T) {
	const count = 1000
	const maxConcurrency = 5
	e := NewExecutor(maxConcurrency)

	concurrency := stats.Tracked[int]{}

	var values []int
	err := CollectSlice(e, countIter(count), &values, func(i int) (int, error) {
		defer concurrency.Incr()()

		time.Sleep(1 * time.Millisecond)

		return i * 10, nil
	})
	require.NoError(t, err)

	require.Len(t, values, count)
	for i := 0; i < count; i++ {
		require.Contains(t, values, i*10)
	}

	require.LessOrEqual(t, concurrency.Max(), maxConcurrency)
}

func Test_ReduceCollectMap(t *testing.T) {
	const count = 1000
	const maxConcurrency = 5
	e := NewExecutor(maxConcurrency)

	concurrency := stats.Tracked[int]{}

	values := map[int]int{}
	err := CollectMap(e, countIter(count), values, func(i int) (int, error) {
		defer concurrency.Incr()()

		time.Sleep(1 * time.Millisecond)

		return i * 10, nil
	})
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
