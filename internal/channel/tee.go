package channel

import (
	"github.com/anchore/go-sync"
)

func Tee[T any](in chan T, receivers ...func(events chan T)) (writer chan T, add func(func(events chan T)) (remove func())) {
	clones := sync.List[chan T]{}
	add = func(receiver func(events chan T)) (remove func()) {
		clone := make(chan T)
		clones.Add(clone)
		go receiver(clone)
		return func() {
			clones.Remove(clone)
			close(clone)
		}
	}
	for _, receiver := range receivers {
		_ = add(receiver)
	}
	go func() {
		defer clones.Each(func(clone chan T) {
			close(clone)
		})
		for val := range in {
			go func(val T) {
				defer func() {
					_ = recover()
				}()
				defer clones.RLock()()
				clones.Each(func(clone chan T) {
					clone <- val
				})
			}(val)
		}
	}()
	return in, add
}
