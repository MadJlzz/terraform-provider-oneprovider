package common

import "slices"

// FindElement Find and return element.
func FindElement[T any](slice []T, predicate func(T) bool) (T, bool) {
	idx := slices.IndexFunc(slice, predicate)
	if idx == -1 {
		var zero T
		return zero, false
	}
	return slice[idx], true
}
