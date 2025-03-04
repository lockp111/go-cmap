package cmap

import (
	"encoding/json"
	"maps"
	"sync"
)

type SafeMap[K comparable, V any] struct {
	m   map[K]V
	mux *sync.RWMutex
}

// NewSafeMap 创建一个新的键和值类型为 K 和 V 的 SafeMap 类型指针
func NewSafeMap[K comparable, V any]() SafeMap[K, V] {
	return SafeMap[K, V]{
		m:   make(map[K]V),
		mux: &sync.RWMutex{},
	}
}

// View 提供了对 SafeMap 中键值对的只读访问视图
func (s SafeMap[K, V]) View(fn func(K, V)) {
	// 读锁保护
	s.mux.RLock()
	// 最终解锁，确保即使发生错误，读锁也会被释放
	defer s.mux.RUnlock()

	// 遍历 SafeMap 中的键值对
	for k, v := range s.m {
		// 对每个键值对执行提供的函数
		fn(k, v)
	}
}

func (s SafeMap[K, V]) Clone() map[K]V {
	// 读锁保护
	s.mux.RLock()
	// 最终解锁，确保即使发生错误，读锁也会被释放
	defer s.mux.RUnlock()
	return maps.Clone(s.m)
}

// Find 允许通过特定的键值集合来查找 SafeMap 中的值，并通过提供的函数进行处理
func (s SafeMap[K, V]) Find(fn func(key K, value V, exist bool), keys ...K) {
	// 读锁保护
	s.mux.RLock()
	// 最终解锁，确保即使发生错误，读锁也会被释放
	defer s.mux.RUnlock()

	// 遍历要查找的键的切片
	for _, k := range keys {
		// 尝试从 SafeMap 中获取键对应的值
		v, ok := s.m[k]
		// 对每个键值对执行提供的函数
		fn(k, v, ok)
	}
}

func (s SafeMap[K, V]) Count() int {
	// 获取读锁
	s.mux.RLock()
	// 在函数退出时解锁
	defer s.mux.RUnlock()

	// 返回map中的元素个数
	return len(s.m)
}

// Get 方法用于获取键对应的值
func (s SafeMap[K, V]) Get(key K) (V, bool) {
	// 读锁保护，保证数据安全
	s.mux.RLock()
	// 使用 defer 确保即使发生错误，读锁也会被释放
	defer s.mux.RUnlock()

	// 尝试获取键对应的值
	v, ok := s.m[key]

	// 返回值和bool类型的存在标志
	return v, ok
}

// GetCb 方法用于获取键对应的值，并调用回调函数
func (s SafeMap[K, V]) GetCb(key K, cb func(value V, exists bool)) {
	// 读锁保护，保证数据安全
	s.mux.RLock()
	// 使用 defer 确保即使发生错误，读锁也会被释放
	defer s.mux.RUnlock()

	// 尝试获取键对应的值
	v, ok := s.m[key]

	// 调用回调函数
	cb(v, ok)
}

// Set 方法用于设置键值对
func (s SafeMap[K, V]) Set(key K, value V) {
	// 写锁保护，保证数据安全
	s.mux.Lock()
	// 使用 defer 确保即使发生错误，写锁也会被释放
	defer s.mux.Unlock()

	// 设置键值对
	s.m[key] = value
}

// Del 方法用于删除 SafeMap 中的指定键值对
func (s SafeMap[K, V]) Del(key K) {
	// 写锁保护
	s.mux.Lock()
	// 使用 defer 确保即使发生错误，写锁也会被释放
	defer s.mux.Unlock()

	// 删除键值对
	delete(s.m, key)
}

// Update 允许通过特定的更新逻辑更新 SafeMap 中的值
func (s SafeMap[K, V]) Update(fn func(map[K]V)) {
	// 写锁保护
	s.mux.Lock()
	// 最终解锁，确保即使发生错误，写锁也会被释放
	defer s.mux.Unlock()

	// 调用提供的更新函数，并获取返回值
	fn(s.m)
}

func (s SafeMap[K, V]) MarshalJSON() ([]byte, error) {
	// 写锁保护
	s.mux.Lock()
	// 最终解锁，确保即使发生错误，写锁也会被释放
	defer s.mux.Unlock()

	return json.Marshal(s.m)
}

// Reverse process of Marshal.
func (s *SafeMap[K, V]) UnmarshalJSON(b []byte) (err error) {
	// 写锁保护
	s.mux.Lock()
	// 最终解锁，确保即使发生错误，写锁也会被释放
	defer s.mux.Unlock()

	return json.Unmarshal(b, &s.m)
}
