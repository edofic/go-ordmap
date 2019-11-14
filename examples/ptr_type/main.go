//go:generate go run github.com/edofic/go-ordmap/cmd/gen -name MyMap -key int -less "<" -value *MyValue -target ./my_map.go
package main

import (
	"fmt"
)

type MyValue struct {
	foo int
}

func main() {
	var m *MyMap
	m = m.Insert(1, &MyValue{1})
	m = m.Insert(2, &MyValue{2})
	m = m.Insert(2, &MyValue{2})
	m = m.Insert(3, &MyValue{3})
	for i := m.Iterate(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue().foo)
	}
}
