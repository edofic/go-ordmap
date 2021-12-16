# go-ordmap

[![main pipeline](https://github.com/edofic/go-ordmap/actions/workflows/main.yml/badge.svg)](https://github.com/edofic/go-ordmap/actions/workflows/main.yml)
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

In order to facilitate safe API and efficient internalization the actual code
is actually generated from a template and specialised to your types. Think
"generics" - inspired by [genny](https://github.com/cheekybits/genny) but
tweaked to this particular case.

## Usage

The recommended usage is through code generation. Place this comment in one of
your `.go` files

```go
//go:generate go run github.com/edofic/go-ordmap/cmd/gen -name IntStrMap -key intKey -value string -target ./int_str_map.go
```

*NOTE* go-ormap uses `go:embed` internally so this requires at least Go 1.16
(otherwise you will get a compilation failure)

And run `go generate ./...`  ([more about generation](https://blog.golang.org/generate))

You need to provide the types (if not stdlib) that are referenced. Mostly
you'll need to implement the key type

```go
type intKey int

func (i intKey) Less(other intKey) bool {
	return int(i) < int(other)
}
```

You will need to provide the `Less` method - so the map knows how to order
itself. Then you can use the generated map type.

### Using the map

You only need to remember to always assign the returned value of all
operations; the original object is never updated in-place:

```go
var m1 *IntStrMap                // zero value is the empty map
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
[examples/gen/main.go](https://github.com/edofic/go-ordmap/blob/master/examples/gen/main.go)
for a fully functioning example.

### Using raw version

If for some reason you don't want to use code generation you can directly use
the template implementation. This less efficient and less safe as it uses an
interface for keys and an empty interface for values.

```go
import "github.com/edofic/go-ordmap"
```

Key will now need to compare against the `ordmap.Key` interface type (less safe)

```go
type intKey int

func (i intKey) Less(other ordmap.Key) bool {
	// unsafe, this can blow up at runtime if you mix key types
	return int(i) < int(other.(intKey))
}
```

Usage then is virtual the same - except for the `interface{}` values

```go
var m *ordmap.OrdMap
m = m.Insert(intKey(1), "foo")
m = m.Insert(intKey(2), "baz")
m = m.Insert(intKey(2), "bar")
m = m.Insert(intKey(3), "baz")
for i := m.Iterate(); !i.Done(); i.Next() {
    // value here is interface{}
    fmt.Println(i.GetKey(), i.GetValue())
}
```

See
[examples/raw/main.go](https://github.com/edofic/go-ordmap/blob/master/examples/raw/main.go)
for a fully functioning example.

## Development

Go 1.16+ required due to use of `go:embed`.

### Generation

If you're changing the template `avl.go` make sure to run `go generate ./...` to
update generated code in examples.

### Testing

Standard testing

```sh
go test ./...
```

100% test coverage expected on the template (`avl.go`)
