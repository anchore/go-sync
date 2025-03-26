package channel

import (
	"github.com/anchore/go-sync"
)

func Tee[T any](in chan T, receivers ...func(events chan T)) (writer chan T, add func(func(events chan T)) (remove func())) {
	clones := sync.List[chan T]{}
	add = func(receiver func(events chan T)) (remove func()) {
		clone := make(chan T)
		clones.Append(clone)
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
		defer func() {
			for clone := range clones.Seq {
				close(clone)
			}
		}()
		for val := range in {
			go func(val T) {
				defer func() {
					_ = recover()
				}()
				for clone := range clones.Seq {
					clone <- val
				}
			}(val)
		}
	}()
	return in, add
}
