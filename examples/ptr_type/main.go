package main

import (
	"fmt"

	"github.com/edofic/go-ordmap/v2"
)

type MyValue struct {
	foo int
}

func main() {
	m := ordmap.New[int, *MyValue](ordmap.Less[int])
	m = m.Insert(1, &MyValue{1})
	m = m.Insert(2, &MyValue{2})
	m = m.Insert(2, &MyValue{2})
	m = m.Insert(3, &MyValue{3})
	for i := m.Iterate(); !i.Done(); i.Next() {
		fmt.Println(i.GetKey(), i.GetValue().foo)
	}
}
