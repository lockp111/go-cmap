# concurrent map [![Build Status](https://travis-ci.com/lockp111/go-cmap.svg?branch=master)](https://travis-ci.com/lockp111/go-cmap)

正如 [这里](http://golang.org/doc/faq#atomic_maps) 和 [这里](http://blog.golang.org/go-maps-in-action)所描述的, Go语言原生的`map`类型并不支持并发读写。`go-cmap`提供了一种高性能的解决方案:通过对内部`map`进行分片，降低锁粒度，从而达到最少的锁等待时间(锁冲突)

在Go 1.9之前，go语言标准库中并没有实现并发`map`。在Go 1.9中，引入了`sync.Map`。新的`sync.Map`与此`go-cmap`有几个关键区别。标准库中的`sync.Map`是专为`append-only`场景设计的。因此，如果您想将`Map`用于一个类似内存数据库，那么使用我们的版本可能会受益。你可以在golang repo上读到更多，[这里](https://github.com/golang/go/issues/21035) and [这里](https://stackoverflow.com/questions/11063473/map-with-concurrent-access)
***译注:`sync.Map`在读多写少性能比较好，否则并发性能很差***

## 为什么fork
因为我觉得原作者写的一些方法不太友好，所以我自己重构了一个版本，并添加了一个独立的线程安全map。
在不需要高并发的情况下，可以使用更简单的线程安全map。

## 用法

导入包:

```go
import (
	"github.com/lockp111/go-cmap"
)

```

```bash
go get github.com/lockp111/go-cmap
```

现在包被导入到了`cmap`命名空间下
***译注:通常包的限定前缀(命名空间)是和目录名一致的，但是这个包有点典型😂，不一致！！！所以用的时候注意***

## 示例

```go

	// 创建一个新的 map.
	m := cmap.New[string]()

	// 设置变量m一个键为"foo"值为"bar"键值对
	m.Set("foo", "bar")

	// 从m中获取指定键值.
	bar, ok := m.Get("foo")

	// 删除键为"foo"的项
	m.Remove("foo")

```

更多使用示例请查看`cmap_test.go`.

运行测试:

```bash
go test "github.com/lockp111/go-cmap"
```

## 贡献说明

我们非常欢迎大家的贡献。如欲合并贡献，请遵循以下指引:
- 新建一个issue,并且叙述为什么这么做(解决一个bug，增加一个功能，等等)
- 根据核心团队对上述问题的反馈，提交一个PR，描述变更并链接到该问题。
- 新代码必须具有测试覆盖率。
- 如果代码是关于性能问题的，则必须在流程中包括基准测试(无论是在问题中还是在PR中)。
- 一般来说，我们希望`go-cmap`尽可能简单，且与原生的`map`有相似的操作。当你新建issue时请注意这一点。

## 许可证
MIT (see [LICENSE](https://github.com/lockp111/go-cmap/blob/master/LICENSE) file)

## 性能对比

我们对ConcurrentMap、sync.Map和标准map+锁进行了多种场景的性能测试，以下是测试结果的分析：

### 读写比例场景测试

| 场景 | ConcurrentMap | sync.Map | 标准map+锁 |
|-----|--------------|----------|-----------|
| 读多写少(90%读) | 59.28 ns/op | 22.89 ns/op | 195.2 ns/op |
| 读写均衡(50%读) | 85.47 ns/op | 72.72 ns/op | 176.0 ns/op |
| 写多读少(90%写) | 99.25 ns/op | 123.9 ns/op | 242.6 ns/op |

### 规模与并发度测试 (读多写少场景)

| 规模 | 并发数 | ConcurrentMap | sync.Map | 标准map+锁 |
|-----|-------|--------------|----------|-----------|
| 1000 | 10 | 49.80 ns/op | 12.52 ns/op | 70.11 ns/op |
| 1000 | 50 | 55.00 ns/op | 18.91 ns/op | 187.1 ns/op |
| 1000 | 100 | 58.60 ns/op | 23.26 ns/op | 193.8 ns/op |
| 10000 | 10 | 51.14 ns/op | 8.948 ns/op | 86.93 ns/op |
| 10000 | 50 | 55.77 ns/op | 9.228 ns/op | 218.2 ns/op |
| 10000 | 100 | 59.46 ns/op | 9.408 ns/op | 210.5 ns/op |

### 性能特点分析

1. **读操作为主的场景**：
   - sync.Map 表现最佳，特别是在大规模数据时性能优势明显
   - ConcurrentMap 性能适中，约为sync.Map的2-6倍时间
   - 标准map+锁性能最差，并且随着并发度增加性能下降显著

2. **读写均衡的场景**：
   - sync.Map 和 ConcurrentMap 性能相近
   - 标准map+锁仍明显落后

3. **写操作为主的场景**：
   - ConcurrentMap 在写密集场景表现最好
   - sync.Map 写入性能较弱，内存分配也更多
   - 标准map+锁在高并发写入时性能最差

4. **扩展性**：
   - 标准map+锁随并发度增加性能下降明显，不适合高并发场景
   - ConcurrentMap 性能随并发度增加略有下降，但保持相对稳定
   - sync.Map 在高并发读取场景非常出色，但在大量写入时表现较差

### 使用建议

- 如果读操作占绝大多数（>90%），推荐使用 sync.Map
- 如果写操作较多或读写比例均衡，推荐使用 ConcurrentMap
- 在任何高并发场景下都应避免使用带锁的标准map
- 对于大规模数据且读多写少的场景，sync.Map的性能优势更为明显
