package index

import (
	"strings"

	"github.com/anchore/go-sync"
)

type Tree[T any] struct {
	root                   TreeNode[T]
	RecursiveChildIndex    PrefixSuffix[Set[*TreeNode[T]]]
	RecursiveChildDirIndex PrefixSuffix[Set[*TreeNode[T]]]
}

type TreeNode[T any] struct {
	sync.Locking
	Parent *TreeNode[T]
	Name   string
	Value  *T
	//Children               []*TreeNode[T]
	ChildIndex KeySplitIndex[*TreeNode[T]]
}

func (f *TreeNode[T]) IsDir() bool {
	return f.Value == nil
}

func (f *TreeNode[T]) FullPath() string {
	if f.Parent != nil {
		return f.Parent.FullPath() + Sep + f.Name
	}
	return f.Name
}

func (f *TreeNode[T]) AddChild(node *TreeNode[T]) {
	defer f.Lock()()
	f.addChild(node)
}

func (f *TreeNode[T]) addChild(node *TreeNode[T]) {
	// NOTE: do not *set* the Parent in this function, it's set elsewhere due to
	// the fact that the child may have multiple parents, e.g. files with symlinks
	f.ChildIndex.Set(node.Name, node)

	// update direct children
	//f.Children = append(f.Children, node)

	// update all parent indexes
	addToIndexes(f, node)
}

func (f *TreeNode[T]) AddPath(path string, value T) {
	f.addPath(strings.Split(path, Sep), value)
}

func (f *TreeNode[T]) addPath(path []string, value T) {
	unlock := f.RLock()
	// don't add blank nodes, so /some/path is equivalent to some/path and some//path
	if path[0] == "" {
		unlock()
		f.addPath(path[1:], value)
		return
	}

	c := f.ChildIndex.Get(path[0])
	if c == nil {
		unlock()
		unlock = f.Lock()
		c = f.ChildIndex.Get(path[0])
		if c == nil {
			var v *T
			if len(path) == 1 {
				v = &value
			}
			c = &TreeNode[T]{
				Parent: f,
				Name:   path[0],
				Value:  v,
			}
			f.addChild(c)
		} else if len(path) == 1 {
			panic("duplicate value for path: " + strings.Join(path, Sep))
		}
		unlock()
	} else {
		unlock()
	}

	if len(path) > 1 {
		c.addPath(path[1:], value)
	}
}

func addToIndexes[T any](f, node *TreeNode[T]) {
	if f.Parent == nil {
		if node.IsDir() {
			f.RecursiveChildDirIndex.Update(node.Name, func(current []*TreeNode[T]) []*TreeNode[T] {
				return append(current, node)
			})
		} else {
			f.RecursiveChildIndex.Update(node.Name, func(current []*TreeNode[T]) []*TreeNode[T] {
				return append(current, node)
			})
		}
	}
	if f.Parent != nil {
		addToIndexes(f.Parent, node)
	}
}

//func addToIndexesLocking[T any](f, node *TreeNode[T]) {
//	defer f.Lock()()
//	addToIndexes(f, node)
//}
