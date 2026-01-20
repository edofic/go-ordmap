package ordmap

import (
	"fmt"
	"testing"

	"github.com/edofic/go-ordmap/v2"
)

func BenchmarkComparison(b *testing.B) {
	for _, M := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("%v", M), func(b *testing.B) {
			b.Run("avl", func(b *testing.B) {
				tree := ordmap.NewBuiltin[int, struct{}]()
				for i := 0; i < M; i++ {
					tree = tree.Insert(i, struct{}{})
				}
				b.Run("Get", func(b *testing.B) {
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
			b.Run("btree", func(b *testing.B) {
				var tree *OrdMap[*myKey, int]
				for i := 0; i < M; i++ {
					tree = tree.Insert(intKey(i), i)
				}
				b.Run("Get", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Get(intKey(5))
					}
				})
				b.Run("Insert", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Insert(intKey(M+1), M+1)
					}
				})
				b.Run("Remove", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Remove(intKey(i % M))
					}
				})
			})
		})
	}
}
