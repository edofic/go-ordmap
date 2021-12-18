# go-ordmap

[![main](https://github.com/edofic/go-ordmap/actions/workflows/main.yml/badge.svg?branch=v2)](https://github.com/edofic/go-ordmap/actions/workflows/main.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/edofic/go-ordmap)](https://goreportcard.com/report/github.com/edofic/go-ordmap)
[![GoDoc](https://godoc.org/github.com/gopherjs/vecty?status.svg)](https://godoc.org/github.com/edofic/go-ordmap)

Persistent generic ordered maps for Go.

This is the branch with `v2` which is based on 1.18 generics. Currently a separate branch as 1.18 is still in beta, will become the default once 1.18.0 is released.
See `master` branch for non-generics version (based on code generation).

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

```sh
go get github.com/edofic/go-ordmap/v2@v2.0.0-beta1
```

You only need to remember to always assign the returned value of all
operations; the original object is never updated in-place:

```go
package main

import (
	"fmt"

	"github.com/edofic/go-ordmap/v2"
)

func main() {
	m1 := ordmap.New[int, string](ordmap.Less[int])
	m1 = m1.Insert(1, "foo") // adding entries
	m1 = m1.Insert(2, "baz")
	m1 = m1.Insert(2, "bar") // will override
	fmt.Println(m1.Get(2))   // access by key
	m1 = m1.Insert(3, "baz")
	// this is how you can efficiently iterate in-order
	for i := m1.Iterate(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
	m1 = m1.Remove(1) // can also remove entries
	fmt.Println(m1.Entries()) // or get a slice of all of them
	// can iterate in reverse as well
	for i := m1.IterateReverse(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
	fmt.Println(m1.Min(), m1.Max()) // access the extremes
}
```

See [examples](https://github.com/edofic/go-ordmap/blob/v2/examples) for more.

You will need to provide the `Less` function - so the map knows how to order
itself. There is a generic `ordmap.Less` function that is available for all types that support the `<` operator, but for custom types you will need to provide your own, e.g.

```go
func compareMyKey(k1, k2 MyKey) bool {
    return k1.v < k2.v
}
```

## Development

Go 1.18+ required.

### Testing

Standard testing

```sh
go test ./...
```

100% test coverage expected on the core implementation (`avl.go`)
