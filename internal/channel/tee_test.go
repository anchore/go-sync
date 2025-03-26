package channel

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	gosync "github.com/anchore/go-sync"
)

func Test_ChannelTee(t *testing.T) {
	events := make(chan any)

	received := &gosync.List[int]{}
	wg := &sync.WaitGroup{}
	closing := atomic.Bool{}
	closing.Store(false)

	makeReceiver := func(number int) func(events chan any) {
		return func(events chan any) {
			for {
				select {
				case event, open := <-events:
					if event != nil {
						// t.Logf("%d: %v", number, event)
						received.Append(number)
						wg.Done()
					}
					if !open {
						if closing.Load() {
							wg.Done()
						}
						// t.Logf("end %d", number)
						return
					}
				}
			}
		}
	}

	events, add := Tee(events, makeReceiver(1), makeReceiver(2))
	remove3 := add(makeReceiver(3))
	_ = add(makeReceiver(4))
	remove5 := add(makeReceiver(5))
	_ = add(makeReceiver(6))
	remove7 := add(makeReceiver(7))

	wg.Add(7)
	go func() {
		events <- fmt.Errorf("hello")
	}()
	wg.Wait()

	require.ElementsMatch(t, received.Values(), []int{1, 2, 3, 4, 5, 6, 7})

	remove5()
	received.Clear()

	wg.Add(6)
	go func() {
		events <- fmt.Errorf("goodbye")
	}()
	wg.Wait()

	require.ElementsMatch(t, received.Values(), []int{1, 2, 3, 4, 6, 7})

	remove3()
	received.Clear()

	wg.Add(5)

	go func() {
		events <- fmt.Errorf("cats")
	}()

	wg.Wait()

	require.ElementsMatch(t, received.Values(), []int{1, 2, 4, 6, 7})

	remove7()
	received.Clear()

	wg.Add(4)

	go func() {
		events <- fmt.Errorf("dogs")
	}()

	wg.Wait()
	require.ElementsMatch(t, received.Values(), []int{1, 2, 4, 6})

	closing.Store(true)
	wg.Add(4)
	close(events)
	wg.Wait()
}
