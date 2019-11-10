package main

import (
	"fmt"
	"testing"
)

func BenchmarkSpecialised(b *testing.B) {
	for _, M := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("%v", M), func(b *testing.B) {
			var m *IntIntMap
			for i := 0; i < M; i++ {
				m = m.Insert(intKey(i), i)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m = m.Insert(intKey(M+1), M+1)
				m = m.Remove(intKey(M + 1))
			}
		})
	}
}
