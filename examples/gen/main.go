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
	var m1 *IntStrMap
	m1 = m1.Insert(intKey(1), "foo")
	m1 = m1.Insert(intKey(2), "baz")
	m1 = m1.Insert(intKey(2), "bar")
	m1 = m1.Insert(intKey(3), "baz")
	for i := m1.Iterate(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}

	var m2 *IntIntMap
	m2 = m2.Insert(intKey(1), 1)
	m2 = m2.Insert(intKey(2), 3)
	m2 = m2.Insert(intKey(2), 2)
	m2 = m2.Insert(intKey(3), 3)
	for i := m2.IterateReverse(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
}
