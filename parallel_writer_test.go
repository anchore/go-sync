package sync

import (
	"bytes"
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-sync/internal/stats"
)

func Test_ParallelWriter(t *testing.T) {
	tests := []struct {
		name           string
		maxConcurrency int
		bufferSize     int
	}{
		{
			name:           "unbounded concurrency",
			maxConcurrency: 0,
			bufferSize:     4,
		},
		{
			name:           "single execution",
			maxConcurrency: 1,
			bufferSize:     100,
		},
		{
			name:           "dual execution",
			maxConcurrency: 2,
			bufferSize:     4,
		},
		{
			name:           "ten-x execution",
			maxConcurrency: 10,
			bufferSize:     4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executed := atomic.Int32{}
			concurrency := stats.Tracked[int]{}

			buf1 := &bytes.Buffer{}
			w1 := funcWriter{
				fn: func(contents []byte) (int, error) {
					defer concurrency.Incr()()
					executed.Add(1)
					buf1.Write(contents)
					return len(contents), nil
				},
			}

			buf2 := &bytes.Buffer{}
			w2 := funcWriter{
				fn: func(contents []byte) (int, error) {
					defer concurrency.Incr()()
					executed.Add(1)
					buf2.Write(contents)
					return len(contents), nil
				},
			}

			buf3 := &bytes.Buffer{}
			w3 := funcWriter{
				fn: func(contents []byte) (int, error) {
					defer concurrency.Incr()()
					executed.Add(1)
					buf3.Write(contents)
					return len(contents), nil
				},
			}

			contents := "some complicated contents"

			ctx := SetContextExecutor(context.Background(), "", NewExecutor(test.maxConcurrency))
			w := ParallelWriter(ctx, "", w1, w2, w3)

			iterations := 0
			for i := 0; i < len(contents); i += test.bufferSize {
				iterations++
				end := i + test.bufferSize
				if end > len(contents) {
					end = len(contents)
				}
				buf := contents[i:end]
				n, err := w.Write([]byte(buf))
				require.NoError(t, err)
				require.Equal(t, len(buf), n)
			}

			require.Equal(t, 3*iterations, int(executed.Load()))
			if test.maxConcurrency > 0 {
				require.LessOrEqual(t, concurrency.Max(), test.maxConcurrency)
			} else {
				require.GreaterOrEqual(t, concurrency.Max(), 1)
			}

			require.Equal(t, contents, buf1.String())
			require.Equal(t, contents, buf2.String())
			require.Equal(t, contents, buf3.String())
		})
	}
}

type funcWriter struct {
	fn func([]byte) (int, error)
}

func (f funcWriter) Write(p []byte) (int, error) {
	return f.fn(p)
}
