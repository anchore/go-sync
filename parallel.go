package sync

import "context"

func Parallel(ctx *context.Context, executorName string, funcs ...func() error) error {
	return Collect[func() error, struct{}](ctx, executorName, ToSeq(funcs), func(f func() error) (struct{}, error) {
		return struct{}{}, f()
	}, nil)
}
