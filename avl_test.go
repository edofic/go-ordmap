package main

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func eq(k1, k2 Key) bool {
	return !k1.Less(k2) && !k2.Less(k1)
}

type intKey int

func (k intKey) Less(k2 Key) bool { return k < k2.(intKey) }

func reprTree(n *Node) string {
	if n == nil {
		return "_"
	}
	if n.children[0] == nil && n.children[1] == nil {
		return fmt.Sprintf("(%v)", n.Entry.Key)
	}
	return fmt.Sprintf("[%v %v %v]", reprTree(n.children[0]), n.Entry.Key, reprTree(n.children[1]))
}

func validateHeight(t *testing.T, tree *Node) {
	if tree == nil {
		return // empty is balanced
	}
	left := tree.children[0].Height()
	right := tree.children[1].Height()
	require.Contains(t, []int{-1, 0, 1}, right-left)
	require.Equal(t, combinedDepth(tree.children[0], tree.children[1]), tree.h)
	validateHeight(t, tree.children[0])
	validateHeight(t, tree.children[1])
}

func validateOrdered(t *testing.T, tree *Node) {
	if tree == nil {
		return
	}
	key := tree.Entry.Key
	if tree.children[0] != nil {
		leftKey := tree.children[0].Entry.Key
		require.True(t, leftKey.Less(key) || eq(key, leftKey))
		validateOrdered(t, tree.children[0])
	}
	if tree.children[1] != nil {
		rightKey := tree.children[1].Entry.Key
		require.True(t, key.Less(rightKey) || eq(key, rightKey))
		validateOrdered(t, tree.children[1])
	}
}

type TreeModel struct {
	t     *testing.T
	elems []Entry
	tree  *Node
	debug bool
}

func NewTreeModel(t *testing.T) *TreeModel {
	return &TreeModel{
		t:     t,
		elems: make([]Entry, 0),
		tree:  nil,
		debug: false, // toggle this for verbose tests
	}
}

func (m *TreeModel) Len() int {
	return len(m.elems)
}

func (m *TreeModel) Insert(value int) {
	key := intKey(value)
	index := -1
	for i, elem := range m.elems {
		if eq(elem.Key, key) {
			index = i
			break
		}
	}
	if index == -1 { // not found
		m.elems = append(m.elems, Entry{key, 0})
		sort.Slice(m.elems, func(i, j int) bool {
			return m.elems[i].Key.Less(m.elems[j].Key)
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
}

func (m *TreeModel) InsertAll(values ...int) {
	for _, value := range values {
		m.Insert(value)
	}
}

func (m *TreeModel) Remove(value int) {
	key := intKey(value)
	// find
	index := -1
	for i, candidate := range m.elems {
		if candidate.Key == key {
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
	key := m.elems[index].Key
	value := int(key.(intKey))
	m.Remove(value)
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
	samples := make(map[int]*Node)
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
	var tree *Node
	tree = tree.Insert(intKey(1), "foo")
	tree = tree.Insert(intKey(2), "bar")
	tree = tree.Insert(intKey(0), "bar")

	value, ok := tree.Get(intKey(1))
	require.True(t, ok)
	require.Equal(t, "foo", value)

	value, ok = tree.Get(intKey(0))
	require.True(t, ok)
	require.Equal(t, "bar", value)

	value, ok = tree.Get(intKey(3))
	require.False(t, ok)
	require.Nil(t, value)
}

func TestMinMax(t *testing.T) {
	var tree *Node
	require.Nil(t, tree.Min())
	require.Nil(t, tree.Max())

	tree = tree.Insert(intKey(1), "foo")
	tree = tree.Insert(intKey(2), "bar")
	tree = tree.Insert(intKey(3), "baz")

	require.Equal(t, &Entry{Key: intKey(1), Value: "foo"}, tree.Min())
	require.Equal(t, &Entry{Key: intKey(3), Value: "baz"}, tree.Max())
}

func TestRemoveMissing(t *testing.T) {
	var tree *Node
	tree = tree.Insert(intKey(1), "foo")
	tree = tree.Insert(intKey(2), "bar")
	require.Equal(t, 2, tree.Len())
	tree.Remove(intKey(0))
	require.Equal(t, 2, tree.Len())
}

func TestIterator(t *testing.T) {
	var tree *Node
	N := 100
	for i := 0; i < N; i++ {
		tree = tree.Insert(intKey(i), i)
	}

	valuesFromEntries := make([]int, N)
	for i, entry := range tree.Entries() {
		valuesFromEntries[i] = entry.Value.(int)
	}

	valuesFromIterator := make([]int, 0, N)
	for iter := tree.Iterator(); !iter.Done(); iter.Next() {
		valuesFromIterator = append(valuesFromIterator, iter.Value().(int))
	}
	require.Equal(t, valuesFromEntries, valuesFromIterator)
}

func TestIteratorReverse(t *testing.T) {
	var tree *Node
	N := 100
	for i := 0; i < N; i++ {
		tree = tree.Insert(intKey(i), i)
	}

	valuesFromEntries := make([]int, N)
	for i, entry := range tree.Entries() {
		valuesFromEntries[N-i-1] = entry.Value.(int)
	}

	valuesFromIterator := make([]int, 0, N)
	for iter := tree.IteratorReverse(); !iter.Done(); iter.Next() {
		valuesFromIterator = append(valuesFromIterator, iter.Value().(int))
	}
	require.Equal(t, valuesFromEntries, valuesFromIterator)
}

func BenchmarkMap(b *testing.B) {
	for _, M := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("%v", M), func(b *testing.B) {
			m := make(map[intKey]int)
			for i := 0; i < M; i++ {
				m[intKey(i)] = i
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m[intKey(M+1)] = M + 1
				delete(m, intKey(M+1))
			}
		})
	}
}

func BenchmarkTree(b *testing.B) {
	for _, M := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("%v", M), func(b *testing.B) {
			var tree *Node
			for i := 0; i < M; i++ {
				tree = tree.Insert(intKey(i), i)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tree = tree.Insert(intKey(M+1), M+1)
				tree = tree.Remove(intKey(M + 1))
			}
		})
	}
}
