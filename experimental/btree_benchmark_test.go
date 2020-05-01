package ordmap

import (
	"fmt"
	"testing"
)

//go:generate go run github.com/edofic/go-ordmap/cmd/gen -name AvlIntSet -key int -less "<" -value struct{} -target ./avl_int_set.go -pkg ordmap
// to benchmark against

func BenchmarkAgainstAvl(b *testing.B) {
	for _, M := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("%v", M), func(b *testing.B) {
			b.Run("avl", func(b *testing.B) {
				var tree *AvlIntSet
				for i := 0; i < M; i++ {
					tree = tree.Insert(i, struct{}{})
				}
				b.Run("Contains", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Get(5)
					}
				})
				b.Run("Insert", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Insert(M+1, struct{}{})
					}
				})
				b.Run("Remove", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Remove(i % M)
					}
				})
			})
			b.Run("234", func(b *testing.B) {
				var tree *Node
				for i := 0; i < M; i++ {
					tree = tree.Insert(i)
				}
				b.Run("Contains", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Contains(5)
					}
				})
				b.Run("Insert", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Insert(M + 1)
					}
				})
				b.Run("Remove", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Remove(i % M)
					}
				})
			})
		})
	}
}
