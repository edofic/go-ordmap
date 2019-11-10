package main

import (
	"fmt"
	"github.com/edofic/go-ordmap"
)

type intKey int

func (i intKey) Less(other ordmap.Key) bool {
	// unsafe, this can blow up at runtime if you mix key types
	return int(i) < int(other.(intKey))
}

func main() {
	var m *ordmap.OrdMap
	m = m.Insert(intKey(1), "foo")
	m = m.Insert(intKey(2), "baz")
	m = m.Insert(intKey(2), "bar")
	m = m.Insert(intKey(3), "baz")
	for i := m.Iterate(); !i.Done(); i.Next() {
		// value here is interface{}
		fmt.Println(i.GetKey(), i.GetValue())
	}
}
