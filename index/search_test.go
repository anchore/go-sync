package index

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Search(t *testing.T) {
	root := setupTree(
		"direct",
		"sub/file",
		"sub/sub/file",
		"sub/sub/sub/file",
		"a/lot/of/nested/directories",
		"with/a/lot/more/nested/directories",
		"home/usr/package-lock.json",
		"usr/bin/node",
		"bin/python",
		"bin/python3",
		"file",
	)

	tests := []struct {
		glob     string
		expected []string
	}{
		{
			glob:     "not-at-root",
			expected: nil,
		},
		{
			glob:     "direct",
			expected: []string{"direct"},
		},
		{
			glob:     "sub/file",
			expected: []string{"file"},
		},
		{
			glob:     "a/lot/of/nested/directories",
			expected: []string{"directories"},
		},
		{
			glob:     "dire*",
			expected: []string{"direct"},
		},
		{
			glob:     "*ect",
			expected: []string{"direct"},
		},
		{
			glob:     "sub/fil*",
			expected: []string{"file"},
		},
		{
			glob:     "**/python",
			expected: []string{"python"},
		},
		{
			glob:     "**/python*",
			expected: []string{"python", "python3"},
		},
		{
			glob:     "**/*.json",
			expected: []string{"package-lock.json"},
		},
		{
			glob:     "**/lot/**/*ories",
			expected: []string{"directories", "directories"},
		},
		{
			glob:     "**/sub/file",
			expected: []string{"file", "file", "file"},
		},
		{
			glob:     "**/file",
			expected: []string{"file", "file", "file", "file"},
		},
		{
			glob:     "**/pack*lock*.json",
			expected: []string{"package-lock.json"},
		},
		{
			glob:     "**/more/**/directories",
			expected: []string{"directories"},
		},
	}

	for _, test := range tests {
		t.Run(test.glob, func(t *testing.T) {
			results := Search(root, test.glob)
			requireSameValues(t, test.expected, nodeNames(results))
		})
	}
}

func setupTree[T string](paths ...string) *TreeNode[T] {
	root := &TreeNode[T]{
		Parent:                 nil,
		Name:                   "",
		Value:                  nil,
		ChildIndex:             KeySplitIndex[*TreeNode[T]]{},
		RecursiveChildIndex:    PrefixSuffix[[]*TreeNode[T]]{},
		RecursiveChildDirIndex: PrefixSuffix[[]*TreeNode[T]]{},
	}

	for _, path := range paths {
		parent := root
		parts := strings.Split(path, "/")
		for i, name := range parts {
			var f *TreeNode[T]

			//for _, child := range parent.Children {
			for _, child := range parent.ChildIndex.Collect() {
				if child.Name == name {
					f = child
					break
				}
			}

			if f == nil {
				var v *T
				if i == len(parts)-1 {
					vt := T(name)
					v = &vt
				}
				f = &TreeNode[T]{
					Parent: parent,
					Name:   name,
					Value:  v,
				}
				parent.AddChild(f)
			}

			//// TODO ensure the proper indexes are being set up
			//if i == len(parts)-1 {
			//	c.fileNameIndex.Update(name, func(current *index.Node[[]*TreeNode[T]]) (newValue []*TreeNode[T]) {
			//		return append(current.Value(), f)
			//	})
			//
			//	//for j := 0; j < i; j++ {
			//	//	parentName := parts[j]
			//	//	c.parentNameIndex.Update(parentName, func(current *index.Node[[]*types.File]) (newValue []*types.File) {
			//	//		v := current.Value()
			//	//		return append(v, f)
			//	//	})
			//	//}
			//} else {
			//	f.IsDir = true
			//	c.dirNameIndex.Update(name, func(current *index.Node[[]*TreeNode[T]]) (newValue []*TreeNode[T]) {
			//		return append(current.Value(), f)
			//	})
			//}

			parent = f
		}
	}
	return root
}

func requireSameValues[T comparable](t *testing.T, v1 []T, v2 []T) {
	t.Helper()
	if len(v1) != len(v2) {
		t.Fatalf("not same length: %#v != %#v", v1, v2)
	}
next:
	for _, i := range v1 {
		for _, j := range v2 {
			if i == j {
				continue next
			}
		}
		t.Fatalf("element not found in array: %#v not in %#v", i, v2)
	}
}

func nodeNames[T any](results *Set[*TreeNode[T]]) (out []string) {
	for f := range results.Seq {
		out = append(out, f.Name)
	}
	return
}

func Test_explodeChoices(t *testing.T) {
	tests := []struct {
		glob     string
		expected []string
	}{
		{
			glob:     "a",
			expected: []string{"a"},
		},
		{
			glob:     "{a|b}",
			expected: []string{"a", "b"},
		},
		{
			glob:     "a{b|c|d}e",
			expected: []string{"abe", "ace", "ade"},
		},
		{
			glob:     "{b|c|d}ee",
			expected: []string{"bee", "cee", "dee"},
		},
		{
			glob:     "ae{b|c|d}",
			expected: []string{"aeb", "aec", "aed"},
		},
	}

	for _, test := range tests {
		t.Run(test.glob, func(t *testing.T) {
			got := explodeChoices(test.glob)
			require.Equal(t, test.expected, got)
		})
	}
}

func Test_segmentMatcher(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		fail    bool
	}{
		{
			pattern: "equals",
			input:   "equals",
		},
		{
			pattern: "different",
			input:   "equals",
			fail:    true,
		},
		{
			pattern: "substr*",
			input:   "substring",
		},
		{
			pattern: "*bstring",
			input:   "substring",
		},
		{
			pattern: "subs*ring",
			input:   "substring",
		},
		{
			pattern: "ch{oice|osen}",
			input:   "choice",
		},
		{
			pattern: "n{oise|olan}",
			input:   "nolan",
		},
		{
			pattern: "{n|so-}{ice|bright}",
			input:   "so-bright",
		},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			m := segmentMatcher(test.pattern)
			got := m(test.input)
			require.Equal(t, !test.fail, got)
		})
	}
}
