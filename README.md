# go-ordmap

[![CircleCI](https://circleci.com/gh/edofic/go-ordmap.svg?style=svg)](https://circleci.com/gh/edofic/go-ordmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/edofic/go-ordmap)](https://goreportcard.com/report/github.com/edofic/go-ordmap)
[![GoDoc](https://godoc.org/github.com/gopherjs/vecty?status.svg)](https://godoc.org/github.com/edofic/go-ordmap)

Persistent generic ordered maps for Go.

## Rationale

Standard library does not provide a key-value store data structure that would
also allow quick access to minimum/maximum and in-order iteration.

This package provides such data structure (called `OrdMap` for "ordered map")
and implements it on top of [AVL trees](https://en.wikipedia.org/wiki/AVL_tree)
(self balancing binary trees). This allows us `O(log n)` read/write access to
arbitrary key as well as to the min & max while still supporting `O(n)`
iteration over elements - just that it's now in order.

One detail that may not be idiomatic for Go is that these `OrdMap`s are
[persistent data
structures](https://en.wikipedia.org/wiki/Persistent_data_structure) - meaning
that each operation gives you back a new map while keeping the old intact. This
is beneficial when you have many concurrent readers - your writer can advance
while the readers still traverse the old versions (kind of similar to
[MVCC](https://en.wikipedia.org/wiki/Multiversion_concurrency_control))

In order to facilitate safe API and efficient internalization the this module ueses type parameters and thus requires go 1.18+.

## Usage

You will need to provide the `Less` function - so the map knows how to order
itself. There is a generic `ordmap.Less` function that is available for all types that support the `<` operator, but for custom types you will need to provide your own, e.g.

```go
func compareMyKey(k1, k2 MyKey) bool {
    return k1.v < k2.v
}
```

### Using the map

You only need to remember to always assign the returned value of all
operations; the original object is never updated in-place:

```go
m1 := ordmap.New[int, string](ordmap.Less[int])
m1 = m1.Insert(intKey(1), "foo") // adding entries
m1 = m1.Insert(intKey(2), "baz")
m1 = m1.Insert(intKey(2), "bar") // will override
fmt.Println(m1.Get(intKey(2)))   // access by key
m1 = m1.Insert(intKey(3), "baz")
// this is how you can efficiently iterate in-order
for i := m1.Iterate(); !i.Done(); i.Next() {
    fmt.Println(i.GetKey(), i.GetValue())
}
m1 = m1.Remove(intKey(1)) // can also remove entries
fmt.Println(m1.Entries()) // or get a slice of all of them
// can iterate in reverse as well
for i := m1.IterateReverse(); !i.Done(); i.Next() {
    fmt.Println(i.GetKey(), i.GetValue())
}
fmt.Println(m1.Min(), m1.Max()) // access the extremes
```

See
[examples/basic/main.go](https://github.com/edofic/go-ordmap/blob/master/examples/basic/main.go)
for a fully functioning example.


## Development

Go 1.18+ required.

### Testing

Standard testing

```sh
go test ./...
```

100% test coverage expected on the core implementation (`avl.go`)
