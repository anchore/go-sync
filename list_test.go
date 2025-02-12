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
	require.Equal(t, expected, sl.Values())
	for i := range expected {
		ev := len(expected) - 1 - i
		exp := expected[:ev]
		v, ok := s.Pop()
		require.True(t, ok)
		require.Equal(t, expected[ev], v)
		require.Equal(t, exp, sl.Values())
	}

	// as a List:
	ls := &List[int]{}
	for _, v := range expected {
		ls.Append(v)
	}
	require.Equal(t, expected, ls.Values())
	for i, e := range expected {
		exp := expected[i+1:]
		ls.Remove(e)
		require.Equal(t, exp, ls.Values())
	}

	// as a queue:
	sl = &List[int]{}
	var q Queue[int] = sl
	for _, e := range expected {
		q.Enqueue(e)
	}
	require.Equal(t, expected, sl.Values())
	for i, e := range expected {
		exp := expected[i+1:]
		got, ok := q.Dequeue()
		require.True(t, ok)
		require.Equal(t, e, got)
		require.Equal(t, exp, sl.Values())
	}

	// from the middle:
	sl = &List[int]{}
	for _, value := range expected {
		sl.Append(value)
	}
	require.Equal(t, expected, sl.Values())
	sl.Remove(4)
	require.Equal(t, []int{0, 1, 2, 3, 5}, sl.Values())
	sl.Remove(2)
	require.Equal(t, []int{0, 1, 3, 5}, sl.Values())
	sl.Remove(1)
	require.Equal(t, []int{0, 3, 5}, sl.Values())
	sl.Remove(5)
	require.Equal(t, []int{0, 3}, sl.Values())
	sl.Remove(0)
	require.Equal(t, []int{3}, sl.Values())
}
