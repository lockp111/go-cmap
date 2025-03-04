# concurrent map [![Build Status](https://travis-ci.com/lockp111/go-cmap.svg?branch=master)](https://travis-ci.com/lockp111/go-cmap)

As explained [here](http://golang.org/doc/faq#atomic_maps) and [here](http://blog.golang.org/go-maps-in-action), the `map` type in Go doesn't support concurrent reads and writes. `go-cmap` provides a high-performance solution to this by sharding the map with minimal time spent waiting for locks.

Prior to Go 1.9, there was no concurrent map implementation in the stdlib. In Go 1.9, `sync.Map` was introduced. The new `sync.Map` has a few key differences from this map. The stdlib `sync.Map` is designed for append-only scenarios. So if you want to use the map for something more like in-memory db, you might benefit from using our version. You can read more about it in the golang repo, for example [here](https://github.com/golang/go/issues/21035) and [here](https://stackoverflow.com/questions/11063473/map-with-concurrent-access)

## Why fork
Because I felt that some of the methods written by the original author were not very user-friendly, I refactored a version myself and added a separate thread-safe map.
It is convenient to use simpler thread-safe map in scenarios that do not require high concurrency.

## Usage

Import the package:

```go
import (
	"github.com/lockp111/go-cmap"
)

```

```bash
go get "github.com/lockp111/go-cmap"
```

The package is now imported under the "cmap" namespace.

## Example

```go

	// Create a new map.
	m := cmap.New[string]()

	// Sets item within map, sets "bar" under key "foo"
	m.Set("foo", "bar")

	// Retrieve item from map.
	bar, ok := m.Get("foo")

	// Removes item under key "foo"
	m.Remove("foo")

```

For more examples have a look at cmap_test.go.

Running tests:

```bash
go test "github.com/lockp111/go-cmap"
```

## Guidelines for contributing

Contributions are highly welcome. In order for a contribution to be merged, please follow these guidelines:
- Open an issue and describe what you are after (fixing a bug, adding an enhancement, etc.).
- According to the core team's feedback on the above mentioned issue, submit a pull request, describing the changes and linking to the issue.
- New code must have test coverage.
- If the code is about performance issues, you must include benchmarks in the process (either in the issue or in the PR).
- In general, we would like to keep `concurrent-map` as simple as possible and as similar to the native `map`. Please keep this in mind when opening issues.

## Language
- [中文说明](./README-zh.md)

## License
MIT (see [LICENSE](https://github.com/lockp111/go-cmap/blob/master/LICENSE) file)

## Performance Comparison

We conducted performance tests on ConcurrentMap, sync.Map, and standard map+lock in various scenarios. Here's an analysis of the results:

### Read/Write Ratio Tests

| Scenario | ConcurrentMap | sync.Map | Standard map+lock |
|----------|---------------|----------|-------------------|
| Read Heavy (90% reads) | 59.28 ns/op | 22.89 ns/op | 195.2 ns/op |
| Balanced (50% reads) | 85.47 ns/op | 72.72 ns/op | 176.0 ns/op |
| Write Heavy (90% writes) | 99.25 ns/op | 123.9 ns/op | 242.6 ns/op |

### Scale and Concurrency Tests (Read Heavy Scenario)

| Size | Goroutines | ConcurrentMap | sync.Map | Standard map+lock |
|------|------------|---------------|----------|-------------------|
| 1000 | 10 | 49.80 ns/op | 12.52 ns/op | 70.11 ns/op |
| 1000 | 50 | 55.00 ns/op | 18.91 ns/op | 187.1 ns/op |
| 1000 | 100 | 58.60 ns/op | 23.26 ns/op | 193.8 ns/op |
| 10000 | 10 | 51.14 ns/op | 8.948 ns/op | 86.93 ns/op |
| 10000 | 50 | 55.77 ns/op | 9.228 ns/op | 218.2 ns/op |
| 10000 | 100 | 59.46 ns/op | 9.408 ns/op | 210.5 ns/op |

### Performance Characteristics Analysis

1. **Read-dominant scenarios**:
   - sync.Map performs best, especially with large data sets
   - ConcurrentMap shows moderate performance, taking about 2-6 times longer than sync.Map
   - Standard map+lock performs worst, with performance deteriorating significantly as concurrency increases

2. **Balanced read/write scenarios**:
   - sync.Map and ConcurrentMap perform similarly
   - Standard map+lock still lags behind

3. **Write-dominant scenarios**:
   - ConcurrentMap performs best in write-intensive scenarios
   - sync.Map has weaker write performance and higher memory allocation
   - Standard map+lock performs worst in high-concurrency write operations

4. **Scalability**:
   - Standard map+lock performance drops significantly with increased concurrency, making it unsuitable for high-concurrency scenarios
   - ConcurrentMap performance slightly decreases with increased concurrency but remains relatively stable
   - sync.Map excels in high-concurrency read scenarios but underperforms with heavy writes

### Usage Recommendations

- For read-dominant workloads (>90% reads), sync.Map is recommended
- For balanced or write-heavy workloads, ConcurrentMap is recommended
- Avoid using standard map with locks in any high-concurrency scenario
- For large-scale data with mostly read operations, sync.Map's performance advantage is most significant