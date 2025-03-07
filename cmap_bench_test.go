package cmap

import (
	"strconv"
	"sync"
	"testing"
)

type Integer int

func (i Integer) String() string {
	return strconv.Itoa(int(i))
}

func BenchmarkMarshalJson(b *testing.B) {
	m := New[Animal]()

	// Insert 100 elements.
	for i := 0; i < 10000; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}
	for i := 0; i < b.N; i++ {
		_, err := m.MarshalJSON()
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkStrconv(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strconv.Itoa(i)
	}
}

func BenchmarkSingleInsertAbsent(b *testing.B) {
	m := New[string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(strconv.Itoa(i), "value")
	}
}

func BenchmarkSingleInsertAbsentSyncMap(b *testing.B) {
	var m sync.Map
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(strconv.Itoa(i), "value")
	}
}

func BenchmarkSingleInsertPresent(b *testing.B) {
	m := New[string]()
	m.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set("key", "value")
	}
}

func BenchmarkSingleInsertPresentSyncMap(b *testing.B) {
	var m sync.Map
	m.Store("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store("key", "value")
	}
}

func benchmarkMultiInsertDifferent(b *testing.B) {
	m := New[string]()
	finished := make(chan struct{}, b.N)
	_, set := GetSet(m, finished)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i), "value")
	}
	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiInsertDifferentSyncMap(b *testing.B) {
	var m sync.Map
	finished := make(chan struct{}, b.N)
	_, set := GetSetSyncMap[string, string](&m, finished)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i), "value")
	}
	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiInsertDifferent_1_Shard(b *testing.B) {
	runWithShards(benchmarkMultiInsertDifferent, b, 1)
}
func BenchmarkMultiInsertDifferent_16_Shard(b *testing.B) {
	runWithShards(benchmarkMultiInsertDifferent, b, 16)
}
func BenchmarkMultiInsertDifferent_32_Shard(b *testing.B) {
	runWithShards(benchmarkMultiInsertDifferent, b, 32)
}
func BenchmarkMultiInsertDifferent_256_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetDifferent, b, 256)
}

func BenchmarkMultiInsertSame(b *testing.B) {
	m := New[string]()
	finished := make(chan struct{}, b.N)
	_, set := GetSet(m, finished)
	m.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set("key", "value")
	}
	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiInsertSameSyncMap(b *testing.B) {
	var m sync.Map
	finished := make(chan struct{}, b.N)
	_, set := GetSetSyncMap[string, string](&m, finished)
	m.Store("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set("key", "value")
	}
	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSame(b *testing.B) {
	m := New[string]()
	finished := make(chan struct{}, b.N)
	get, _ := GetSet(m, finished)
	m.Set("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go get("key", "value")
	}
	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSameSyncMap(b *testing.B) {
	var m sync.Map
	finished := make(chan struct{}, b.N)
	get, _ := GetSetSyncMap[string, string](&m, finished)
	m.Store("key", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go get("key", "value")
	}
	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func benchmarkMultiGetSetDifferent(b *testing.B) {
	m := New[string]()
	finished := make(chan struct{}, 2*b.N)
	get, set := GetSet(m, finished)
	m.Set("-1", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i-1), "value")
		go get(strconv.Itoa(i), "value")
	}
	for i := 0; i < 2*b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSetDifferentSyncMap(b *testing.B) {
	var m sync.Map
	finished := make(chan struct{}, 2*b.N)
	get, set := GetSetSyncMap[string, string](&m, finished)
	m.Store("-1", "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i-1), "value")
		go get(strconv.Itoa(i), "value")
	}
	for i := 0; i < 2*b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSetDifferent_1_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetDifferent, b, 1)
}
func BenchmarkMultiGetSetDifferent_16_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetDifferent, b, 16)
}
func BenchmarkMultiGetSetDifferent_32_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetDifferent, b, 32)
}
func BenchmarkMultiGetSetDifferent_256_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetDifferent, b, 256)
}

func benchmarkMultiGetSetBlock(b *testing.B) {
	m := New[string]()
	finished := make(chan struct{}, 2*b.N)
	get, set := GetSet(m, finished)
	for i := 0; i < b.N; i++ {
		m.Set(strconv.Itoa(i%100), "value")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i%100), "value")
		go get(strconv.Itoa(i%100), "value")
	}
	for i := 0; i < 2*b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSetBlockSyncMap(b *testing.B) {
	var m sync.Map
	finished := make(chan struct{}, 2*b.N)
	get, set := GetSetSyncMap[string, string](&m, finished)
	for i := 0; i < b.N; i++ {
		m.Store(strconv.Itoa(i%100), "value")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i%100), "value")
		go get(strconv.Itoa(i%100), "value")
	}
	for i := 0; i < 2*b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSetBlock_1_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetBlock, b, 1)
}
func BenchmarkMultiGetSetBlock_16_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetBlock, b, 16)
}
func BenchmarkMultiGetSetBlock_32_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetBlock, b, 32)
}
func BenchmarkMultiGetSetBlock_256_Shard(b *testing.B) {
	runWithShards(benchmarkMultiGetSetBlock, b, 256)
}

func GetSet[K comparable, V any](m ConcurrentMap[K, V], finished chan struct{}) (set func(key K, value V), get func(key K, value V)) {
	return func(key K, value V) {
			for i := 0; i < 10; i++ {
				m.Get(key)
			}
			finished <- struct{}{}
		}, func(key K, value V) {
			for i := 0; i < 10; i++ {
				m.Set(key, value)
			}
			finished <- struct{}{}
		}
}

func GetSetSyncMap[K comparable, V any](m *sync.Map, finished chan struct{}) (get func(key K, value V), set func(key K, value V)) {
	get = func(key K, value V) {
		for i := 0; i < 10; i++ {
			m.Load(key)
		}
		finished <- struct{}{}
	}
	set = func(key K, value V) {
		for i := 0; i < 10; i++ {
			m.Store(key, value)
		}
		finished <- struct{}{}
	}
	return
}

func runWithShards(bench func(b *testing.B), b *testing.B, shardsCount int) {
	oldShardsCount := SHARD_COUNT
	SHARD_COUNT = shardsCount
	bench(b)
	SHARD_COUNT = oldShardsCount
}

func BenchmarkKeys(b *testing.B) {
	m := New[Animal]()

	// Insert 100 elements.
	for i := 0; i < 10000; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}
	for i := 0; i < b.N; i++ {
		m.Keys()
	}
}

// BenchmarkSetIfExists 测试SetIfExists方法性能
func BenchmarkSetIfExists(b *testing.B) {
	// 键存在的情况
	b.Run("key_exists", func(b *testing.B) {
		m := New[string]()
		m.Set("key", "value")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.SetIfExists("key", "newvalue")
		}
	})

	// 键不存在的情况
	b.Run("key_not_exists", func(b *testing.B) {
		m := New[string]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.SetIfExists("key"+strconv.Itoa(i), "value")
		}
	})
}

// BenchmarkNewStringer 测试NewStringer方法性能
func BenchmarkNewStringer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewStringer[Integer, string]()
	}
}

// BenchmarkNewWithCustomShardingFunction 测试NewWithCustomShardingFunction方法性能
func BenchmarkNewWithCustomShardingFunction(b *testing.B) {
	customFunc := func(key string) uint32 {
		return uint32(len(key))
	}

	for i := 0; i < b.N; i++ {
		NewWithCustom[string, string](customFunc)
	}
}

// BenchmarkGetOrInsert 测试GetOrInsert方法性能
func BenchmarkGetOrInsert(b *testing.B) {
	// 键不存在的情况
	b.Run("key_not_exists", func(b *testing.B) {
		m := New[string]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i)
			m.GetOrInsert(key, func() string {
				return "value"
			})
		}
	})

	// 键存在的情况
	b.Run("key_exists", func(b *testing.B) {
		m := New[string]()
		for i := 0; i < 1000; i++ {
			key := "key" + strconv.Itoa(i)
			m.Set(key, "value")
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%1000)
			m.GetOrInsert(key, func() string {
				return "newvalue"
			})
		}
	})
}

// BenchmarkHas 测试Has方法性能
func BenchmarkHas(b *testing.B) {
	// 键存在的情况
	b.Run("key_exists", func(b *testing.B) {
		m := New[string]()
		for i := 0; i < 1000; i++ {
			key := "key" + strconv.Itoa(i)
			m.Set(key, "value")
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%1000)
			m.Has(key)
		}
	})

	// 键不存在的情况
	b.Run("key_not_exists", func(b *testing.B) {
		m := New[string]()
		for i := 0; i < 1000; i++ {
			key := "key" + strconv.Itoa(i)
			m.Set(key, "value")
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "nonexistent" + strconv.Itoa(i)
			m.Has(key)
		}
	})
}

// BenchmarkClear 测试Clear方法性能
func BenchmarkClear(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			m := New[string]()
			for j := 0; j < 100; j++ {
				m.Set("key"+strconv.Itoa(j), "value")
			}
			b.StartTimer()

			m.Clear()
		}
	})

	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			m := New[string]()
			for j := 0; j < 10000; j++ {
				m.Set("key"+strconv.Itoa(j), "value")
			}
			b.StartTimer()

			m.Clear()
		}
	})
}

// BenchmarkPop 测试Pop方法性能
func BenchmarkPop(b *testing.B) {
	b.Run("existing_key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := New[string]()
			m.Set("key", "value")
			_, _ = m.Pop("key")
		}
	})

	b.Run("non_existing_key", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := New[string]()
			_, _ = m.Pop("nonexistent")
		}
	})
}

// BenchmarkGetCb 测试GetCb方法性能
func BenchmarkGetCb(b *testing.B) {
	// 键存在的情况
	b.Run("key_exists", func(b *testing.B) {
		m := New[string]()
		for i := 0; i < 1000; i++ {
			key := "key" + strconv.Itoa(i)
			m.Set(key, "value")
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i%1000)
			m.GetCb(key, func(val string, exists bool) {
				// 回调函数
			})
		}
	})

	// 键不存在的情况
	b.Run("key_not_exists", func(b *testing.B) {
		m := New[string]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "key" + strconv.Itoa(i)
			m.GetCb(key, func(val string, exists bool) {
				// 回调函数
			})
		}
	})
}

// BenchmarkRemoveCb 测试RemoveCb方法性能
func BenchmarkRemoveCb(b *testing.B) {
	// 键存在且条件成立时删除
	b.Run("remove_when_true", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := New[string]()
			m.Set("key", "value")

			m.RemoveCb("key", func(val string, exists bool) bool {
				return true
			})
		}
	})

	// 键存在但条件不成立时不删除
	b.Run("not_remove_when_false", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := New[string]()
			m.Set("key", "value")

			m.RemoveCb("key", func(val string, exists bool) bool {
				return false
			})
		}
	})

	// 键不存在时的情况
	b.Run("key_not_exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := New[string]()

			m.RemoveCb("nonexistent", func(val string, exists bool) bool {
				return true
			})
		}
	})
}

// BenchmarkUnmarshalJSON 测试UnmarshalJSON方法性能
func BenchmarkUnmarshalJSON(b *testing.B) {
	// 准备一个小的JSON数据
	smallMap := New[string]()
	for i := 0; i < 100; i++ {
		smallMap.Set("key"+strconv.Itoa(i), "value")
	}
	smallJSON, _ := smallMap.MarshalJSON()

	// 准备一个大的JSON数据
	largeMap := New[string]()
	for i := 0; i < 10000; i++ {
		largeMap.Set("key"+strconv.Itoa(i), "value")
	}
	largeJSON, _ := largeMap.MarshalJSON()

	// 测试小数据量
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := New[string]()
			m.UnmarshalJSON(smallJSON)
		}
	})

	// 测试大数据量
	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := New[string]()
			m.UnmarshalJSON(largeJSON)
		}
	})
}
