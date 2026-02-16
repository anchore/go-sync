package index

import (
	"slices"
)

func remove[T comparable](slice []T, removal T) []T {
	return removeIndex(slice, slices.Index(slice, removal))
}

func removeIndex[T comparable](slice []T, idx int) []T {
	if idx >= 0 && idx < len(slice) {
		return append(slice[:idx], slice[idx+1:]...)
	}
	return slice
}

func removeZeros[T comparable](vals []T) (out []T) {
	var zero T
	for _, v := range vals {
		if v == zero {
			continue
		}
		out = append(out, v)
	}
	return
}

func Flatten[T any](slices ...[]T) (out []T) {
	for _, slice := range slices {
		out = append(out, slice...)
	}
	return
}
