package main

import (
	"fmt"

	"github.com/edofic/go-ordmap/v2"
)

type MyValue struct {
	foo int
}

func main() {
	m := ordmap.NewBuiltin[int, *MyValue]()
	m = m.Insert(1, &MyValue{1})
	m = m.Insert(2, &MyValue{2})
	m = m.Insert(2, &MyValue{2})
	m = m.Insert(3, &MyValue{3})
	for k, v := range m.All() {
		fmt.Println(k, v.foo)
	}
}
