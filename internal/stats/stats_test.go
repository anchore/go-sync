package stats

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Stats(t *testing.T) {
	const (
		stat1 Stat = iota
		stat2
	)

	s := NewStats(
		stat1,
		stat2,
	)

	s.Set(stat1, 1)

	require.Equal(t, 1.0, s.Get(stat1))

	s.Set(stat2, 2)

	require.Equal(t, 1.0, s.Get(stat1))
	require.Equal(t, 2.0, s.Get(stat2))

	s.Add(stat2, 6)

	require.Equal(t, 1.0, s.Get(stat1))
	require.Equal(t, 8.0, s.Get(stat2))

	s.Add(stat1, -2)

	require.Equal(t, -1.0, s.Get(stat1))
	require.Equal(t, 8.0, s.Get(stat2))
}
