# go-ordmap

[![main](https://github.com/edofic/go-ordmap/actions/workflows/main.yml/badge.svg?branch=v2)](https://github.com/edofic/go-ordmap/actions/workflows/main.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/edofic/go-ordmap)](https://goreportcard.com/report/github.com/edofic/go-ordmap)
[![Go Reference](https://pkg.go.dev/badge/github.com/edofic/go-ordmap/v2.svg)](https://pkg.go.dev/github.com/edofic/go-ordmap/v2)

Persistent generic ordered maps for Go.

This is `v2` which utilizes Go 1.23+ iterators (`iter.Seq2`) for efficient and idiomatic collection traversal. If you can't (or don't want to) use Go 1.23, there is still `v1` which uses code generation. See the [v1 branch](https://github.com/edofic/go-ordmap/tree/v1).

## Features

- **Ordered**: Keys are always sorted, enabling O(1) access to Min/Max and O(n) in-order iteration.
- **Persistent (Immutable)**: Every modification returns a new map; the original remains untouched. This enables safe concurrent reads without locks (similar to MVCC).
- **Generic**: Fully type-safe via Go generics. No code generation or reflection required.
- **Idiomatic**: Leverages Go 1.23 range-over-func iterators for zero-allocation forward, backward, and range-scoped traversal.
- **Fast**: O(log n) lookups, insertions, and deletions via AVL self-balancing trees.
- **Extensible**: Includes sub-packages for high-churn workloads (`generational`) and multi-indexed repositories (`indexed`).

## Installation

```sh
go get github.com/edofic/go-ordmap/v2@latest
```

Requires **Go 1.23+**.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/edofic/go-ordmap/v2"
)

func main() {
	// Create a map for built-in comparable types (int, string, float64, etc.)
	m := ordmap.NewBuiltin[int, string]()

	// Insert returns a new map; always re-assign the result
	m = m.Insert(1, "foo")
	m = m.Insert(2, "bar")
	m = m.Insert(2, "baz") // overwrites existing key

	// Lookup
	v, ok := m.Get(2)
	fmt.Println(v, ok) // "baz" true

	// Ordered iteration (forward)
	for k, v := range m.All() {
		fmt.Println(k, v)
	}

	// Ordered iteration (backward)
	for k, v := range m.Backward() {
		fmt.Println(k, v)
	}

	// Range scan: iterate from key >= 2
	for k, v := range m.From(2) {
		fmt.Println("From 2:", k, v)
	}

	// Min / Max
	fmt.Println(m.Min(), m.Max())

	// Remove
	m = m.Remove(1)
	fmt.Println(m.Len()) // 1
}
```

> **Important**: Because the map is persistent, you must always assign the return value of `Insert` and `Remove`. The original value is never modified in-place.

## Rationale

The Go standard library does not provide a key-value data structure that also allows quick access to minimum/maximum and in-order iteration.

This package provides such a data structure (called `OrdMap` for "ordered map") implemented on top of [AVL trees](https://en.wikipedia.org/wiki/AVL_tree). This gives us:

| Operation | Complexity |
|-----------|-----------|
| Insert    | O(log n)  |
| Remove    | O(log n)  |
| Get       | O(log n)  |
| Min / Max | O(log n)  |
| All / From| O(n)      |

One detail that may not be idiomatic for Go is that these `OrdMap`s are [persistent data structures](https://en.wikipedia.org/wiki/Persistent_data_structure) — meaning every operation gives you back a new map while keeping the old intact. This is beneficial when you have many concurrent readers; writers can advance while readers safely traverse older versions (similar to [MVCC](https://en.wikipedia.org/wiki/Multiversion_concurrency_control)).

## Custom Key Types

For built-in types (`int`, `string`, `float64`, etc.), use `ordmap.NewBuiltin`. For custom types, implement the `Comparable` interface:

```go
type Comparable[A any] interface {
	Less(A) bool
}
```

Example with a composite key:

```go
type CompositeKey struct {
	User       int
	Preference int
}

func (c1 CompositeKey) Less(c2 CompositeKey) bool {
	if c1.User < c2.User {
		return true
	}
	if c1.User > c2.User {
		return false
	}
	return c1.Preference < c2.Preference
}

m := ordmap.New[CompositeKey, bool]()
m = m.Insert(CompositeKey{1, 1}, true)

// Range scan / prefix query: all preferences for user 2
targetUser := 2
for k, v := range m.From(CompositeKey{targetUser, 0}) {
	if k.User != targetUser {
		break
	}
	fmt.Println(k, v)
}
```

## API Overview

### Core Types

- `Node[K, V]` — The underlying AVL tree node. Works with any `Comparable[K]` key.
- `NodeBuiltin[K, V]` — Convenience wrapper for built-in types. Returned by `NewBuiltin`.

### Methods

All methods are available on both `Node` and `NodeBuiltin` (with appropriate key type unwrapping for `NodeBuiltin`):

| Method | Description |
|--------|-------------|
| `Get(key) (V, bool)` | Lookup a value by key. |
| `Insert(key, value) T` | Insert or update a key. Returns a new map. |
| `Remove(key) T` | Delete a key. Returns a new map. |
| `Len() int` | Number of elements. |
| `Entries() []Entry[K, V]` | Slice of all entries sorted by key. |
| `Min() *Entry[K, V]` | Entry with the smallest key. |
| `Max() *Entry[K, V]` | Entry with the largest key. |
| `All() iter.Seq2[K, V]` | Forward in-order iterator. |
| `Backward() iter.Seq2[K, V]` | Reverse in-order iterator. |
| `From(k) iter.Seq2[K, V]` | Forward iterator starting from the first key >= `k`. |
| `BackwardFrom(k) iter.Seq2[K, V]` | Reverse iterator starting from the first key <= `k`. |

## Sub-packages

### `generational`

For use cases with **high churn** (many short-lived items), the `generational` package provides an optimized wrapper. It uses a "Young" and "Old" generation approach inspired by generational garbage collection and LSM-trees.

- **Writes**: Extremely fast (only affect the small Young generation).
- **Reads**: Slightly slower (check both generations).
- **Iteration**: Performed via a live merge of both generations.

```go
import "github.com/edofic/go-ordmap/v2/generational"

// Create a map with a young generation limit of 1000
m := generational.New[MyKey, MyValue](1000)

m = m.Insert(k, v)
m = m.Remove(k)
val, ok := m.Get(k)

// All iterators work just like the standard OrdMap
for k, v := range m.All() {
    // ...
}
for k, v := range m.From(k) {
    // ...
}
```

See [`examples/generational`](examples/generational/main.go) for a complete demo.

### `indexed`

The `indexed` package provides a **multi-indexed, persistent, in-memory repository** built on top of `go-ordmap`. It allows you to define secondary indexes on entity fields and query by exact match or range scan.

```go
import (
	"github.com/edofic/go-ordmap/v2"
	"github.com/edofic/go-ordmap/v2/indexed"
)

type Order struct {
	ID       int
	Customer string
	Amount   int
}

// Define index handles (type-safe)
var ByCustomer = indexed.StringIndex(func(o Order) string { return o.Customer })
var ByAmount   = indexed.IntIndex(func(o Order) int { return o.Amount })

// Setup repository with a primary key extractor
repo := indexed.NewBuiltinRepo(func(o Order) int { return o.ID })
repo = indexed.AddIndex(repo, ByCustomer)
repo = indexed.AddIndex(repo, ByAmount)

// Insert / Delete (returns new repo state)
repo = repo.Insert(Order{ID: 1, Customer: "Alice", Amount: 100})
repo = repo.Delete(ordmap.Box(1))

// Query by primary key
order, found := repo.Get(ordmap.Box(1))

// Query by secondary index (exact match)
for pk, order := range indexed.GetBy(repo, ByCustomer, ordmap.Box("Alice")) {
	fmt.Println(order)
}

// Range scan on secondary index (Amount >= 150)
for pk, order := range indexed.From(repo, ByAmount, ordmap.Box(150)) {
	fmt.Println(order)
}
```

See [`go-indexed.md`](go-indexed.md) for the full design document.

## Examples

The [`examples`](examples) directory contains working programs demonstrating:

- [`basic`](examples/basic/main.go) — Standard usage with built-in types.
- [`composite_index`](examples/composite_index/main.go) — Custom keys and prefix/range scans.
- [`generational`](examples/generational/main.go) — Generational map with flush, shadowing, and tombstones.
- [`ptr_type`](examples/ptr_type/main.go) — Storing pointer values.

## Development

### Requirements

- Go 1.23+

### Testing

```sh
go test ./...
```

100% test coverage is expected on the core implementation (`avl.go`).

### Benchmarks

```sh
go test -bench=. -benchmem ./...
```

## License

MIT
