package cmap

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

// BenchmarkConcurrentReadWrite 测试并发读写场景下的性能
func BenchmarkConcurrentReadWrite(b *testing.B) {
	// 测试并发度
	numGoroutines := 100

	// 预先准备数据
	m1 := New[string]()
	m2 := &sync.Map{}
	m3 := struct {
		m map[string]string
		sync.RWMutex
	}{m: make(map[string]string)}

	for i := 0; i < 1000; i++ {
		key := strconv.Itoa(i)
		value := "value" + key
		m1.Set(key, value)
		m2.Store(key, value)
		m3.Lock()
		m3.m[key] = value
		m3.Unlock()
	}

	// 测试读多写少场景 (90% 读, 10% 写)
	b.Run("ReadHeavy/ConcurrentMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%10 == 0 {
					// 写操作
					m1.Set(key, "new-value-"+key)
				} else {
					// 读操作
					m1.Get(key)
				}
				counter++
			}
		})
	})

	b.Run("ReadHeavy/SyncMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%10 == 0 {
					// 写操作
					m2.Store(key, "new-value-"+key)
				} else {
					// 读操作
					m2.Load(key)
				}
				counter++
			}
		})
	})

	b.Run("ReadHeavy/StdMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%10 == 0 {
					// 写操作
					m3.Lock()
					m3.m[key] = "new-value-" + key
					m3.Unlock()
				} else {
					// 读操作
					m3.RLock()
					_ = m3.m[key]
					m3.RUnlock()
				}
				counter++
			}
		})
	})

	// 测试读写均衡场景 (50% 读, 50% 写)
	b.Run("Balanced/ConcurrentMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%2 == 0 {
					// 写操作
					m1.Set(key, "new-value-"+key)
				} else {
					// 读操作
					m1.Get(key)
				}
				counter++
			}
		})
	})

	b.Run("Balanced/SyncMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%2 == 0 {
					// 写操作
					m2.Store(key, "new-value-"+key)
				} else {
					// 读操作
					m2.Load(key)
				}
				counter++
			}
		})
	})

	b.Run("Balanced/StdMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%2 == 0 {
					// 写操作
					m3.Lock()
					m3.m[key] = "new-value-" + key
					m3.Unlock()
				} else {
					// 读操作
					m3.RLock()
					_ = m3.m[key]
					m3.RUnlock()
				}
				counter++
			}
		})
	})

	// 测试写多读少场景 (10% 读, 90% 写)
	b.Run("WriteHeavy/ConcurrentMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%10 != 0 {
					// 写操作
					m1.Set(key, "new-value-"+key)
				} else {
					// 读操作
					m1.Get(key)
				}
				counter++
			}
		})
	})

	b.Run("WriteHeavy/SyncMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%10 != 0 {
					// 写操作
					m2.Store(key, "new-value-"+key)
				} else {
					// 读操作
					m2.Load(key)
				}
				counter++
			}
		})
	})

	b.Run("WriteHeavy/StdMap", func(b *testing.B) {
		b.ResetTimer()
		b.SetParallelism(numGoroutines)
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				key := strconv.Itoa(counter % 1000)
				if counter%10 != 0 {
					// 写操作
					m3.Lock()
					m3.m[key] = "new-value-" + key
					m3.Unlock()
				} else {
					// 读操作
					m3.RLock()
					_ = m3.m[key]
					m3.RUnlock()
				}
				counter++
			}
		})
	})
}

// BenchmarkConcurrentScale 测试不同规模和goroutine数量下的性能
func BenchmarkConcurrentScale(b *testing.B) {
	// 测试不同数据规模
	sizes := []int{100, 1000, 10000}
	// 测试不同并发度
	goroutineCounts := []int{10, 50, 100}

	for _, size := range sizes {
		// 预先准备数据
		m1 := New[string]()
		m2 := &sync.Map{}
		m3 := struct {
			m map[string]string
			sync.RWMutex
		}{m: make(map[string]string)}

		// 初始化测试数据
		for i := 0; i < size; i++ {
			key := strconv.Itoa(i)
			value := "value" + key
			m1.Set(key, value)
			m2.Store(key, value)
			m3.Lock()
			m3.m[key] = value
			m3.Unlock()
		}

		// 对每种并发度进行测试
		for _, goroutines := range goroutineCounts {
			testName := fmt.Sprintf("Size_%d_Goroutines_%d", size, goroutines)

			// 读为主的场景 (90% 读, 10% 写)
			b.Run(testName+"/ConcurrentMap", func(b *testing.B) {
				b.ResetTimer()
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						key := strconv.Itoa(counter % size)
						if counter%10 == 0 {
							// 写操作
							m1.Set(key, "new-value-"+key)
						} else {
							// 读操作
							m1.Get(key)
						}
						counter++
					}
				})
			})

			b.Run(testName+"/SyncMap", func(b *testing.B) {
				b.ResetTimer()
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						key := strconv.Itoa(counter % size)
						if counter%10 == 0 {
							// 写操作
							m2.Store(key, "new-value-"+key)
						} else {
							// 读操作
							m2.Load(key)
						}
						counter++
					}
				})
			})

			b.Run(testName+"/StdMap", func(b *testing.B) {
				b.ResetTimer()
				b.SetParallelism(goroutines)
				b.RunParallel(func(pb *testing.PB) {
					counter := 0
					for pb.Next() {
						key := strconv.Itoa(counter % size)
						if counter%10 == 0 {
							// 写操作
							m3.Lock()
							m3.m[key] = "new-value-" + key
							m3.Unlock()
						} else {
							// 读操作
							m3.RLock()
							_ = m3.m[key]
							m3.RUnlock()
						}
						counter++
					}
				})
			})
		}
	}
}
