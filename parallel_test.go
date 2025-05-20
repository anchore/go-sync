package sync

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Parallel(t *testing.T) {
	count := atomic.Int32{}
	count.Store(0)

	wg1 := sync.WaitGroup{}
	wg1.Add(1)

	wg2 := sync.WaitGroup{}
	wg2.Add(1)

	wg3 := sync.WaitGroup{}
	wg3.Add(1)

	err1 := fmt.Errorf("error-1")
	err2 := fmt.Errorf("error-2")
	err3 := fmt.Errorf("error-3")

	order := ""

	c := SetContextExecutor(context.TODO(), ExecutorDefault, NewExecutor(-1))

	got := Parallel(
		&c,
		"",
		func() error {
			wg1.Wait()
			count.Add(1)
			order = order + "_0"
			return nil
		},
		func() error {
			wg3.Wait()
			defer wg2.Done()
			count.Add(10)
			order = order + "_1"
			return err1
		},
		func() error {
			wg2.Wait()
			defer wg1.Done()
			count.Add(100)
			order = order + "_2"
			return err2
		},
		func() error {
			defer wg3.Done()
			count.Add(1000)
			order = order + "_3"
			return err3
		},
	)
	require.Equal(t, int32(1111), count.Load())
	require.Equal(t, "_3_1_2_0", order)

	errs := got.(interface{ Unwrap() []error }).Unwrap()

	// cannot check equality to a slice with err1,2,3 because the functions above are running in parallel, for example:
	// after func()#4 returns and the `wg3.Done()` has executed, the thread could immediately pause
	// and the remaining functions execute first and err3 becomes the last in the list instead of the first
	require.Contains(t, errs, err1)
	require.Contains(t, errs, err2)
	require.Contains(t, errs, err3)
}
