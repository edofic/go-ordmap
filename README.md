# go-ordmap

[![main](https://github.com/edofic/go-ordmap/actions/workflows/main.yml/badge.svg?branch=v2)](https://github.com/edofic/go-ordmap/actions/workflows/main.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/edofic/go-ordmap)](https://goreportcard.com/report/github.com/edofic/go-ordmap)
[![GoDoc](https://godoc.org/github.com/gopherjs/vecty?status.svg)](https://godoc.org/github.com/edofic/go-ordmap)

Persistent generic ordered maps for Go.

This is `v2` which is based on 1.18 generics. If you can't (or don't want to)
use generics there is still `v1` which uses code generation. See [v1
branch](https://github.com/edofic/go-ordmap/tree/v1).

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
go get github.com/edofic/go-ordmap/v2@v2.0.0-beta2
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
	m1 := ordmap.NewBuiltin[int, string]()

	m1 = m1.Insert(1, "foo") // adding entries
	m1 = m1.Insert(2, "baz")
	m1 = m1.Insert(2, "bar") // will override
	fmt.Println(m1.Get(2))   // access by key
	m1 = m1.Insert(3, "baz")
	// this is how you can efficiently iterate in-order
	for i := m1.Iterate(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
	m1 = m1.Remove(1)         // can also remove entries
	fmt.Println(m1.Entries()) // or get a slice of all of them

	// can use another map of different type in the same package
	m2 := ordmap.NewBuiltin[int, int]()
	v, ok := m2.Get(0)
	fmt.Println("wat", v, ok)
	m2 = m2.Insert(1, 1) // this one has "raw" ints for keys
	m2 = m2.Insert(2, 3) // in order to support this you will also need to pass
	m2 = m2.Insert(2, 2) // `-less "<"` to the genreeator in order to use
	m2 = m2.Insert(3, 3) // native comparator
	// can iterate in reverse as well
	for i := m2.IterateReverse(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
	fmt.Println(m2.Min(), m2.Max()) // access the extremes
}
```

See [examples](https://github.com/edofic/go-ordmap/blob/v2/examples) for more.

You will need to provide the `Less` method on your  - so the map knows how to
order itself. Or if you want to use one of the builtin types (e.g. `int`) you
can use `NewBulitin` which only takes supported types.`

```go
func (k MyKey) Less(k2 MyKey) bool {
    ...
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
