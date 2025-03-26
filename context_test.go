package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

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
