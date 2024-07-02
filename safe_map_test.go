package cmap

import (
	"encoding/json"
	"slices"
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

	// 比较两个字符串切片是否相等 忽略顺序
	sliceEqual := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}

		for _, v := range a {
			if !slices.Contains(b, v) {
				return false
			}
		}

		return true
	}

	// 调用 View 方法
	safeMap.View(verifyFn)

	// 验证结果
	expected := []string{"key1", "key2"}
	if !sliceEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// 测试用例：正常情况
func TestSafeMap_Find(t *testing.T) {
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
func TestSafeMap_FindNotFound(t *testing.T) {
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
func TestSafeMap_FindEmptyKeys(t *testing.T) {
	// 创建 SafeMap 实例
	safeMap := NewSafeMap[string, int]()

	// 调用 SafeMap.Find 函数，传入空的键切片
	safeMap.Find(func(key string, value int, exist bool) {
		// 验证空键切片的处理是否正确
		t.Errorf("Expected no calls, but got calls")
	})
}

// TestUpdate 测试 Update 函数
func TestSafeMap_Update(t *testing.T) {
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

// TestMarshalJSON 测试 MarshalJSON 方法
func TestSafeMap_MarshalJSON(t *testing.T) {
	// 初始化测试数据
	safeMap := NewSafeMap[string, int]()
	safeMap.Set("key1", 1)
	safeMap.Set("key2", 2)

	// 调用 MarshalJSON 方法
	jsonData, err := safeMap.MarshalJSON()

	// 断言错误为 nil
	if err != nil {
		t.Errorf("MarshalJSON 方法错误: %v", err)
	}

	// 断言生成的 JSON 数据正确
	expectedJSON := `{"key1":1,"key2":2}`
	if string(jsonData) != expectedJSON {
		t.Errorf("MarshalJSON 方法生成的 JSON 数据不正确, 期望: %s, 实际: %s", expectedJSON, jsonData)
	}
}

// TestUnmarshalJSON 测试 UnmarshalJSON 函数
func TestSafeMap_UnmarshalJSON(t *testing.T) {
	// 创建一个 SafeMap 实例
	safeMap := NewSafeMap[string, int]()

	// 准备测试数据
	jsonData := []byte(`{"key1": 10, "key2": 20}`)

	// 调用 Unmarshal 函数
	err := json.Unmarshal(jsonData, &safeMap)

	// 检查错误是否为 nil
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	// 检查解包后的数据是否正确
	if safeMap.m["key1"] != 10 || safeMap.m["key2"] != 20 {
		t.Errorf("Expected map values: key1=10, key2=20, but got: %v", safeMap.m)
	}
}
