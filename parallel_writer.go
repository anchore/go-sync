package sync

import (
	"errors"
	"io"
	"sync"
)

type parallelWriter struct {
	executor Executor
	writers  []io.Writer
}

func ParallelWriter(executor Executor, writers ...io.Writer) io.Writer {
	return &parallelWriter{
		executor: executor,
		writers:  writers,
	}
}

func (w *parallelWriter) Write(p []byte) (int, error) {
	errs := List[error]{}
	wg := sync.WaitGroup{}
	wg.Add(len(w.writers))
	for _, writer := range w.writers {
		w.executor.Execute(func() {
			defer wg.Done()
			_, err := writer.Write(p)
			if err != nil {
				errs.Append(err)
			}
		})
	}
	wg.Wait()
	if errs.Len() > 0 {
		return 0, errors.Join(errs.Values()...)
	}
	return len(p), nil
}

var _ io.Writer = (*parallelWriter)(nil)
