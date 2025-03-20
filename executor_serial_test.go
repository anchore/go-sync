package sync

import (
	"context"
	"sync"
	"testing"
)

func Test_serialSubcontext(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx := context.TODO()
	ctx = SetContextExecutor(ctx, "", serialExecutor{})
	ContextExecutor(&ctx, "").Execute(func() {
		// context should be able to continue
		ContextExecutor(&ctx, "").Execute(func() {
			// context should be able to continue
			ContextExecutor(&ctx, "").Execute(func() {
				wg.Done()
			})
		})
	})
	wg.Wait() // only done by sub-executor
}
