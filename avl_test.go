package ordmap

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func eq[K Comparable[K]](k1, k2 K) bool {
	if k1.Less(k2) {
		return false
	}
	if k2.Less(k1) {
		return false
	}
	return true
}

func reprTree[K Comparable[K], V any](n *Node[K, V]) string {
	if n == nil {
		return "_"
	}
	if n.children[0] == nil && n.children[1] == nil {
		return fmt.Sprintf("(%v)", n.entry.K)
	}
	return fmt.Sprintf("[%v %v %v]", reprTree(n.children[0]), n.entry.K, reprTree(n.children[1]))
}

func validateHeight[K Comparable[K], V any](t *testing.T, tree *Node[K, V]) {
	if tree == nil {
		return // empty is balanced
	}
	left := tree.children[0].height()
	right := tree.children[1].height()
	require.Contains(t, []int{-1, 0, 1}, right-left)
	require.Equal(t, combinedDepth(tree.children[0], tree.children[1]), tree.h)
	validateHeight(t, tree.children[0])
	validateHeight(t, tree.children[1])
}

func validateOrdered[K Comparable[K], V any](t *testing.T, root *Node[K, V]) {
	var step func(tree *Node[K, V])
	step = func(tree *Node[K, V]) {
		if tree == nil {
			return
		}
		key := tree.entry.K
		if tree.children[0] != nil {
			leftKey := tree.children[0].entry.K
			require.True(t, leftKey.Less(key) || eq(key, leftKey))
			step(tree.children[0])
		}
		if tree.children[1] != nil {
			rightKey := tree.children[1].entry.K
			require.True(t, key.Less(rightKey) || eq(key, rightKey))
			step(tree.children[1])
		}
	}
	step(root)
}

func requireIterEq[K Comparable[K], V any](t *testing.T, i1, i2 Iterator[K, V]) {
	require.Equal(t, i1.direction, i2.direction)
	// skip `less` as it's not pertinent to iterator behavior once constructed
	require.Equal(t, i1.stack, i2.stack)
	require.Equal(t, i1.currentEntry, i2.currentEntry)
}

type TreeModel struct {
	t     *testing.T
	elems []Entry[Builtin[int], int]
	tree  *Node[Builtin[int], int]
	debug bool
}

func NewTreeModel(t *testing.T) *TreeModel {
	return &TreeModel{
		t:     t,
		elems: make([]Entry[Builtin[int], int], 0),
		debug: false, // toggle this for verbose tests
	}
}

func (m *TreeModel) Len() int {
	return len(m.elems)
}

func (m *TreeModel) Insert(value int) {
	key := Builtin[int]{value}
	index := -1
	for i, elem := range m.elems {
		if elem.K == key {
			index = i
			break
		}
	}
	if index == -1 { // not found
		m.elems = append(m.elems, Entry[Builtin[int], int]{key, 0})
		sort.Slice(m.elems, func(i, j int) bool {
			return m.elems[i].K.Less(m.elems[j].K)
		})
	}
	if m.debug {
		fmt.Println("Inserting key", key, "into", reprTree(m.tree))
	}
	m.tree = m.tree.Insert(key, 0)
	if m.debug {
		fmt.Println(reprTree(m.tree))
	}
	validateHeight(m.t, m.tree)
	validateOrdered(m.t, m.tree)
	require.Equal(m.t, m.elems, m.tree.Entries())
	require.Equal(m.t, len(m.elems), m.tree.Len())
}

func (m *TreeModel) InsertAll(values ...int) {
	for _, value := range values {
		m.Insert(value)
	}
}

func (m *TreeModel) Remove(value int) {
	key := Builtin[int]{value}
	// find
	index := -1
	for i, candidate := range m.elems {
		if candidate.K == key {
			index = i
			break
		}
	}
	if index == -1 {
		m.t.Fatalf("could not find key %v", key)
	}
	// delete
	copy(m.elems[index:], m.elems[index+1:])
	m.elems = m.elems[:len(m.elems)-1]
	if m.debug {
		fmt.Println("Remove value", key, "from", reprTree(m.tree))
	}
	m.tree = m.tree.Remove(key)
	if m.debug {
		fmt.Println(reprTree(m.tree))
	}
	validateHeight(m.t, m.tree)
	validateOrdered(m.t, m.tree)
	require.Equal(m.t, m.elems, m.tree.Entries())
}

func (m *TreeModel) RemoveRandom() bool {
	if len(m.elems) == 0 {
		return false
	}
	index := rand.Intn(len(m.elems))
	key := m.elems[index].K
	m.Remove(key.value)
	return true
}

func TestModel(t *testing.T) {
	model := NewTreeModel(t)
	model.InsertAll(4, 2, 7, 6, 6, 9)
	require.Equal(t, 5, model.Len())
	model.Remove(4)
	model.Remove(6)
	require.Equal(t, 3, model.Len())
}

func TestModelRandom(t *testing.T) {
	N := 100
	model := NewTreeModel(t)
	samples := make(map[int]*Node[Builtin[int], int])
	sizes := make(map[int]int)
	for i := 0; i < N; i++ {
		if rand.Float64() < 0.7 { // skewed so the tree can grow
			model.Insert(rand.Intn(N))
		} else {
			model.RemoveRandom()
		}
		samples[i] = model.tree
		sizes[i] = model.tree.Len()
	}

	// check persistence
	for i, sample := range samples {
		require.Equal(t, sizes[i], sample.Len())
	}
}

func TestGet(t *testing.T) {
	var tree *Node[Builtin[int], string]

	value, ok := tree.Get(Builtin[int]{0})
	require.False(t, ok)
	require.Equal(t, "", value)

	tree = tree.Insert(Builtin[int]{1}, "foo")
	tree = tree.Insert(Builtin[int]{2}, "bar")
	tree = tree.Insert(Builtin[int]{0}, "bar")

	value, ok = tree.Get(Builtin[int]{1})
	require.True(t, ok)
	require.Equal(t, "foo", value)

	value, ok = tree.Get(Builtin[int]{0})
	require.True(t, ok)
	require.Equal(t, "bar", value)

	value, ok = tree.Get(Builtin[int]{3})
	require.False(t, ok)
	require.Equal(t, "", value)
}

func TestMinMax(t *testing.T) {
	var tree *Node[Builtin[int], string]
	require.Nil(t, tree.Min())
	require.Nil(t, tree.Max())

	tree = tree.Insert(Builtin[int]{1}, "foo")
	tree = tree.Insert(Builtin[int]{2}, "bar")
	tree = tree.Insert(Builtin[int]{3}, "baz")

	require.Equal(t, &Entry[Builtin[int], string]{K: Builtin[int]{1}, V: "foo"}, tree.Min())
	require.Equal(t, &Entry[Builtin[int], string]{K: Builtin[int]{3}, V: "baz"}, tree.Max())
}

func TestRemoveMissing(t *testing.T) {
	var tree *Node[Builtin[int], string]
	tree = tree.Insert(Builtin[int]{1}, "foo")
	tree = tree.Insert(Builtin[int]{2}, "bar")
	require.Equal(t, 2, tree.Len())
	tree.Remove(Builtin[int]{0})
	require.Equal(t, 2, tree.Len())
}

func TestIteratorEmpty(t *testing.T) {
	var tree *Node[Builtin[int], string]
	count := 0
	for iter := tree.Iterate(); !iter.Done(); iter.Next() {
		count += 1
	}
	require.Equal(t, 0, count)
}

func TestIterator(t *testing.T) {
	var tree *Node[Builtin[int], int]
	N := 100
	for i := 0; i < N; i++ {
		tree = tree.Insert(Builtin[int]{i}, i)
	}

	valuesFromEntries := make([]int, N)
	for i, entry := range tree.Entries() {
		valuesFromEntries[i] = entry.V
	}

	keysFromIterator := make([]int, 0, N)
	valuesFromIterator := make([]int, 0, N)
	for iter := tree.Iterate(); !iter.Done(); iter.Next() {
		keysFromIterator = append(keysFromIterator, iter.GetKey().value)
		valuesFromIterator = append(valuesFromIterator, iter.GetValue())
	}
	require.Equal(t, valuesFromEntries, keysFromIterator)
	require.Equal(t, valuesFromEntries, valuesFromIterator)
}

func TestIteratorBuiltin(t *testing.T) {
	var tree NodeBuiltin[int, int]
	N := 100
	for i := 0; i < N; i++ {
		tree = tree.Insert(i, i)
	}

	valuesFromEntries := make([]int, N)
	for i, entry := range tree.Entries() {
		valuesFromEntries[i] = entry.V
	}

	keysFromIterator := make([]int, 0, N)
	valuesFromIterator := make([]int, 0, N)
	for iter := tree.Iterate(); !iter.Done(); iter.Next() {
		keysFromIterator = append(keysFromIterator, iter.GetKey())
		valuesFromIterator = append(valuesFromIterator, iter.GetValue())
	}
	require.Equal(t, valuesFromEntries, keysFromIterator)
	require.Equal(t, valuesFromEntries, valuesFromIterator)
}

func TestIteratorReverse(t *testing.T) {
	var tree *Node[Builtin[int], int]
	N := 100
	for i := 0; i < N; i++ {
		tree = tree.Insert(Builtin[int]{i}, i)
	}

	valuesFromEntries := make([]int, N)
	for i, entry := range tree.Entries() {
		valuesFromEntries[N-i-1] = entry.V
	}

	valuesFromIterator := make([]int, 0, N)
	for iter := tree.IterateReverse(); !iter.Done(); iter.Next() {
		valuesFromIterator = append(valuesFromIterator, iter.GetValue())
	}
	require.Equal(t, valuesFromEntries, valuesFromIterator)
}

func TestIterateFrom(t *testing.T) {
	var tree *Node[Builtin[int], int]
	N := 100
	for i := 0; i < N; i++ {
		tree = tree.Insert(Builtin[int]{i}, i)
	}

	t.Run("forward range", func(t *testing.T) {
		valuesFromIterator := make([]int, 0, N)
		for iter := tree.IterateFrom(Builtin[int]{37}); !iter.Done(); iter.Next() {
			value := iter.GetValue()
			if value >= 42 {
				break
			}
			valuesFromIterator = append(valuesFromIterator, value)
		}
		require.Equal(t, []int{37, 38, 39, 40, 41}, valuesFromIterator)
	})

	t.Run("forward whole", func(t *testing.T) {
		valuesFromIterator := make([]int, 0, N)
		for iter := tree.IterateFrom(Builtin[int]{0}); !iter.Done(); iter.Next() {
			value := iter.GetValue()
			valuesFromIterator = append(valuesFromIterator, value)
		}
		require.Len(t, valuesFromIterator, 100)
	})

	t.Run("reverse range", func(t *testing.T) {
		valuesFromIterator := make([]int, 0, N)
		for iter := tree.IterateReverseFrom(Builtin[int]{41}); !iter.Done(); iter.Next() {
			value := iter.GetValue()
			if value < 37 {
				break
			}
			valuesFromIterator = append(valuesFromIterator, value)
		}
		require.Equal(t, []int{41, 40, 39, 38, 37}, valuesFromIterator)
	})

	t.Run("reverse whole", func(t *testing.T) {
		valuesFromIterator := make([]int, 0, N)
		for iter := tree.IterateReverseFrom(Builtin[int]{100}); !iter.Done(); iter.Next() {
			value := iter.GetValue()
			valuesFromIterator = append(valuesFromIterator, value)
		}
		require.Len(t, valuesFromIterator, 100)
	})

	t.Run("internal state", func(t *testing.T) {
		for iter := tree.Iterate(); !iter.Done(); iter.Next() {
			key := iter.GetKey()
			iterFrom := tree.IterateFrom(key)
			requireIterEq(t, iter, iterFrom)
		}
	})

	t.Run("internal state reverse", func(t *testing.T) {
		for iter := tree.IterateReverse(); !iter.Done(); iter.Next() {
			key := iter.GetKey()
			iterFrom := tree.IterateReverseFrom(key)
			requireIterEq(t, iter, iterFrom)
		}
	})
}

func TestEmptyLen(t *testing.T) {
	var empty *Node[Builtin[int], int]
	require.Equal(t, 0, empty.Len())
}

func BenchmarkMap(b *testing.B) {
	for _, M := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("%v", M), func(b *testing.B) {
			m := make(map[int]int)
			for i := 0; i < M; i++ {
				m[i] = i
			}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m[M+1] = M + 1
				delete(m, M+1)
			}
		})
	}
}

func BenchmarkTree(b *testing.B) {
	for _, M := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("%v", M), func(b *testing.B) {
				var tree *Node[Builtin[int], int]
			for i := 0; i < M; i++ {
				tree = tree.Insert(Builtin[int]{i}, i)
			}
			b.Run("InsertRemove", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					tree1 := tree.Insert(Builtin[int]{M+1}, M+1)
					_ = tree1.Remove(Builtin[int]{M+1})
				}
			})
			b.Run("Entries", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					tree.Entries()
				}
			})
			b.Run("Iterator", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					for iter := tree.Iterate(); !iter.Done(); iter.Next() {
						// no-op, just consume
					}
				}
			})
			b.Run("Iterator5", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					count := 0
					for iter := tree.Iterate(); !iter.Done() && count < 5; iter.Next() {
						count += 1
					}
				}
			})
			b.Run("Min", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					tree.Min()
				}
			})
			b.Run("Get", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					tree.Get(Builtin[int]{5})
				}
			})
		})
	}
}
