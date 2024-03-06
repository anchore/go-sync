package sync

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_List(t *testing.T) {
	expected := []int{0, 1, 2, 3, 4, 5}

	// as a stack:
	var sl = &List[int]{}
	var s Stack[int] = sl
	for _, e := range expected {
		s.Push(e)
	}
	require.Equal(t, expected, collect(sl))
	for i := range expected {
		ev := len(expected) - 1 - i
		exp := expected[:ev]
		v, ok := s.Pop()
		require.True(t, ok)
		require.Equal(t, expected[ev], v)
		require.Equal(t, exp, collect(sl))
	}

	// as a List:
	ls := &List[int]{}
	for _, v := range expected {
		ls.Add(v)
	}
	require.Equal(t, expected, collect(ls))
	for i, e := range expected {
		exp := expected[i+1:]
		ls.Remove(e)
		require.Equal(t, exp, collect(ls))
	}

	// as a queue:
	sl = &List[int]{}
	var q Queue[int] = sl
	for _, e := range expected {
		q.Enqueue(e)
	}
	require.Equal(t, expected, collect(sl))
	for i, e := range expected {
		exp := expected[i+1:]
		got, ok := q.Dequeue()
		require.True(t, ok)
		require.Equal(t, e, got)
		require.Equal(t, exp, collect(sl))
	}

	// from the middle:
	sl = &List[int]{}
	for _, value := range expected {
		sl.Add(value)
	}
	require.Equal(t, expected, collect(sl))
	sl.Remove(4)
	require.Equal(t, []int{0, 1, 2, 3, 5}, collect(sl))
	sl.Remove(2)
	require.Equal(t, []int{0, 1, 3, 5}, collect(sl))
	sl.Remove(1)
	require.Equal(t, []int{0, 3, 5}, collect(sl))
	sl.Remove(5)
	require.Equal(t, []int{0, 3}, collect(sl))
	sl.Remove(0)
	require.Equal(t, []int{3}, collect(sl))
}

func collect(values Iterator[int]) (out []int) {
	out = []int{} // require.Equal() tests will fail when returning nil, since the expected slices are never nil
	values.Each(func(value int) {
		out = append(out, value)
	})
	return out
}
