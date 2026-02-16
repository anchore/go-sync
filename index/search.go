package index

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/anchore/go-sync"
)

const (
	AnyDir          = "**"
	SepRune         = '/'
	Sep             = string(SepRune)
	ChoiceStartRune = '{'
	ChoiceEndRune   = '}'
	ChoiceSplitRune = '|'
	WildcardRune    = '*'
)

func Search[T any](root *Tree[T], glob string) *Set[*TreeNode[T]] {
	glob = strings.Trim(glob, Sep)
	segments := strings.Split(glob, Sep)

	ctx := &searchContext[T]{
		root: root,
		out:  &Set[*TreeNode[T]]{},
	}
	ctx._search(root, segments, nil, true)

	return ctx.out
}

type searchContext[T any] struct {
	out  *Set[*TreeNode[T]]
	root *Tree[T]
}

func (ctx *searchContext[T]) _search(parent *TreeNode[T], segments []string, visited []*TreeNode[T], checkVisited bool) {
	if len(segments) == 0 {
		panic("should not be here")
	}

	if checkVisited {
		for _, v := range visited {
			if v == parent {
				return
			}
		}
		visited = append(visited, parent)
	}

	segment := segments[0]

	choiceIdx := strings.IndexRune(segment, ChoiceStartRune)
	if choiceIdx > 0 {
		for _, choice := range explodeChoices(segment) {
			childSegments := append([]string{choice}, segments[1:]...)
			ctx._search(parent, childSegments, visited, false)
		}
		return
	}

	var childProvider = func(parent *TreeNode[T]) (files []*TreeNode[T]) {
		child := parent.ChildIndex.Get(segment)
		if child != nil {
			return []*TreeNode[T]{child}
		}
		return nil
	}

	if segment == AnyDir {
		childProvider = func(parent *TreeNode[T]) (files []*TreeNode[T]) {
			// ** matches both the current directory _and_ any child directory
			files = append(files, parent)
			collectAllDirs(parent, &files)
			return files
		}
		// special case the form of a single double star followed by a single file that we have an index for
		if len(visited) == 1 {
			// this is the first child (visited == 0) the form is **/<something>
			switch len(segments) {
			case 2:
				// there are exactly 2 segments, the form is **/<file>
				fileSegment := segments[1]
				wcIdx := strings.IndexRune(fileSegment, WildcardRune)
				wcLastIdx := strings.LastIndexByte(fileSegment, byte(WildcardRune))

				switch {
				case wcIdx == 0 && wcLastIdx == 0:
					// *something form
					ctx.out.AppendAll(sync.ToSeq(Flatten(ctx.root.RecursiveChildIndex.BySuffix(fileSegment[1:])...)))
					return
				case wcIdx == len(fileSegment)-1 && wcLastIdx == wcIdx:
					// something* form
					ctx.out.AppendAll(sync.ToSeq(Flatten(ctx.root.RecursiveChildIndex.ByPrefix(fileSegment[:len(fileSegment)-1])...)))
					return
				case wcIdx < 0 && wcLastIdx < 0:
					// equal match form
					ctx.out.AppendAll(sync.ToSeq(ctx.root.RecursiveChildIndex.Get(fileSegment)))
					return
				}
			default:
				// there are more than 2 segments, so the second is directory of the form **/<dir>/<more-stuff>,
				// which we can use indexes to find in many cases
				nextDirSegment := segments[1]
				wcIdx := strings.IndexRune(nextDirSegment, WildcardRune)
				wcLastIdx := strings.LastIndexByte(nextDirSegment, byte(WildcardRune))

				switch {
				case wcIdx == 0 && wcLastIdx == 0:
					// *something form
					childProvider = func(parent *TreeNode[T]) (files []*TreeNode[T]) {
						return Flatten(ctx.root.RecursiveChildDirIndex.BySuffix(nextDirSegment[1:])...)
					}
					segments = segments[1:]
				case wcIdx == len(nextDirSegment)-1 && wcLastIdx == wcIdx:
					// something* form
					childProvider = func(parent *TreeNode[T]) (files []*TreeNode[T]) {
						return Flatten(ctx.root.RecursiveChildDirIndex.ByPrefix(nextDirSegment[:len(nextDirSegment)-1])...)
					}
					segments = segments[1:]
				case wcIdx < 0 && wcLastIdx < 0:
					// equal match form
					childProvider = func(parent *TreeNode[T]) (files []*TreeNode[T]) {
						return ctx.root.RecursiveChildDirIndex.Get(nextDirSegment)
					}
					segments = segments[1:]
				// cases below are for multiple wildcards, e.g. *some*thing, *something*,
				case wcIdx != 0:
					// multiple wildcards, but the first is NOT a wildcard; form: some*th*ng or some*thing*
					childProvider = func(parent *TreeNode[T]) (out []*TreeNode[T]) {
						results := ctx.root.RecursiveChildDirIndex.ByPrefix(nextDirSegment[:wcIdx-1])
						matches := segmentMatcher(nextDirSegment)
						for i := range results {
							byPrefix := results[i]
							for i := 0; i < len(byPrefix); i++ {
								f := byPrefix[i]
								if matches(f.Name) {
									out = append(out, f)
								}
							}
							results[i] = byPrefix
						}
						return out
					}
					segments = segments[1:]
				case wcLastIdx != len(nextDirSegment)-1:
					// multiple wildcards, but the last is NOT a wildcard; form: some*th*ng or *some*thing
					childProvider = func(parent *TreeNode[T]) (out []*TreeNode[T]) {
						results := ctx.root.RecursiveChildDirIndex.ByPrefix(nextDirSegment[wcLastIdx+1:])
						matches := segmentMatcher(nextDirSegment)
						for i := range results {
							byPrefix := results[i]
							for i := 0; i < len(byPrefix); i++ {
								f := byPrefix[i]
								if matches(f.Name) {
									out = append(out, f)
								}
							}
							results[i] = byPrefix
						}
						return out
					}
					segments = segments[1:]
				}
			}
		}

		// we have exact matches and prefix/suffix matches for subdirectories indexed

		// **/something will find in the root dir _and_ subdirs
		for _, c := range childProvider(parent) {
			ctx._search(c, segments[1:], visited, c != parent)
		}
		return
	}

	if strings.Contains(segment, AnyDir) {
		// doublestar contained in a larger string
		panic(fmt.Sprintf("larger patterns containing doublestars are not supported. got: %s", segment))
	}

	wcIdx := strings.IndexRune(segment, WildcardRune)
	if wcIdx >= 0 {
		if wcIdx == len(segment)-1 {
			if len(segment) == 1 {
				// we only have a single wildcard at the end, so we are able to use the file node index
				childProvider = func(parent *TreeNode[T]) []*TreeNode[T] {
					return parent.ChildIndex.Collect()
				}
			} else {
				// we only have a single wildcard at the end, so we are able to use the file node index
				childProvider = func(parent *TreeNode[T]) []*TreeNode[T] {
					return parent.ChildIndex.ByPrefix(segment[:len(segment)-1])
				}
			}
		} else {
			nameMatcher := segmentMatcher(segment)
			childProvider = func(parent *TreeNode[T]) (files []*TreeNode[T]) {
				for _, c := range parent.ChildIndex.Collect() {
					if nameMatcher(c.Name) {
						files = append(files, c)
					}
				}
				return files
			}
		}
	}

	if len(segments) == 1 {
		// last segment, the file match
		for _, child := range childProvider(parent) {
			if child.IsDir() {
				continue
			}
			ctx.out.Append(child)
		}
		return
	}

	// directory match, continue on with the results
	for _, child := range childProvider(parent) {
		ctx._search(child, segments[1:], visited, true)
	}
}

//func filterDirs[T any](get []*TreeNode[T]) []*TreeNode[T] {
//	var out []*TreeNode[T]
//	for _, child := range get {
//		if child.IsDir() {
//			out = append(out, child)
//		}
//	}
//	return out
//}

func collectAllDirs[T any](f *TreeNode[T], out *[]*TreeNode[T]) {
	collectAllChildrenFilter(f, out, func(f *TreeNode[T]) bool {
		return f.IsDir()
	})
}

func collectAllChildrenFilter[T any](file *TreeNode[T], out *[]*TreeNode[T], keep func(*TreeNode[T]) bool) {
	for _, child := range file.ChildIndex.Collect() {
		if keep(child) {
			*out = append(*out, child)
		}
		collectAllChildrenFilter(child, out, keep)
	}
}

var choicePattern = regexp.MustCompile(`\[\{]([^}]+)\[}]`)

func segmentMatcher(pattern string) func(segment string) bool {
	replace := func(s, pat string, r func(s string) string) string {
		return regexp.MustCompile(pat).ReplaceAllStringFunc(s, r)
	}
	pattern = replace(pattern, ".", func(s string) string {
		return "[" + s + "]"
	})
	pattern = choicePattern.ReplaceAllString(pattern, "($1)")
	pattern = strings.ReplaceAll(pattern, "[|]", "|")
	pattern = strings.ReplaceAll(pattern, "[*]", ".*")

	matcher := regexp.MustCompile("^" + pattern + "$")

	return func(segment string) bool {
		return matcher.MatchString(segment)
	}
}

func explodeChoices(glob string) (exploded []string) {
	_explodeChoices(glob, &exploded, "")
	return
}

func _explodeChoices(glob string, out *[]string, prior string) {
	i := 0
	for ; i < len(glob); i++ {
		if glob[i] == ChoiceStartRune {
			prior := prior + glob[:i]
			i++
			var choices []string
		findEnd:
			for end := i; end < len(glob); end++ {
				switch glob[end] {
				case ChoiceSplitRune:
					choices = append(choices, glob[i:end])
					i = end + 1
				case ChoiceEndRune:
					choices = append(choices, glob[i:end])
					i = end + 1
					break findEnd
				}
			}
			for _, choice := range choices {
				if i < len(glob) {
					_explodeChoices(glob[i:], out, prior+choice)
				} else {
					*out = append(*out, prior+choice)
				}
			}
			return
		}
	}
	*out = append(*out, prior+glob)
}

//func collectUnion[T comparable](exec executor.Executor, providers ...func() Set[T]) Set[T] {
//	collector := executor.NewCollector[Set[T]](exec)
//	for _, provider := range providers {
//		collector.Provide(provider)
//	}
//	results := collector.Collect()
//	if len(results) > 0 {
//		out := results[0]
//		for _, result := range results {
//			out.AddAll(result)
//		}
//		return out
//	}
//	return nil
//}

//func collectIntersection[T comparable](exec executor.Executor, providers ...func() Set[T]) Set[T] {
//	collector := executor.NewCollector[Set[T]](exec)
//	for _, provider := range providers {
//		collector.Provide(provider)
//	}
//	results := collector.Collect()
//	if len(results) > 0 {
//		out := results[0]
//		for _, result := range results {
//			out.KeepAll(result)
//		}
//		return out
//	}
//	return nil
//}
