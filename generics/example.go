package main

import "fmt"

func main() {
	var m1 = NewOrdMap[int, string](func(i, j int) bool { return i < j })
	fmt.Println(m1.Entries())
	m1 = m1.Insert(1, "foo") // adding entries
	fmt.Println(m1.Entries())
	m1 = m1.Insert(2, "baz")
	fmt.Println(m1.Entries())
	m1 = m1.Insert(2, "bar") // will override
	fmt.Println(m1.Entries())
	fmt.Println(m1.Get(2)) // access by key
	m1 = m1.Insert(3, "baz")
	fmt.Println(m1.Entries())
	m1 = m1.Remove(1) // can also remove entries
	fmt.Println(m1.Entries())
}
