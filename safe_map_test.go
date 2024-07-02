package cmap

import (
	"testing"
)

func TestSafeMap_View(t *testing.T) {
	// 创建一个 SafeMap 实例
	safeMap := NewSafeMap[string, int]()

	// 向 SafeMap 中添加一些数据
	safeMap.m["key1"] = 10
	safeMap.m["key2"] = 20

	// 定义一个用于验证的函数
	var result []string
	verifyFn := func(k string, v int) {
		result = append(result, k)
	}

	// 调用 View 方法
	safeMap.View(verifyFn)

	// 验证结果
	expected := []string{"key1", "key2"}
	if !sliceEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// 比较两个字符串切片是否相等
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// 测试用例：正常情况
func TestSafeMapFind(t *testing.T) {
	// 创建 SafeMap 实例，并设置初始键值对
	safeMap := NewSafeMap[string, int]()
	safeMap.Set("key1", 10)
	safeMap.Set("key2", 20)

	// 定义回调函数，用于验证查找结果
	expectedMap := make(map[string]int)
	expectedMap["key1"] = 10
	expectedMap["key2"] = 20

	// 调用 SafeMap.Find 函数
	safeMap.Find(func(key string, value int, exist bool) {
		// 验证查找到的键值对是否与预期的一致
		if value != expectedMap[key] {
			t.Errorf("Expected value: %d, got value: %d", expectedMap[key], value)
		}
	})
}

// 测试用例：找不到的键
func TestSafeMapFindNotFound(t *testing.T) {
	// 创建 SafeMap 实例
	safeMap := NewSafeMap[string, int]()

	// 调用 SafeMap.Find 函数，查找不存在的键
	safeMap.Find(func(key string, value int, exist bool) {
		// 验证不存在的键是否正确处理
		if exist {
			t.Errorf("Expected not found, but found")
		}
	})
}

// 测试用例：空的键切片
func TestSafeMapFindEmptyKeys(t *testing.T) {
	// 创建 SafeMap 实例
	safeMap := NewSafeMap[string, int]()

	// 调用 SafeMap.Find 函数，传入空的键切片
	safeMap.Find(func(key string, value int, exist bool) {
		// 验证空键切片的处理是否正确
		t.Errorf("Expected no calls, but got calls")
	})
}

// TestUpdate 测试 Update 函数
func TestUpdate(t *testing.T) {
	// 创建一个 SafeMap 实例
	safeMap := NewSafeMap[string, int]()

	// 定义一个更新函数
	updateFn := func(m map[string]int) bool {
		// 在这里进行一些更新操作
		m["key1"] = 10
		return true
	}

	// 调用 Update 函数
	result := safeMap.Update(updateFn)

	// 验证结果
	if !result {
		t.Errorf("Update 函数返回错误结果")
	}

	// 验证 SafeMap 中的数据是否被正确更新
	value, ok := safeMap.Get("key1")
	if !ok || value != 10 {
		t.Errorf("SafeMap 中的数据未被正确更新")
	}
}
