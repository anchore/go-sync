package atomic

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Float64(t *testing.T) {
	val := Float64{}
	val2 := &atomic.Int64{}

	concurrency := 3925
	num := 3
	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			val.Add(float64(num))
			val2.Add(int64(num))
			wg.Done()
		}()
	}
	wg.Wait()
	require.Equal(t, val2.Load(), int64(val.Load()))
	require.Equal(t, int64(concurrency*num), int64(val.Load()))
	require.Equal(t, int64(concurrency*num), val2.Load())
}
