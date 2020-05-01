package ordmap

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

type Model234 struct {
	t     *testing.T
	tree  *Node234
	elems []int
	r     *rand.Rand
}

func NewModel234(t *testing.T) *Model234 {
	m := &Model234{
		t:     t,
		elems: []int{},
		r:     rand.New(rand.NewSource(0)),
	}
	m.checkInvariants()
	return m
}

func (m *Model234) checkInvariants() {
	m.checkNodesValidity()
	m.checkBalance()
	m.checkElements()
}

func (m *Model234) checkNodesValidity() {
	var step func(*Node234)
	step = func(n *Node234) {
		if n == nil {
			return
		}
		require.GreaterOrEqual(m.t, n.order, uint8(1))
		require.LessOrEqual(m.t, n.order, uint8(3))
		for i := int(n.order); i < len(n.keys); i++ {
			require.Equal(m.t, 0, n.keys[i], fmt.Sprintf("%s: %d %v", n.visual(), n.order, n.keys))
			//require.Nil(m.t, n.subtrees[i+1])
		}
		children := 0
		for i := 0; i <= int(n.order); i++ {
			if n.subtrees[i] != nil {
				children += 1
			}
		}
		if n.leaf {
			require.Equal(m.t, children, 0)
			return
		}
		require.Greater(m.t, children, 0)
		for i := 0; i <= int(n.order); i++ {
			step(n.subtrees[i])
		}
	}
	step(m.tree)
}

func (m *Model234) checkBalance() {
	var depth func(*Node234) int
	depth = func(n *Node234) int {
		if n == nil {
			return 0
		}
		if n.leaf {
			return 1
		}
		d1 := depth(n.subtrees[0])
		for _, n = range n.subtrees[1 : n.order+1] {
			require.Equal(m.t, d1, depth(n))
		}
		return d1
	}
	depth(m.tree)
}

func (m *Model234) checkElements() {
	require.Equal(m.t, m.elems, m.tree.Keys())
}

func (m *Model234) Insert(key int) {
	m.tree = m.tree.Insert(key)
	m.insertElems(key)
	m.checkInvariants()
}

func (m *Model234) insertElems(key int) {
	for _, e := range m.elems {
		if e == key {
			return
		}
	}
	m.elems = append(m.elems, key)
	sort.Ints(m.elems)
}

func (m *Model234) Delete(key int) {
	m.tree = m.tree.Remove(key)
	m.deleteElems(key)
	m.checkInvariants()
}

func (m *Model234) deleteElems(key int) {
	for i, e := range m.elems {
		if e == key {
			copy(m.elems[i:], m.elems[i+1:])
			m.elems = m.elems[:len(m.elems)-1]
		}
	}
}

func TestBasic234(t *testing.T) {
	N := 7

	require.True(t, true)
	var n *Node234
	elems := []int{}
	for i := 0; i < N; i++ {
		elems = append(elems, i)
		n = n.Insert(i)
	}
	for i := -N / 2; i < 3*N/2; i++ {
		shouldContain := i >= 0 && i < N
		require.Equal(t, shouldContain, n.Contains(i), i)
	}

	toDelete := []int{}
	for _, e := range toDelete {
		for i := 0; i < len(elems); i++ {
			if elems[i] == e {
				copy(elems[i:], elems[i+1:])
				elems = elems[:len(elems)-1]
				break
			}
		}
		n = n.Remove(e)
		require.Equal(t, elems, n.Keys(), e)
	}
}

func TestStealLeft(t *testing.T) {
	var n *Node234
	for _, e := range []int{0, 10, 20, 1, 2} {
		n = n.Insert(e)
	}
	n = n.Remove(20)
	require.Equal(t, []int{0, 1, 2, 10}, n.Keys())
}

func TestModel(t *testing.T) {
	sizes := []int{10, 20, 30, 100} // , 400}
	for _, N := range sizes {
		t.Run(fmt.Sprintf("insert_%03d", N), func(t *testing.T) {
			m := NewModel234(t)
			for i := 0; i < N; i++ {
				e := m.r.Intn(N)
				m.Insert(e)
			}
		})
	}
	sizes = []int{1, 3, 4, 5, 7, 8, 9, 11, 12, 13, 20, 30} // , 100, 400}
	for _, N := range sizes {
		t.Run(fmt.Sprintf("delete_%03d", N), func(t *testing.T) {
			m := NewModel234(t)
			for i := 0; i < N; i++ {
				m.Insert(i)
			}
			for i := 0; i < N; i++ {
				e := m.r.Intn(N)
				m.Delete(e)
			}
		})
	}
}
