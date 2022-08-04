package helpers

import (
	"slices"
)

func SortedKeys[T any](m map[string]T) []string {
	keys := make([]string, len(m))
	idx := 0
	for key := range m {
		keys[idx] = key
		idx = idx + 1
	}
	slices.Sort(keys)
	return keys
}
