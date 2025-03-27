package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_defaultExecutor(t *testing.T) {
	t.Run("only default executor", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), ExecutorDefault, &unboundedExecutor{})

		e := ContextExecutor(&ctx, "cpu")
		require.IsType(t, &unboundedExecutor{}, e)
	})

	t.Run("default executor with named", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), ExecutorDefault, &unboundedExecutor{})
		ctx = SetContextExecutor(ctx, "cpu", &queuedExecutor{})

		e := ContextExecutor(&ctx, "cpu")
		require.IsType(t, &queuedExecutor{}, e)
	})

	t.Run("no default executor with named", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), "cpu", &queuedExecutor{})

		e := ContextExecutor(&ctx, "cpu")
		require.IsType(t, &queuedExecutor{}, e)
	})

	t.Run("no default executor with different named", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), "cpu", &queuedExecutor{})

		e := ContextExecutor(&ctx, "io")
		require.IsType(t, serialExecutor{}, e)
	})

	t.Run("no executor", func(t *testing.T) {
		ctx := context.Background()

		e := ContextExecutor(&ctx, "io")
		require.IsType(t, serialExecutor{}, e)
	})

	t.Run("no executor get default", func(t *testing.T) {
		ctx := context.Background()

		e := ContextExecutor(&ctx, ExecutorDefault)
		require.IsType(t, serialExecutor{}, e)
	})

	t.Run("no context", func(t *testing.T) {
		e := ContextExecutor(nil, "cpu")
		require.IsType(t, serialExecutor{}, e)
	})

	t.Run("no context typed nil", func(t *testing.T) {
		var ctx context.Context

		e := ContextExecutor(&ctx, "cpu")
		require.IsType(t, serialExecutor{}, e)
	})
}

func Test_HasContextExecutor(t *testing.T) {
	t.Run("WithExecutorInContext", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), "cpu", &unboundedExecutor{})

		require.True(t, HasContextExecutor(ctx, "cpu"))
	})

	t.Run("WithOtherExecutorInContext", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), "cpu", &unboundedExecutor{})

		require.False(t, HasContextExecutor(ctx, "io"))
	})
}

func Test_ContextExecutor(t *testing.T) {
	t.Run("WithExecutorInContext", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), "cpu", &unboundedExecutor{})

		result := ContextExecutor(&ctx, "cpu")

		require.NotNil(t, result)
		require.IsType(t, &unboundedExecutor{}, result)
	})

	t.Run("WithoutExecutorInContext", func(t *testing.T) {
		ctx := context.Background()

		result := ContextExecutor(&ctx, "cpu")

		require.NotNil(t, result)
		require.IsType(t, serialExecutor{}, result)
	})

	t.Run("WithDifferentExecutorInContext", func(t *testing.T) {
		ctx := SetContextExecutor(context.Background(), "io", &unboundedExecutor{})

		result := ContextExecutor(&ctx, "cpu")

		require.NotNil(t, result)
		require.IsType(t, serialExecutor{}, result)
	})
}
