package index

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_treeAddPath(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected []namedNode
	}{
		{
			name: "/some/path",
			paths: []string{
				"/some/path",
			},
			expected: []namedNode{
				{
					Name: "some",
					Children: []namedNode{
						{
							Name: "path",
						},
					},
				},
			},
		},
		{
			name: "/some/paths",
			paths: []string{
				"/some/path1",
				"/some/path2",
			},
			expected: []namedNode{
				{
					Name: "some",
					Children: []namedNode{
						{Name: "path1"},
						{Name: "path2"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := TreeNode[any]{}
			for i, p := range test.paths {
				root.AddPath(p, i)
			}
			got := nodes(&root)
			expected := namedNode{
				Children: test.expected,
			}
			if diff := cmp.Diff(expected, got); diff != "" {
				t.Fatalf("mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

type namedNode struct {
	Name     string
	Children []namedNode
}

func nodes(n *TreeNode[any]) namedNode {
	var children []namedNode
	for _, child := range n.Children {
		children = append(children, nodes(child))
	}
	slices.SortFunc(children, func(a, b namedNode) int {
		return strings.Compare(a.Name, b.Name)
	})
	return namedNode{
		Name:     n.Name,
		Children: children,
	}
}
