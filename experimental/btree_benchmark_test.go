package ordmap

import (
	"fmt"
	"testing"

	"github.com/edofic/go-ordmap/v2"
)

type benchKey int

func (i benchKey) Less(i2 benchKey) bool {
	return i < i2
}

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
				b.Run("All", func(b *testing.B) {
					count := 0
					b.ReportAllocs()
					for range b.N {
						for range tree.All() {
							count++
						}
					}
					benchCount = count
				})
				b.Run("All5", func(b *testing.B) {
					count := 0
					b.ReportAllocs()
					for range b.N {
						seen := 0
						for range tree.All() {
							count++
							seen++
							if seen >= 5 {
								break
							}
						}
					}
					benchCount = count
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
			b.Run("btree_value", func(b *testing.B) {
				var tree *OrdMap[benchKey, int]
				for i := 0; i < M; i++ {
					tree = tree.Insert(benchKey(i), i)
				}
				b.Run("Get", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Get(benchKey(5))
					}
				})
				b.Run("Insert", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Insert(benchKey(M+1), M+1)
					}
				})
				b.Run("Remove", func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						tree.Remove(benchKey(i % M))
					}
				})
				b.Run("All", func(b *testing.B) {
					sum := 0
					b.ReportAllocs()
					for range b.N {
						for _, value := range tree.All() {
							sum += value
						}
					}
					benchCount = sum
				})
				b.Run("All5", func(b *testing.B) {
					sum := 0
					b.ReportAllocs()
					for range b.N {
						seen := 0
						for _, value := range tree.All() {
							sum += value
							seen++
							if seen >= 5 {
								break
							}
						}
					}
					benchCount = sum
				})
			})
		})
	}
}

var benchCount int
