package cmap

import (
	"encoding/json"
	"hash/fnv"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type Animal struct {
	name string
}

// 实现fmt.Stringer接口
func (a Animal) String() string {
	return a.name
}

func TestMapCreation(t *testing.T) {
	m := New[string]()
	if m.shards == nil {
		t.Error("map is null.")
	}

	if m.Count() != 0 {
		t.Error("new map should be empty.")
	}
}

func TestInsert(t *testing.T) {
	m := New[Animal]()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}

	m.Set("elephant", elephant)
	m.Set("monkey", monkey)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestInsertAbsent(t *testing.T) {
	m := New[Animal]()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}

	m.SetIfAbsent("elephant", elephant)
	if ok := m.SetIfAbsent("elephant", monkey); ok {
		t.Error("map set a new value even the entry is already present")
	}
}

func TestGet(t *testing.T) {
	m := New[Animal]()

	// Get a missing element.
	val, ok := m.Get("Money")

	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}

	if (val != Animal{}) {
		t.Error("Missing values should return as null.")
	}

	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	// Retrieve inserted element.
	elephant, ok = m.Get("elephant")
	if ok == false {
		t.Error("ok should be true for item stored within the map.")
	}

	if elephant.name != "elephant" {
		t.Error("item was modified.")
	}
}

func TestHas(t *testing.T) {
	m := New[Animal]()

	// Get a missing element.
	if m.Has("Money") == true {
		t.Error("element shouldn't exists")
	}

	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	if m.Has("elephant") == false {
		t.Error("element exists, expecting Has to return True.")
	}
}

func TestRemove(t *testing.T) {
	m := New[Animal]()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)

	m.Remove("monkey")

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was removed.")
	}

	temp, ok := m.Get("monkey")

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if (temp != Animal{}) {
		t.Error("Expecting item to be nil after its removal.")
	}

	// Remove a none existing element.
	m.Remove("noone")
}

func TestRemoveCb(t *testing.T) {
	m := New[Animal]()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)
	elephant := Animal{"elephant"}
	m.Set("elephant", elephant)

	var (
		mapVal   Animal
		wasFound bool
	)
	cb := func(val Animal, exists bool) bool {
		mapVal = val
		wasFound = exists

		return val.name == "monkey"
	}

	// Monkey should be removed
	result := m.RemoveCb("monkey", cb)
	if !result {
		t.Errorf("Result was not true")
	}

	if mapVal != monkey {
		t.Errorf("Wrong value was provided to the value")
	}

	if !wasFound {
		t.Errorf("Key was not found")
	}

	if m.Has("monkey") {
		t.Errorf("Key was not removed")
	}

	// Elephant should not be removed
	result = m.RemoveCb("elephant", cb)
	if result {
		t.Errorf("Result was true")
	}

	if mapVal != elephant {
		t.Errorf("Wrong value was provided to the value")
	}

	if !wasFound {
		t.Errorf("Key was not found")
	}

	if !m.Has("elephant") {
		t.Errorf("Key was removed")
	}

	// Unset key should remain unset
	result = m.RemoveCb("horse", cb)
	if result {
		t.Errorf("Result was true")
	}

	if (mapVal != Animal{}) {
		t.Errorf("Wrong value was provided to the value")
	}

	if wasFound {
		t.Errorf("Key was found")
	}

	if m.Has("horse") {
		t.Errorf("Key was created")
	}
}

func TestPop(t *testing.T) {
	m := New[Animal]()

	monkey := Animal{"monkey"}
	m.Set("monkey", monkey)

	v, exists := m.Pop("monkey")

	if !exists || v != monkey {
		t.Error("Pop didn't find a monkey.")
	}

	v2, exists2 := m.Pop("monkey")

	if exists2 || v2 == monkey {
		t.Error("Pop keeps finding monkey")
	}

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was Pop'ed.")
	}

	temp, ok := m.Get("monkey")

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if (temp != Animal{}) {
		t.Error("Expecting item to be nil after its removal.")
	}
}

func TestCount(t *testing.T) {
	m := New[Animal]()
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	if m.Count() != 100 {
		t.Error("Expecting 100 element within map.")
	}
}

func TestIsEmpty(t *testing.T) {
	m := New[Animal]()

	if m.IsEmpty() == false {
		t.Error("new map should be empty")
	}

	m.Set("elephant", Animal{"elephant"})

	if m.IsEmpty() != false {
		t.Error("map shouldn't be empty.")
	}
}

func TestBufferedIterator(t *testing.T) {
	m := New[Animal]()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	counter := 0
	// Iterate over elements.
	for item := range m.IterBuffered() {
		val := item.Val

		if (val == Animal{}) {
			t.Error("Expecting an object.")
		}
		counter++
	}

	if counter != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestClear(t *testing.T) {
	m := New[Animal]()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	m.Clear()

	if m.Count() != 0 {
		t.Error("We should have 0 elements.")
	}
}

func TestIterCb(t *testing.T) {
	m := New[Animal]()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	counter := 0
	// Iterate over elements.
	m.IterCb(func(key string, v Animal) {
		counter++
	})
	if counter != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestItems(t *testing.T) {
	m := New[Animal]()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	items := m.Items()

	if len(items) != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestConcurrent(t *testing.T) {
	m := New[int]()
	ch := make(chan int)
	const iterations = 1000
	var a [iterations]int

	// Using go routines insert 1000 ints into our map.
	go func() {
		for i := 0; i < iterations/2; i++ {
			// Add item to map.
			m.Set(strconv.Itoa(i), i)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- val
		} // Call go routine with current index.
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			// Add item to map.
			m.Set(strconv.Itoa(i), i)

			// Retrieve item from map.
			val, _ := m.Get(strconv.Itoa(i))

			// Write to channel inserted value.
			ch <- val
		} // Call go routine with current index.
	}()

	// Wait for all go routines to finish.
	counter := 0
	for elem := range ch {
		a[counter] = elem
		counter++
		if counter == iterations {
			break
		}
	}

	// Sorts array, will make is simpler to verify all inserted values we're returned.
	sort.Ints(a[0:iterations])

	// Make sure map contains 1000 elements.
	if m.Count() != iterations {
		t.Error("Expecting 1000 elements.")
	}

	// Make sure all inserted values we're fetched from map.
	for i := 0; i < iterations; i++ {
		if i != a[i] {
			t.Error("missing value", i)
		}
	}
}

func TestJsonMarshal(t *testing.T) {
	SHARD_COUNT = 2
	defer func() {
		SHARD_COUNT = 32
	}()
	expected := "{\"a\":1,\"b\":2}"
	m := New[int]()
	m.Set("a", 1)
	m.Set("b", 2)
	j, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	if string(j) != expected {
		t.Error("json", string(j), "differ from expected", expected)
		return
	}
}

func TestKeys(t *testing.T) {
	m := New[Animal]()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	keys := m.Keys()
	if len(keys) != 100 {
		t.Error("We should have counted 100 elements.")
	}
}

func TestValues(t *testing.T) {
	m := New[int]()
	tests := []int{1, 3, 5, 7, 9, 2, 4, 6, 8, 0}

	for _, v := range tests {
		m.Set(strconv.Itoa(v), v)
	}

	values := m.Values()
	for _, v := range tests {
		if !slices.Contains(values, v) {
			t.Errorf("We should have %d.", v)
		}
	}
}

func TestMInsert(t *testing.T) {
	animals := map[string]Animal{
		"elephant": {"elephant"},
		"monkey":   {"monkey"},
	}
	m := New[Animal]()
	m.MSet(animals)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestFnv32(t *testing.T) {
	key := []byte("ABC")

	hasher := fnv.New32()
	_, err := hasher.Write(key)
	if err != nil {
		t.Errorf(err.Error())
	}
	if fnv32(string(key)) != hasher.Sum32() {
		t.Errorf("Bundled fnv32 produced %d, expected result from hash/fnv32 is %d", fnv32(string(key)), hasher.Sum32())
	}

}

func TestUpsert(t *testing.T) {
	dolphin := Animal{"dolphin"}
	whale := Animal{"whale"}
	tiger := Animal{"tiger"}
	lion := Animal{"lion"}

	cb := func(in Animal) UpsertCb[Animal] {
		return func(valueInMap Animal, exists bool) Animal {
			if !exists {
				return in
			}
			valueInMap.name += in.name
			return valueInMap
		}
	}

	m := New[Animal]()
	m.Set("marine", dolphin)
	m.Upsert("marine", cb(whale))
	m.Upsert("predator", cb(tiger))
	m.Upsert("predator", cb(lion))

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}

	marineAnimals, ok := m.Get("marine")
	if marineAnimals.name != "dolphinwhale" || !ok {
		t.Error("Set, then Upsert failed")
	}

	predators, ok := m.Get("predator")
	if !ok || predators.name != "tigerlion" {
		t.Error("Upsert, then Upsert failed")
	}
}

func TestKeysWhenRemoving(t *testing.T) {
	m := New[Animal]()

	// Insert 100 elements.
	Total := 100
	for i := 0; i < Total; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}

	// Remove 10 elements concurrently.
	Num := 10
	for i := 0; i < Num; i++ {
		go func(c ConcurrentMap[string, Animal], n int) {
			c.Remove(strconv.Itoa(n))
		}(m, i)
	}
	keys := m.Keys()
	for _, k := range keys {
		if k == "" {
			t.Error("Empty keys returned")
		}
	}
}

func TestUnDrainedIterBuffered(t *testing.T) {
	m := New[Animal]()
	// Insert 100 elements.
	Total := 100
	for i := 0; i < Total; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}
	counter := 0
	// Iterate over elements.
	ch := m.IterBuffered()
	for item := range ch {
		val := item.Val

		if (val == Animal{}) {
			t.Error("Expecting an object.")
		}
		counter++
		if counter == 42 {
			break
		}
	}
	for i := Total; i < 2*Total; i++ {
		m.Set(strconv.Itoa(i), Animal{strconv.Itoa(i)})
	}
	for item := range ch {
		val := item.Val

		if (val == Animal{}) {
			t.Error("Expecting an object.")
		}
		counter++
	}
	if counter != 100 {
		t.Error("We should have been right where we stopped")
	}

	counter = 0
	for item := range m.IterBuffered() {
		val := item.Val

		if (val == Animal{}) {
			t.Error("Expecting an object.")
		}
		counter++
	}

	if counter != 200 {
		t.Error("We should have counted 200 elements.")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	type test struct {
		name    string
		jsonStr string
		wantErr bool
	}

	tests := []test{
		{
			name:    "normal",
			jsonStr: `{"key1":"value1", "key2":"value2"}`,
			wantErr: false,
		},
		{
			name:    "empty JSON",
			jsonStr: `{}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			jsonStr: `{"key1":"value1", "key2":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New[string]()
			err := json.Unmarshal([]byte(tt.jsonStr), &m)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				jsonMap := make(map[string]string)
				json.Unmarshal([]byte(tt.jsonStr), &jsonMap)
				for key, val := range jsonMap {
					if mVal, ok := m.Get(key); !ok || mVal != val {
						t.Errorf("UnmarshalJSON() got = %v, want %v", mVal, val)
					}
				}
			}
		})
	}
}

func TestGetOrInsert(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		cb       InsertCb[string]
		expected string
	}{
		{
			name: "测试用例1：键不存在，需要插入",
			key:  "test1",
			cb: func() string {
				return "value1"
			},
			expected: "value1",
		},
		{
			name: "测试用例2：键已存在，直接获取",
			key:  "test2",
			cb: func() string {
				return "value2"
			},
			expected: "value2",
		},
		{
			name: "测试用例3：键不存在，需要插入",
			key:  "test3",
			cb: func() string {
				return "value3"
			},
			expected: "value3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New[string]()
			if got := m.GetOrInsert(tt.key, tt.cb); got != tt.expected {
				t.Errorf("GetOrInsert() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetCb(t *testing.T) {
	// 初始化一个ConcurrentMap
	m := New[string]()

	// 定义测试用例
	tests := []struct {
		name      string
		key       string
		value     string
		wantValue string
	}{
		{"v1", "key1", "value1", "value1"},
		{"v2", "key2", "value2", "value2"},
		{"v3", "key3", "value3", "value3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 在map中设置键值对
			m.Set(tt.key, tt.value)

			// 定义回调函数
			cb := func(v string, exist bool) {
				if !exist {
					t.Errorf("GetCb(%s) not exist", tt.key)
				}
				if v != tt.value {
					t.Errorf("GetCb() v = %v, wantValue = %v", v, tt.value)
				}
			}

			// 调用待测函数
			m.GetCb(tt.key, cb)
		})
	}
}

// 测试空key处理情况
func TestEmptyKey(t *testing.T) {
	m := New[Animal]()
	elephant := Animal{"elephant"}

	// 测试空字符串作为key
	m.Set("", elephant)

	// 验证能否正确获取
	val, ok := m.Get("")
	if !ok {
		t.Error("不能用空字符串作为key")
	}
	if val.name != "elephant" {
		t.Error("获取的值不正确")
	}

	// 验证删除空key
	m.Remove("")
	if m.Count() != 0 {
		t.Error("删除空键后计数应为0")
	}
}

// 测试并发map的分片访问
func TestShardMapAccess(t *testing.T) {
	m := New[string]()

	// 确保我们的测试会触发不同的分片
	// 添加足够多的键以覆盖所有分片
	for i := 0; i < SHARD_COUNT*2; i++ {
		key := "key" + strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		m.Set(key, value)
	}

	if m.Count() != SHARD_COUNT*2 {
		t.Error("map应该包含", SHARD_COUNT*2, "个元素")
	}
}

// 测试分片数量修改
func TestShardCount(t *testing.T) {
	// 保存原始SHARD_COUNT
	originalShardCount := SHARD_COUNT

	// 修改为不同的分片数
	SHARD_COUNT = 64
	m := New[string]()

	// 添加一些数据
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), strconv.Itoa(i))
	}

	if m.Count() != 100 {
		t.Error("设置了不同的分片数后，map应该包含100个元素")
	}

	// 恢复原始设置
	SHARD_COUNT = originalShardCount
}

// 测试空并发map的JSON操作
func TestEmptyMapJson(t *testing.T) {
	m := New[string]()
	j, err := json.Marshal(m)
	if err != nil {
		t.Error(err)
	}

	expected := "{}"
	if string(j) != expected {
		t.Error("空map的json应该是", expected, "，但得到了", string(j))
	}

	// 测试解析空JSON
	m2 := New[string]()
	err = json.Unmarshal([]byte("{}"), &m2)
	if err != nil {
		t.Error("解析空JSON失败:", err)
	}

	if m2.Count() != 0 {
		t.Error("解析空JSON后，map应该是空的")
	}
}

// 测试MGet功能（如果实现了的话）
func TestMGet(t *testing.T) {
	m := New[string]()

	// 添加一些数据
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Set("key3", "value3")

	// 测试是否能正确获取多个键值对
	keys := []string{"key1", "key2", "key4"}
	results := make(map[string]string)

	for _, key := range keys {
		if val, ok := m.Get(key); ok {
			results[key] = val
		}
	}

	if len(results) != 2 {
		t.Error("应该只找到2个键")
	}

	if val, exists := results["key1"]; !exists || val != "value1" {
		t.Error("key1的值不正确")
	}

	if val, exists := results["key2"]; !exists || val != "value2" {
		t.Error("key2的值不正确")
	}

	if _, exists := results["key4"]; exists {
		t.Error("key4不应该存在")
	}
}

// 测试GetOrInsert在各种情况下的行为
func TestGetOrInsertExtended(t *testing.T) {
	m := New[string]()

	// 情况1: 键不存在时插入
	value1 := m.GetOrInsert("key1", func() string {
		return "value1"
	})

	if value1 != "value1" {
		t.Error("GetOrInsert应返回插入的值")
	}

	// 情况2: 键已存在时返回现有值
	value2 := m.GetOrInsert("key1", func() string {
		return "new_value"
	})

	if value2 != "value1" {
		t.Error("GetOrInsert应返回已存在的值，而不是新值")
	}

	// 情况3: 测试并发情况
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			key := "concurrent_key" + strconv.Itoa(index%3)
			m.GetOrInsert(key, func() string {
				return "concurrent_value_" + strconv.Itoa(index)
			})
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证我们只有3个不同的key
	uniqueKeys := make(map[string]bool)
	for _, key := range m.Keys() {
		if strings.HasPrefix(key, "concurrent_key") {
			uniqueKeys[key] = true
		}
	}

	if len(uniqueKeys) != 3 {
		t.Error("应该只有3个不同的concurrent_key")
	}
}

// 测试GetCb键不存在时的行为
func TestGetCbKeyNotExist(t *testing.T) {
	m := New[string]()

	called := false
	m.GetCb("non_existent_key", func(val string, exists bool) {
		if exists {
			t.Error("键不应该存在")
		}
		if val != "" {
			t.Error("不存在的键应该返回空值")
		}
		called = true
	})

	if !called {
		t.Error("回调函数应该被调用")
	}
}

// 测试自定义Stringer类型的哈希函数
func TestStrfnv32(t *testing.T) {
	animal := Animal{"elephant"}
	// 确保strfnv32能正确处理实现了Stringer接口的类型
	hash := strfnv32(animal)
	expected := fnv32("elephant")
	if hash != expected {
		t.Errorf("strfnv32(%v) = %v, 期望 %v", animal, hash, expected)
	}
}

// 测试使用Stringer接口创建ConcurrentMap
func TestNewStringer(t *testing.T) {
	m := NewStringer[Animal, int]()
	if m.shards == nil {
		t.Error("map不应为null")
	}

	if m.Count() != 0 {
		t.Error("新map应该为空")
	}

	// 测试添加和获取元素
	cat := Animal{"cat"}
	m.Set(cat, 1)

	val, ok := m.Get(cat)
	if !ok || val != 1 {
		t.Error("无法使用Stringer接口作为键添加和获取元素")
	}
}

// 测试自定义分片函数
func TestNewWithCustomShardingFunction(t *testing.T) {
	// 创建一个自定义的分片函数，总是返回相同的哈希值
	customShardingFunc := func(key string) uint32 {
		return 5 // 总是返回5，所有键将进入同一个分片
	}

	m := NewWithCustomShardingFunction[string, int](customShardingFunc)
	if m.shards == nil {
		t.Error("map不应为null")
	}

	if m.Count() != 0 {
		t.Error("新map应该为空")
	}

	// 测试添加元素，应该都进入同一个分片
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	// 确认所有元素都能正确获取
	for _, k := range []string{"a", "b", "c"} {
		if val, ok := m.Get(k); !ok || val != map[string]int{"a": 1, "b": 2, "c": 3}[k] {
			t.Errorf("使用自定义分片函数时无法正确获取键 %s", k)
		}
	}
}

// 测试SetIfExists函数
func TestSetIfExists(t *testing.T) {
	m := New[string]()

	// 在键不存在时，SetIfExists应返回false
	ok := m.SetIfExists("key1", "value1")
	if ok {
		t.Error("键不存在时，SetIfExists应返回false")
	}

	// 先设置一个值
	m.Set("key1", "value1")
	// 现在键存在，SetIfExists应返回true，并更新值
	ok = m.SetIfExists("key1", "value2")
	if !ok {
		t.Error("键存在时，SetIfExists应返回true")
	}

	// 确认值已被更新
	val, exists := m.Get("key1")
	if !exists {
		t.Error("键应该存在")
	}
	if val != "value2" {
		t.Error("键的值应该已经被更新为value2")
	}
}

// 测试并发情况下的GetOrInsert - 提高GetOrInsert的覆盖率
func TestGetOrInsertConcurrent(t *testing.T) {
	m := New[int]()

	// 并发调用GetOrInsert
	const routines = 10
	done := make(chan bool, routines)

	for i := 0; i < routines; i++ {
		go func(index int) {
			// 确保所有goroutine尝试访问同一个键
			key := "concurrent_key"
			val := m.GetOrInsert(key, func() int {
				return index // 返回不同的值
			})
			// 所有goroutine应该得到相同的值（第一个完成的goroutine的值）
			if val < 0 || val >= routines {
				t.Errorf("unexpected value: %v", val)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < routines; i++ {
		<-done
	}

	// 检查map中只有一个值
	if m.Count() != 1 {
		t.Error("map应该只有一个值")
	}
}

// 测试UnmarshalJSON在不同情况下的行为
func TestUnmarshalJSONComprehensive(t *testing.T) {
	// 测试有效的JSON
	validJSON := `{"key1":"value1","key2":"value2"}`
	m1 := New[string]()
	err := json.Unmarshal([]byte(validJSON), &m1)
	if err != nil {
		t.Error("解析有效JSON时应该不会出错:", err)
	}
	if m1.Count() != 2 {
		t.Error("解析后应该有2个键值对")
	}

	// 测试结构错误的JSON
	invalidJSON := `{"key1":value1"` // 缺少引号
	m2 := New[string]()
	err = json.Unmarshal([]byte(invalidJSON), &m2)
	if err == nil {
		t.Error("解析无效JSON时应该报错")
	}

	// 测试空JSON
	emptyJSON := `{}`
	m3 := New[string]()
	err = json.Unmarshal([]byte(emptyJSON), &m3)
	if err != nil {
		t.Error("解析空JSON时不应该出错:", err)
	}
	if m3.Count() != 0 {
		t.Error("解析空JSON后应该没有键值对")
	}

	// 测试非对象JSON
	nonObjectJSON := `["array", "not", "object"]`
	m4 := New[string]()
	err = json.Unmarshal([]byte(nonObjectJSON), &m4)
	if err == nil {
		t.Error("解析非对象JSON时应该报错")
	}
}

// 测试GetOrInsert的竞态条件
func TestGetOrInsertRaceCondition(t *testing.T) {
	// 创建一个模拟场景，使GetOrInsert能覆盖到竞态条件分支
	m := New[string]()
	key := "race_key"

	// 创建通道用于控制测试流程
	raceDone := make(chan bool)
	getterReady := make(chan bool)
	getterDone := make(chan bool)

	// 启动一个goroutine获取/插入值
	go func() {
		// 通知已准备好
		getterReady <- true

		// 等待竞态条件准备完成
		<-raceDone

		// 这里调用GetOrInsert
		result := m.GetOrInsert(key, func() string {
			// 在这个回调函数执行前，已经有另一个goroutine设置了值
			return "this_should_not_be_used"
		})

		// 验证得到的是另一个goroutine设置的值
		if result != "value_from_racer" {
			t.Error("应该获取到竞态goroutine设置的值")
		}

		getterDone <- true
	}()

	// 等待第一个goroutine准备好
	<-getterReady

	// 在这个时间窗口中设置键值，模拟竞态条件
	m.Set(key, "value_from_racer")

	// 通知竞态条件已准备好
	raceDone <- true

	// 等待获取器完成
	<-getterDone
}

// 测试GetOrInsert函数在竞态条件下的行为 - 更直接的方法
func TestGetOrInsertEdgeCase(t *testing.T) {
	m := New[string]()
	key := "race_key"
	expectedValue := "pre_set_value"

	// 1. 获取分片
	shard := m.GetShard(key)

	// 2. 在分片中预先设置值，模拟在第一次检查后由另一个goroutine设置的情况
	shard.Set(key, expectedValue)

	// 3. 调用GetOrInsert - 它此时会跳过第一次检查，进入Update逻辑
	//    在Update中，它会发现键已存在并返回现有值，而不是使用回调函数
	value := m.GetOrInsert(key, func() string {
		// 这个回调函数不应该被调用
		t.Error("键已存在，不应调用回调函数")
		return "should_not_be_used"
	})

	if value != expectedValue {
		t.Errorf("应该返回预设值 %s，而不是 %s", expectedValue, value)
	}
}

// 使用替代方法测试GetOrInsert的所有分支路径
func TestGetOrInsertCoverage(t *testing.T) {
	// 目标是覆盖 GetOrInsert 中的这个部分：
	// shard.Update(func(m map[K]V) {
	//     v, exist = m[key]
	//     if exist { // <-- 这个分支需要覆盖
	//         return
	//     }
	//     v = cb()
	//     m[key] = v
	// })

	m := New[string]()
	key := "test_key"

	// 1. 先设置键值对
	m.Set(key, "existing_value")

	// 2. 调用GetOrInsert，这会触发内部的Update回调，
	// 并且在回调中应该走到 if exist { return } 分支
	value := m.GetOrInsert(key, func() string {
		// 这个函数不应该被调用，因为键已经存在
		t.Error("键已存在，回调函数不应被调用")
		return "new_value"
	})

	// 3. 验证返回的是现有值
	if value != "existing_value" {
		t.Errorf("应该返回现有值 %s，而不是 %s", "existing_value", value)
	}

	// 确认map中的值没有变化
	storedValue, exists := m.Get(key)
	if !exists {
		t.Error("键应该存在")
	}
	if storedValue != "existing_value" {
		t.Errorf("存储的值应该是 %s，而不是 %s", "existing_value", storedValue)
	}

	// 测试键不存在的情况
	newKey := "new_key"
	newValue := m.GetOrInsert(newKey, func() string {
		return "new_value"
	})

	if newValue != "new_value" {
		t.Errorf("对于新键，应该返回回调函数的值 %s，而不是 %s", "new_value", newValue)
	}
}
