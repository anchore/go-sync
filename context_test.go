package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContextExecutor(t *testing.T) {
	t.Run("WithExecutorInContext", func(t *testing.T) {
		executor := NewExecutor(1)
		ctx := SetContextExecutor(context.Background(), executor)

		result := ContextExecutor(ctx)

		require.NotNil(t, result)
		require.IsType(t, &boundedExecutor{}, result)
	})

	t.Run("WithoutExecutorInContext", func(t *testing.T) {
		ctx := context.Background()

		result := ContextExecutor(ctx)

		require.NotNil(t, result)
		require.IsType(t, sequentialExecutor{}, result)
	})
}
