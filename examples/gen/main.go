//go:generate go run github.com/edofic/go-ordmap/cmd/gen -name IntStrMap -key intKey -value string -target ./int_str_map.go
//go:generate go run github.com/edofic/go-ordmap/cmd/gen -name IntIntMap -key intKey -value int -target ./int_int_map.go
package main

import (
	"fmt"
)

type intKey int

func (i intKey) Less(other intKey) bool {
	return int(i) < int(other)
}

func main() {
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

	// can use another map of different type in the same package
	var m2 *IntIntMap
	m2 = m2.Insert(intKey(1), 1)
	m2 = m2.Insert(intKey(2), 3)
	m2 = m2.Insert(intKey(2), 2)
	m2 = m2.Insert(intKey(3), 3)
	// can iterate in reverse as well
	for i := m2.IterateReverse(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
	fmt.Println(m2.Min(), m2.Max()) // access the extremes
}
