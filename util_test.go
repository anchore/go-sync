package sync

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ToSeqToSlice(t *testing.T) {
	expected := []int{10, 11, 12, 13, 14}

	seq := ToSeq(expected)

	got := ToSlice(seq)

	require.EqualValues(t, expected, got)
}

func Test_ToKeyValueSeq(t *testing.T) {
	expected := map[string]int{"zero": 0, "one": 1, "two": 2, "three": 3, "four": 4}

	seq := ToSeq2(expected)

	got := keyValueSeqToMap(toKeyValueIterator(seq))

	require.EqualValues(t, expected, got)
}

func Test_ToIndexSeq(t *testing.T) {
	slice := []int{10, 11, 12, 13, 14}

	expected := map[int]int{
		0: 10,
		1: 11,
		2: 12,
		3: 13,
		4: 14,
	}

	seq := ToIndexSeq(slice)

	got := keyValueSeqToMap(toKeyValueIterator(seq))

	require.EqualValues(t, expected, got)
}

// keyValueSeqToMap converts an iter.Seq[KeyValue[K,V]] to a map[K]V
func keyValueSeqToMap[K comparable, V any](values iter.Seq[keyValue[K, V]]) map[K]V {
	out := map[K]V{}
	for kv := range values {
		out[kv.Key] = kv.Value
	}
	return out
}
