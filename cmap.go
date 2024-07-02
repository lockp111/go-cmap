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
	shards   []*SafeMap[K, V]
}

func create[K comparable, V any](sharding ShardingFunc[K, V]) *ConcurrentMap[K, V] {
	m := ConcurrentMap[K, V]{
		sharding: sharding,
		shards:   make([]*SafeMap[K, V], SHARD_COUNT),
	}
	for i := 0; i < SHARD_COUNT; i++ {
		m.shards[i] = NewSafeMap[K, V]()
	}
	return &m
}

// Creates a new concurrent map.
func New[V any]() *ConcurrentMap[string, V] {
	return create[string, V](fnv32)
}

// Creates a new concurrent map.
func NewStringer[K Stringer, V any]() *ConcurrentMap[K, V] {
	return create[K, V](strfnv32[K])
}

// Creates a new concurrent map.
func NewWithCustomShardingFunction[K comparable, V any](sharding ShardingFunc[K, V]) *ConcurrentMap[K, V] {
	return create[K, V](sharding)
}

// GetShard returns shard under given key
func (m *ConcurrentMap[K, V]) GetShard(key K) *SafeMap[K, V] {
	return m.shards[uint(m.sharding(key))%uint(SHARD_COUNT)]
}

func (m *ConcurrentMap[K, V]) MSet(data map[K]V) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Set(key, value)
	}
}

// Sets the given value under the specified key.
func (m *ConcurrentMap[K, V]) Set(key K, value V) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Set(key, value)
}

type UpsertCb[V any] func(oldValue V, exist bool) V

// Insert or Update - updates existing element or inserts a new one using UpsertCb
func (m *ConcurrentMap[K, V]) Upsert(key K, cb UpsertCb[V]) {
	shard := m.GetShard(key)
	shard.Update(func(m map[K]V) bool {
		v, exist := m[key]
		m[key] = cb(v, exist)
		return true
	})
}

// Sets the given value under the specified key if no value was associated with it.
func (m *ConcurrentMap[K, V]) SetIfAbsent(key K, value V) bool {
	// Get map shard.
	shard := m.GetShard(key)
	return shard.Update(func(m map[K]V) bool {
		_, ok := m[key]
		if !ok {
			m[key] = value
		}
		// Keep the existing value.
		return !ok
	})
}

// Get retrieves an element from map under given key.
func (m *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	// Get shard
	shard := m.GetShard(key)
	return shard.Get(key)
}

// Count returns the number of elements within the map.
func (m *ConcurrentMap[K, V]) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m.shards[i]
		count += shard.Count()
	}
	return count
}

// Looks up an item under specified key
func (m *ConcurrentMap[K, V]) Has(key K) bool {
	// Get shard
	shard := m.GetShard(key)
	_, ok := shard.Get(key)
	return ok
}

// Remove removes an element from the map.
func (m *ConcurrentMap[K, V]) Remove(key K) {
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
func (m *ConcurrentMap[K, V]) RemoveCb(key K, cb RemoveCb[K, V]) bool {
	// Try to get shard.
	shard := m.GetShard(key)
	return shard.Update(func(m map[K]V) bool {
		v, ok := m[key]
		result := cb(key, v, ok)
		if ok && result {
			delete(m, key)
		}
		return ok && result
	})
}

// Pop removes an element from the map and returns it
func (m *ConcurrentMap[K, V]) Pop(key K) (value V, exists bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Update(func(m map[K]V) bool {
		value, exists = m[key]
		delete(m, key)
		return true
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
func (m *ConcurrentMap[K, V]) IterBuffered() <-chan Tuple[K, V] {
	ch := make(chan Tuple[K, V], 1e3)
	go fanIn(m.snapshot(), ch)
	return ch
}

// Clear removes all items from map.
func (m *ConcurrentMap[K, V]) Clear() {
	for item := range m.IterBuffered() {
		m.Remove(item.Key)
	}
}

func (m *ConcurrentMap[K, V]) snapshot() []map[K]V {
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

// GetMap returns all items as map[string]V
func (m *ConcurrentMap[K, V]) GetMap() map[K]V {
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
func (m *ConcurrentMap[K, V]) IterCb(fn IterCb[K, V]) {
	for item := range m.IterBuffered() {
		fn(item.Key, item.Val)
	}
}

// Keys returns all keys as []K
func (m *ConcurrentMap[K, V]) Keys() []K {
	// Generate keys
	keys := make([]K, 0, m.Count())
	for item := range m.IterBuffered() {
		keys = append(keys, item.Key)
	}
	return keys
}

// Values returns all Values as []V
func (m *ConcurrentMap[K, V]) Values() []V {
	// Generate values
	values := make([]V, 0, m.Count())
	for item := range m.IterBuffered() {
		values = append(values, item.Val)
	}
	return values
}

// Reviles ConcurrentMap "private" variables to json marshal.
func (m *ConcurrentMap[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.GetMap())
}
func strfnv32[K fmt.Stringer](key K) uint32 {
	return fnv32(key.String())
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// Reverse process of Marshal.
func (m *ConcurrentMap[K, V]) UnmarshalJSON(b []byte) (err error) {
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
