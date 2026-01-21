package generational

import (
	"testing"

	"github.com/edofic/go-ordmap/v2"
)

const (
	benchInitialSize = 100_000
	benchYoungLimit  = 100
)

func BenchmarkChurnGenerational(b *testing.B) {
	m := New[Int, int](benchYoungLimit)
	for i := 0; i < benchInitialSize; i++ {
		m = m.Insert(Int(i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newKey := Int(benchInitialSize + i)
		m = m.Insert(newKey, i)
		m = m.Remove(newKey)
	}
}

func BenchmarkChurnAVL(b *testing.B) {
	m := ordmap.New[Int, int]()
	for i := 0; i < benchInitialSize; i++ {
		m = m.Insert(Int(i), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newKey := Int(benchInitialSize + i)
		m = m.Insert(newKey, i)
		m = m.Remove(newKey)
	}
}
