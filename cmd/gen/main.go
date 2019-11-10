//go:generate go run github.com/go-bindata/go-bindata/go-bindata ../../avl.go
package main

import "fmt"

func main() {
	template := string(MustAsset("../../avl.go"))
	fmt.Println(template)
}
