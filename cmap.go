package cmap

import (
	"encoding/json"
	"fmt"
	"sync"
)

var SHARD_COUNT = 32

type Stringer interface {
	fmt.Stringer
	comparable
}

type ShardingFunc[K comparable, V any] func(key K) uint32

// A "thread" safe map of type string:Anything.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type ConcurrentMap[K comparable, V any] struct {
	sharding ShardingFunc[K, V]
	shards   []SafeMap[K, V]
}

// fnv32 函数实现了 FNV-1a 哈希算法的 32 位版本
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	// prime32 作为乘法因子，与 FNV-1a 哈希算法中使用的固定质数相对应
	const prime32 = uint32(16777619)
	keyLength := len(key)
	// 遍历输入的字符串
	for i := 0; i < keyLength; i++ {
		// 乘以 prime32 作为乘法因子
		hash *= prime32
		// 通过异或操作计算新的哈希值
		hash ^= uint32(key[i])
	}
	// 返回最终的哈希值
	return hash
}

func strfnv32[K fmt.Stringer](key K) uint32 {
	return fnv32(key.String())
}

func create[K comparable, V any](sharding ShardingFunc[K, V]) ConcurrentMap[K, V] {
	m := ConcurrentMap[K, V]{
		sharding: sharding,
		shards:   make([]SafeMap[K, V], SHARD_COUNT),
	}
	for i := 0; i < SHARD_COUNT; i++ {
		m.shards[i] = NewSafeMap[K, V]()
	}
	return m
}

// Creates a new concurrent map.
func New[V any]() ConcurrentMap[string, V] {
	return create[string, V](fnv32)
}

// Creates a new concurrent map.
func NewStringer[K Stringer, V any]() ConcurrentMap[K, V] {
	return create[K, V](strfnv32[K])
}

// Creates a new concurrent map.
func NewWithCustomShardingFunction[K comparable, V any](sharding ShardingFunc[K, V]) ConcurrentMap[K, V] {
	return create[K, V](sharding)
}

// GetShard returns shard under given key
func (m ConcurrentMap[K, V]) GetShard(key K) SafeMap[K, V] {
	return m.shards[uint(m.sharding(key))%uint(SHARD_COUNT)]
}

func (m ConcurrentMap[K, V]) MSet(data map[K]V) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Set(key, value)
	}
}

// Sets the given value under the specified key.
func (m ConcurrentMap[K, V]) Set(key K, value V) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Set(key, value)
}

type UpsertCb[V any] func(oldValue V, exist bool) V

// Insert or Update - updates existing element or inserts a new one using UpsertCb
func (m ConcurrentMap[K, V]) Upsert(key K, cb UpsertCb[V]) (result V) {
	shard := m.GetShard(key)
	shard.Update(func(m map[K]V) {
		v, exist := m[key]
		result = cb(v, exist)
		m[key] = result
	})
	return
}

// Sets the given value under the specified key if no value was associated with it.
func (m ConcurrentMap[K, V]) SetIfAbsent(key K, value V) (ok bool) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Update(func(m map[K]V) {
		_, ok = m[key]
		if !ok {
			m[key] = value
		}
	})
	return !ok
}

// Get retrieves an element from map under given key.
func (m ConcurrentMap[K, V]) Get(key K) (V, bool) {
	// Get shard
	shard := m.GetShard(key)
	return shard.Get(key)
}

type InsertCb[V any] func() V

// GetOrInsert The method is used to retrieve the value corresponding to a key in ConcurrentMap.
// If the key does not exist, a new value will be inserted using the provided callback function.
func (m ConcurrentMap[K, V]) GetOrInsert(key K, cb InsertCb[V]) V {
	// Get shard
	shard := m.GetShard(key)
	v, exist := shard.Get(key)
	if exist {
		return v
	}
	// update
	shard.Update(func(m map[K]V) {
		v, exist = m[key]
		if exist {
			return
		}
		v = cb()
		m[key] = v
	})
	return v
}

// Count returns the number of elements within the map.
func (m ConcurrentMap[K, V]) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m.shards[i]
		count += shard.Count()
	}
	return count
}

// Looks up an item under specified key
func (m ConcurrentMap[K, V]) Has(key K) bool {
	// Get shard
	shard := m.GetShard(key)
	_, ok := shard.Get(key)
	return ok
}

// Remove removes an element from the map.
func (m ConcurrentMap[K, V]) Remove(key K) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Del(key)
}

// RemoveCb is a callback executed in a map.RemoveCb() call, while Lock is held
// If returns true, the element will be removed from the map
type RemoveCb[K any, V any] func(key K, value V, exists bool) bool

// RemoveCb locks the shard containing the key, retrieves its current value and calls the callback with those params
// If callback returns true and element exists, it will remove it from the map
// Returns the value returned by the callback (even if element was not present in the map)
func (m ConcurrentMap[K, V]) RemoveCb(key K, cb RemoveCb[K, V]) (ok bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Update(func(m map[K]V) {
		v, exist := m[key]
		result := cb(key, v, exist)
		ok = exist && result
		if ok {
			delete(m, key)
		}
	})
	return
}

// Pop removes an element from the map and returns it
func (m ConcurrentMap[K, V]) Pop(key K) (value V, exists bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Update(func(m map[K]V) {
		value, exists = m[key]
		delete(m, key)
	})
	return
}

// IsEmpty checks if map is empty.
func (m ConcurrentMap[K, V]) IsEmpty() bool {
	return m.Count() == 0
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple[K comparable, V any] struct {
	Key K
	Val V
}

// IterBuffered returns a Iter iterator which could be used in a for range loop.
func (m ConcurrentMap[K, V]) IterBuffered() <-chan Tuple[K, V] {
	ch := make(chan Tuple[K, V], 1e3)
	go fanIn(m.snapshot(), ch)
	return ch
}

// Clear removes all items from map.
func (m ConcurrentMap[K, V]) Clear() {
	for item := range m.IterBuffered() {
		m.Remove(item.Key)
	}
}

func (m ConcurrentMap[K, V]) snapshot() []map[K]V {
	list := make([]map[K]V, 0, SHARD_COUNT)
	for _, shard := range m.shards {
		list = append(list, shard.Clone())
	}
	return list
}

func fanIn[K comparable, V any](shards []map[K]V, ch chan Tuple[K, V]) {
	wg := sync.WaitGroup{}
	for _, shard := range shards {
		wg.Add(1)
		go func(m map[K]V) {
			for k, v := range m {
				ch <- Tuple[K, V]{k, v}
			}
			wg.Done()
		}(shard)
	}
	wg.Wait()
	close(ch)
}

// Items returns all items as map[K]V
func (m ConcurrentMap[K, V]) Items() map[K]V {
	tmp := make(map[K]V)

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

// Iterator callbacalled for every key,value found in
// maps. RLock is held for all calls for a given shard
// therefore callback sess consistent view of a shard,
// but not across the shards
type IterCb[K comparable, V any] func(key K, v V)

// Callback based iterator, cheapest way to read
// all elements in a map.
func (m ConcurrentMap[K, V]) IterCb(fn IterCb[K, V]) {
	for _, shard := range m.shards {
		shard.View(fn)
	}
}

// Keys returns all keys as []K
func (m ConcurrentMap[K, V]) Keys() []K {
	// Generate keys
	keys := make([]K, 0, m.Count())
	for item := range m.IterBuffered() {
		keys = append(keys, item.Key)
	}
	return keys
}

// Values returns all Values as []V
func (m ConcurrentMap[K, V]) Values() []V {
	// Generate values
	values := make([]V, 0, m.Count())
	for item := range m.IterBuffered() {
		values = append(values, item.Val)
	}
	return values
}

// Reviles ConcurrentMap "private" variables to json marshal.
func (m ConcurrentMap[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Items())
}

// Reverse process of Marshal.
func (m *ConcurrentMap[K, V]) UnmarshalJSON(b []byte) error {
	tmp := make(map[K]V)

	// Unmarshal into a single map.
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	// foreach key,value pair in temporary map insert into our concurrent map.
	for key, val := range tmp {
		m.Set(key, val)
	}
	return nil
}
