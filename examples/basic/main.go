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
	for k, v := range m1.All() {
		fmt.Println(k, v)
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
	for k, v := range m2.Backward() {
		fmt.Println(k, v)
	}
	fmt.Println(m2.Min(), m2.Max()) // access the extremes
}
