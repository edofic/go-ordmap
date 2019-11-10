//go:generate go run github.com/edofic/go-ordmap/cmd/gen -name IntStrMap -key intKey -value string -target ./int_str_map.go
package main

import (
	"fmt"
)

type intKey int

func (i intKey) Less(other intKey) bool {
	return int(i) < int(other)
}

func main() {
	var m *IntStrMap
	m = m.Insert(intKey(1), "foo")
	m = m.Insert(intKey(2), "baz")
	m = m.Insert(intKey(2), "bar")
	m = m.Insert(intKey(3), "baz")
	for i := m.Iterate(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
}
