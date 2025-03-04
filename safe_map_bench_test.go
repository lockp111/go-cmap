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

// 准备测试数据
func prepareSafeMap(size int) SafeMap[string, int] {
	safeMap := NewSafeMap[string, int]()
	for i := 0; i < size; i++ {
		key := "key" + strconv.Itoa(i)
		safeMap.Set(key, i)
	}
	return safeMap
}

// BenchmarkSafeMap_Set 测试Set方法性能
func BenchmarkSafeMap_Set(b *testing.B) {
	// 小数据量
	b.Run("small", func(b *testing.B) {
		safeMap := NewSafeMap[string, int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%100)
			safeMap.Set(key, i)
		}
	})

	// 大数据量
	b.Run("large", func(b *testing.B) {
		safeMap := NewSafeMap[string, int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%10000)
			safeMap.Set(key, i)
		}
	})
}

// BenchmarkSafeMap_Get 测试Get方法性能
func BenchmarkSafeMap_Get(b *testing.B) {
	// 小数据量
	b.Run("small", func(b *testing.B) {
		safeMap := prepareSafeMap(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%100)
			safeMap.Get(key)
		}
	})

	// 大数据量
	b.Run("large", func(b *testing.B) {
		safeMap := prepareSafeMap(10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%10000)
			safeMap.Get(key)
		}
	})
}

// BenchmarkSafeMap_GetCb 测试GetCb方法性能
func BenchmarkSafeMap_GetCb(b *testing.B) {
	safeMap := prepareSafeMap(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i%1000)
		safeMap.GetCb(key, func(value int, exists bool) {
			// 回调函数
		})
	}
}

// BenchmarkSafeMap_Del 测试Del方法性能
func BenchmarkSafeMap_Del(b *testing.B) {
	b.Run("existing_keys", func(b *testing.B) {
		safeMap := prepareSafeMap(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%1000)
			safeMap.Del(key)
		}
	})

	b.Run("non_existing_keys", func(b *testing.B) {
		safeMap := prepareSafeMap(1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "nonexistent" + strconv.Itoa(i%1000)
			safeMap.Del(key)
		}
	})
}

// BenchmarkSafeMap_Update 测试Update方法性能
func BenchmarkSafeMap_Update(b *testing.B) {
	safeMap := prepareSafeMap(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		safeMap.Update(func(m map[string]int) {
			key := "key" + strconv.Itoa(i%1000)
			m[key] = i
		})
	}
}

// BenchmarkSafeMap_Count 测试Count方法性能
func BenchmarkSafeMap_Count(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		safeMap := prepareSafeMap(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.Count()
		}
	})

	b.Run("large", func(b *testing.B) {
		safeMap := prepareSafeMap(10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.Count()
		}
	})
}

// BenchmarkSafeMap_View 测试View方法性能
func BenchmarkSafeMap_View(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		safeMap := prepareSafeMap(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.View(func(key string, value int) {
				// 仅遍历
			})
		}
	})

	b.Run("large", func(b *testing.B) {
		safeMap := prepareSafeMap(10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.View(func(key string, value int) {
				// 仅遍历
			})
		}
	})
}

// BenchmarkSafeMap_Find 测试Find方法性能
func BenchmarkSafeMap_Find(b *testing.B) {
	safeMap := prepareSafeMap(1000)

	// 测试查找单个键
	b.Run("single_key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%1000)
			safeMap.Find(func(key string, value int, exist bool) {
				// 回调函数
			}, key)
		}
	})

	// 测试查找多个键
	b.Run("multiple_keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key1 := "key" + strconv.Itoa(i%1000)
			key2 := "key" + strconv.Itoa((i+500)%1000)
			key3 := "nonexistent" + strconv.Itoa(i%100)
			safeMap.Find(func(key string, value int, exist bool) {
				// 回调函数
			}, key1, key2, key3)
		}
	})
}

// BenchmarkSafeMap_Clone 测试Clone方法性能
func BenchmarkSafeMap_Clone(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		safeMap := prepareSafeMap(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.Clone()
		}
	})

	b.Run("large", func(b *testing.B) {
		safeMap := prepareSafeMap(10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.Clone()
		}
	})
}

// BenchmarkSafeMap_MarshalJSON 测试MarshalJSON方法性能
func BenchmarkSafeMap_MarshalJSON(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		safeMap := prepareSafeMap(100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.MarshalJSON()
		}
	})

	b.Run("large", func(b *testing.B) {
		safeMap := prepareSafeMap(10000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap.MarshalJSON()
		}
	})
}

// BenchmarkSafeMap_UnmarshalJSON 测试UnmarshalJSON方法性能
func BenchmarkSafeMap_UnmarshalJSON(b *testing.B) {
	// 准备小型JSON数据
	smallMap := prepareSafeMap(100)
	smallJSON, _ := smallMap.MarshalJSON()

	// 准备大型JSON数据
	largeMap := prepareSafeMap(10000)
	largeJSON, _ := largeMap.MarshalJSON()

	b.Run("small", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap := NewSafeMap[string, int]()
			safeMap.UnmarshalJSON(smallJSON)
		}
	})

	b.Run("large", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			safeMap := NewSafeMap[string, int]()
			safeMap.UnmarshalJSON(largeJSON)
		}
	})
}
