package cmap

import (
	"maps"
	"strconv"
	"testing"
)

func stdCloneMap(m map[string]int) map[string]int {
	return maps.Clone(m)
}

func selfCloneMap(m map[string]int) map[string]int {
	newMap := make(map[string]int, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func BenchmarkCloneMap(b *testing.B) {
	m := map[string]int{}
	for i := 0; i < 10000; i++ {
		m[strconv.Itoa(i)] = i
	}

	b.Run("std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stdCloneMap(m)
		}
	})

	b.Run("self", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			selfCloneMap(m)
		}
	})
}
