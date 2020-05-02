package ordmap

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

type myKey int

func (i *myKey) Cmp(i2 Key) int {
	return int(*i) - int(*(i2.(*myKey)))
}

// using a pointer based key where k1.Cmp(k2) == 0 does not imply k1 == k2 so we can fish out ==-based bugs with tests
func intKey(i int) Key {
	k := myKey(i)
	return &k
}

func TestIntKey(t *testing.T) {
	require.Equal(t, 0, intKey(1).Cmp(intKey(1)))
	require.Less(t, intKey(1).Cmp(intKey(2)), 0)
	require.Greater(t, intKey(5).Cmp(intKey(2)), 0)
}

type Model struct {
	t       *testing.T
	tree    *Node
	entries []Entry
	r       *rand.Rand
}

func NewModel(t *testing.T) *Model {
	m := &Model{
		t:       t,
		entries: []Entry{},
		r:       rand.New(rand.NewSource(0)),
	}
	m.checkInvariants()
	return m
}

func (m *Model) checkInvariants() {
	m.checkNodesValidity()
	m.checkBalance()
	m.checkElements()
	m.checkMinMax()
	m.checkIterator()
}

func (m *Model) checkNodesValidity() {
	var step func(*Node)
	step = func(n *Node) {
		if n == nil {
			return
		}
		require.GreaterOrEqual(m.t, n.order, uint8(1))
		require.LessOrEqual(m.t, n.order, uint8(MAX))
		for i := int(n.order); i < len(n.entries); i++ {
			require.Equal(m.t, zeroEntry, n.entries[i], fmt.Sprintf("%s: %d %v", n.visual(), n.order, n.entries))
			require.Nil(m.t, n.subtrees[i+1])
		}
		children := 0
		for i := 0; i <= int(n.order); i++ {
			if n.subtrees[i] != nil {
				children += 1
			}
		}
		if n.height == 1 {
			require.Equal(m.t, 0, children)
			return
		}
		require.Equal(m.t, int(n.order+1), children)
		for i := 0; i <= int(n.order); i++ {
			step(n.subtrees[i])
		}
	}
	step(m.tree)
}

func (m *Model) checkBalance() {
	var depth func(*Node) uint8
	depth = func(n *Node) uint8 {
		if n == nil {
			return 0
		}
		if n.height > 1 {
			for _, s := range n.subtrees[:n.order+1] {
				require.Equal(m.t, n.height-1, depth(s))
			}
		}
		return n.height
	}
	depth(m.tree)
}

func (m *Model) checkElements() {
	require.Equal(m.t, m.entries, m.tree.Entries(), m.tree.visual())
}

func (m *Model) checkIterator() {
	entries := make([]Entry, 0, len(m.entries))
	for i := m.tree.Iterate(); !i.Done(); i.Next() {
		entries = append(entries, i.Entry)
	}
	require.Equal(m.t, m.entries, entries)
}

func (m *Model) checkMinMax() {
	if len(m.entries) == 0 {
		require.Nil(m.t, m.tree.Min())
		require.Nil(m.t, m.tree.Max())
	} else {
		min := m.tree.Min()
		require.Equal(m.t, *min, m.entries[0])
		max := m.tree.Max()
		require.Equal(m.t, *max, m.entries[len(m.entries)-1])
	}
}

func (m *Model) Insert(key Key, value Value) {
	oldTree := m.tree
	oldEntries := oldTree.Entries()

	m.tree = m.tree.Insert(key, value)
	m.insertEntry(key, value)
	m.checkInvariants()

	require.Equal(m.t, oldEntries, oldTree.Entries(), "old tree changed") // persistence check
}

func (m *Model) insertEntry(key Key, value Value) {
	for i, e := range m.entries {
		if e.K.Cmp(key) == 0 {
			m.entries[i].V = value
			return
		}
	}
	m.entries = append(m.entries, Entry{key, value})
	sort.Slice(m.entries, func(i, j int) bool {
		return m.entries[i].K.Cmp(m.entries[j].K) < 0
	})
}

func (m *Model) Delete(key Key) {
	oldTree := m.tree
	oldEntries := oldTree.Entries()

	m.tree = m.tree.Remove(key)
	m.deleteEntry(key)
	m.checkInvariants()

	require.Equal(m.t, oldEntries, oldTree.Entries()) // persistence check
}

func (m *Model) deleteEntry(key Key) {
	for i, e := range m.entries {
		if e.K.Cmp(key) == 0 {
			copy(m.entries[i:], m.entries[i+1:])
			m.entries = m.entries[:len(m.entries)-1]
		}
	}
}

func TestModel(t *testing.T) {
	sizes := []int{10, 20, 30, 100} // , 400}
	for _, N := range sizes {
		t.Run(fmt.Sprintf("insert_%03d", N), func(t *testing.T) {
			m := NewModel(t)
			for i := 0; i < N; i++ {
				k := m.r.Intn(N)
				v := m.r.Intn(N)
				m.Insert(intKey(k), v)
			}
		})
	}
	sizes = []int{1, 3, 4, 5, 7, 8, 9, 11, 12, 13, 20, 30, 100} //, 400}
	for _, N := range sizes {
		t.Run(fmt.Sprintf("delete_%03d", N), func(t *testing.T) {
			m := NewModel(t)
			for i := 0; i < N; i++ {
				m.Insert(intKey(i), struct{}{})
			}
			for i := 0; i < N; i++ {
				k := m.r.Intn(N)
				m.Delete(intKey(k))
			}
		})
	}
}

func TestModelGrowing(t *testing.T) {
	N := 200
	m := NewModel(t)
	for i := 0; i < N; i++ {
		if rand.Float64() < 0.7 { // skewed so the tree can grow
			k := m.r.Intn(N)
			v := m.r.Intn(N)
			m.Insert(intKey(k), v)
		} else {
			k := m.r.Intn(N)
			m.Delete(intKey(k))
		}
	}
}
