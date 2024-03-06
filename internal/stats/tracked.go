package stats

import (
	"github.com/anchore/go-sync/internal/atomic"
)

type Tracked[T int | int32 | int64] struct {
	count atomic.Int64
	val   atomic.Int64
	max   atomic.Int64
	min   atomic.Int64
}

func (t *Tracked[T]) Add(val T) {
	// each time a value is added, track a count so we can provide an average
	t.count.Add(1)

	c := t.val.Add(int64(val))

	m := t.max.Load()
	for c > m {
		if !t.max.CompareAndSwap(m, c) {
			m = t.max.Load()
			continue
		}
		break
	}

	m = t.min.Load()
	for c < m {
		if !t.min.CompareAndSwap(m, c) {
			m = t.min.Load()
			continue
		}
		break
	}
}

func (t *Tracked[T]) Incr() func() {
	t.Add(1)
	return t.Decr
}

func (t *Tracked[T]) Decr() {
	t.Add(-1)
}

func (t *Tracked[T]) Val() T {
	return T(t.val.Load())
}

func (t *Tracked[T]) Max() T {
	return T(t.max.Load())
}

func (t *Tracked[T]) Min() T {
	return T(t.min.Load())
}

func (t *Tracked[T]) Avg() float64 {
	return float64(t.val.Load()) / float64(t.count.Load())
}

// --- TrackedFloat ---

type TrackedFloat[T int | int32 | int64 | float32 | float64] struct {
	count  atomic.Float64
	val    atomic.Float64
	max    atomic.Float64
	min    atomic.Float64
	maxSet func(val T)
}

func (t *TrackedFloat[T]) Add(val T) {
	// each time a value is added, track a count so we can provide an average
	t.count.Add(1)

	c := t.val.Add(float64(val))

	m := t.max.Load()
	for c > m {
		if !t.max.CompareAndSwap(m, c) {
			m = t.max.Load()
			continue
		}
		if t.maxSet != nil {
			t.maxSet(T(c))
		}
		break
	}

	m = t.min.Load()
	for c < m {
		if !t.min.CompareAndSwap(m, c) {
			m = t.min.Load()
			continue
		}
		break
	}
}

func (t *TrackedFloat[T]) Incr() func() {
	t.Add(1)
	return t.Decr
}

func (t *TrackedFloat[T]) Decr() {
	t.Add(-1)
}

func (t *TrackedFloat[T]) Val() T {
	return T(t.val.Load())
}

func (t *TrackedFloat[T]) Max() T {
	return T(t.max.Load())
}

func (t *TrackedFloat[T]) Min() T {
	return T(t.min.Load())
}

func (t *TrackedFloat[T]) Avg() float64 {
	return t.val.Load() / t.count.Load()
}

func (t *TrackedFloat[T]) OnMaxSet(f func(max T)) {
	if t.maxSet != nil {
		existing := t.maxSet
		t.maxSet = func(val T) {
			existing(val)
			f(val)
		}
		return
	}
	t.maxSet = f
}
