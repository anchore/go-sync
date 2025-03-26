package atomic

import (
	"math"
	"sync/atomic"
)

// Float64 provides a struct implementing an atomic version of a float64 value, like sync/atomic.Int64
type Float64 struct {
	value atomic.Uint64
}

// Load atomically loads and returns the value stored in x.
func (x *Float64) Load() float64 {
	return math.Float64frombits(x.value.Load())
}

// Store atomically stores val into x.
func (x *Float64) Store(val float64) {
	x.value.Store(math.Float64bits(val))
}

// Swap atomically stores new into x and returns the previous value.
func (x *Float64) Swap(newVal float64) (oldVal float64) {
	return math.Float64frombits(x.value.Swap(math.Float64bits(newVal)))
}

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Float64) CompareAndSwap(oldVal, newVal float64) (swapped bool) {
	return x.value.CompareAndSwap(math.Float64bits(oldVal), math.Float64bits(newVal))
}

// Add atomically adds delta to x and returns the new value.
func (x *Float64) Add(delta float64) (updated float64) {
	current := x.Load()
	if delta == 0 {
		return current
	}
	for {
		updated = current + delta
		if !x.CompareAndSwap(current, updated) {
			current = x.Load()
			continue
		}
		break
	}
	return
}
