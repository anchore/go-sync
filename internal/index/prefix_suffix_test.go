package index

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_PrefixSuffix(t *testing.T) {
	i := PrefixSuffix[int]{}
	i.Set("one", 1)
	i.Set("once", 11)
	i.Set("onesie", 111)
	i.Set("two", 2)
	i.Set("done", 99)
	require.Equal(t, 1, i.Get("one"))
	require.Equal(t, 11, i.Get("once"))
	require.Equal(t, 111, i.Get("onesie"))
	require.Equal(t, 2, i.Get("two"))
	require.ElementsMatch(t, i.ByPrefix("on"), []int{1, 11, 111})
	require.ElementsMatch(t, i.BySuffix("e"), []int{1, 11, 111, 99})
	require.ElementsMatch(t, i.BySuffix(""), []int{1, 2, 11, 111, 99})
	require.ElementsMatch(t, i.BySuffix("one"), []int{1, 99})
	require.ElementsMatch(t, i.BySuffix("sie"), []int{111})
}

func Test_reverse(t *testing.T) {
	require.Equal(t, "case", reverse("esac"))
}
