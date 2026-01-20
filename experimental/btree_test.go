package ordmap

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

type myKey int

func (i *myKey) Less(i2 *myKey) bool {
	return *i < *i2
}

// using a pointer based key where k1.Cmp(k2) == 0 does not imply k1 == k2 so we can fish out ==-based bugs with tests
func intKey(i int) *myKey {
	k := myKey(i)
	return &k
}

func TestIntKey(t *testing.T) {
	require.False(t, intKey(1).Less(intKey(1)))
	require.True(t, intKey(1).Less(intKey(2)))
	require.False(t, intKey(5).Less(intKey(2)))
}

type Model struct {
	t       *testing.T
	tree    *OrdMap[*myKey, int]
	entries []Entry[*myKey, int]
	r       *rand.Rand
}

func NewModel(t *testing.T) *Model {
	m := &Model{
		t:       t,
		entries: []Entry[*myKey, int]{},
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
	var step func(*OrdMap[*myKey, int])
	step = func(n *OrdMap[*myKey, int]) {
		if n == nil {
			return
		}
		require.GreaterOrEqual(m.t, n.order, uint8(1))
		require.LessOrEqual(m.t, n.order, uint8(MAX))
		for i := int(n.order); i < len(n.entries); i++ {
			require.Equal(m.t, Entry[*myKey, int]{}, n.entries[i], fmt.Sprintf("%s: %d %v", n.visual(), n.order, n.entries))
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
	var depth func(*OrdMap[*myKey, int]) uint8
	depth = func(n *OrdMap[*myKey, int]) uint8 {
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
	allEntries := make([]Entry[*myKey, int], 0, len(m.entries))
	for k, v := range m.tree.All() {
		allEntries = append(allEntries, Entry[*myKey, int]{k, v})
	}
	require.Equal(m.t, m.entries, allEntries)

	backwardEntries := make([]Entry[*myKey, int], 0, len(m.entries))
	for k, v := range m.tree.Backward() {
		backwardEntries = append(backwardEntries, Entry[*myKey, int]{k, v})
	}
	reversed := make([]Entry[*myKey, int], len(m.entries))
	for i, e := range m.entries {
		reversed[len(m.entries)-1-i] = e
	}
	require.Equal(m.t, reversed, backwardEntries)

	if len(m.entries) == 0 {
		return
	}
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

func (m *Model) Insert(key *myKey, value int) {
	oldTree := m.tree
	oldEntries := oldTree.Entries()

	m.tree = m.tree.Insert(key, value)
	m.insertEntry(key, value)
	m.checkInvariants()

	require.Equal(m.t, oldEntries, oldTree.Entries(), "old tree changed") // persistence check
}

func (m *Model) insertEntry(key *myKey, value int) {
	for i, e := range m.entries {
		if e.K.Less(key) == false && key.Less(e.K) == false {
			m.entries[i].V = value
			return
		}
	}
	m.entries = append(m.entries, Entry[*myKey, int]{key, value})
	sort.Slice(m.entries, func(i, j int) bool {
		return m.entries[i].K.Less(m.entries[j].K)
	})
}

func (m *Model) Delete(key *myKey) {
	oldTree := m.tree
	oldEntries := oldTree.Entries()

	m.tree = m.tree.Remove(key)
	m.deleteEntry(key)
	m.checkInvariants()

	require.Equal(m.t, oldEntries, oldTree.Entries()) // persistence check
}

func (m *Model) deleteEntry(key *myKey) {
	for i, e := range m.entries {
		if e.K.Less(key) == false && key.Less(e.K) == false {
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
				m.Insert(intKey(i), i)
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

func TestFrom(t *testing.T) {
	t.Run("empty_tree", func(t *testing.T) {
		var tree *OrdMap[*myKey, int]
		count := 0
		for range tree.From(intKey(0)) {
			count++
		}
		require.Equal(t, 0, count)
	})

	t.Run("single_element", func(t *testing.T) {
		var tree *OrdMap[*myKey, int]
		tree = tree.Insert(intKey(5), 5)

		t.Run("start_below", func(t *testing.T) {
			entries := []Entry[*myKey, int]{}
			for k, v := range tree.From(intKey(0)) {
				entries = append(entries, Entry[*myKey, int]{k, v})
			}
			require.Equal(t, []Entry[*myKey, int]{{intKey(5), 5}}, entries)
		})

		t.Run("start_at", func(t *testing.T) {
			entries := []Entry[*myKey, int]{}
			for k, v := range tree.From(intKey(5)) {
				entries = append(entries, Entry[*myKey, int]{k, v})
			}
			require.Equal(t, []Entry[*myKey, int]{{intKey(5), 5}}, entries)
		})

		t.Run("start_above", func(t *testing.T) {
			entries := []Entry[*myKey, int]{}
			for k, v := range tree.From(intKey(10)) {
				entries = append(entries, Entry[*myKey, int]{k, v})
			}
			require.Empty(t, entries)
		})
	})

	t.Run("multiple_sizes", func(t *testing.T) {
		sizes := []int{1, 2, 3, 4, 5, 7, 8, 9, 11, 12, 13, 20, 30, 100}
		for _, N := range sizes {
			t.Run(fmt.Sprintf("size_%03d", N), func(t *testing.T) {
				m := NewModel(t)

				// Build a tree with consecutive keys
				for i := 0; i < N; i++ {
					m.Insert(intKey(i), i)
				}

				// Test starting from each key in the tree
				for startKey := 0; startKey < N; startKey++ {
					t.Run(fmt.Sprintf("start_at_%d", startKey), func(t *testing.T) {
						entries := []Entry[*myKey, int]{}
						for k, v := range m.tree.From(intKey(startKey)) {
							entries = append(entries, Entry[*myKey, int]{k, v})
						}
						expected := make([]Entry[*myKey, int], 0, N-startKey)
						for i := startKey; i < N; i++ {
							expected = append(expected, Entry[*myKey, int]{intKey(i), i})
						}
						require.Equal(t, expected, entries)
					})
				}

				// Test starting from keys between consecutive entries
				for startKey := 0; startKey < N-1; startKey++ {
					t.Run(fmt.Sprintf("start_between_%d_and_%d", startKey, startKey+1), func(t *testing.T) {
						entries := []Entry[*myKey, int]{}
						for k, v := range m.tree.From(intKey(startKey + 1)) {
							entries = append(entries, Entry[*myKey, int]{k, v})
						}
						expected := make([]Entry[*myKey, int], 0, N-startKey-1)
						for i := startKey + 1; i < N; i++ {
							expected = append(expected, Entry[*myKey, int]{intKey(i), i})
						}
						require.Equal(t, expected, entries)
					})
				}

				// Test starting before all entries
				t.Run("start_before_all", func(t *testing.T) {
					entries := []Entry[*myKey, int]{}
					for k, v := range m.tree.From(intKey(-1)) {
						entries = append(entries, Entry[*myKey, int]{k, v})
					}
					require.Equal(t, m.entries, entries)
				})

				// Test starting after all entries
				t.Run("start_after_all", func(t *testing.T) {
					entries := []Entry[*myKey, int]{}
					for k, v := range m.tree.From(intKey(N + 1)) {
						entries = append(entries, Entry[*myKey, int]{k, v})
					}
					require.Empty(t, entries)
				})
			})
		}
	})

	t.Run("random_operations", func(t *testing.T) {
		N := 200
		m := NewModel(t)
		for i := 0; i < N; i++ {
			if rand.Float64() < 0.7 {
				k := m.r.Intn(N)
				v := m.r.Intn(N)
				m.Insert(intKey(k), v)
			} else {
				k := m.r.Intn(N)
				m.Delete(intKey(k))
			}
		}

		if len(m.entries) == 0 {
			return
		}

		// Test From from various positions
		testPositions := []int{0, 1, len(m.entries) / 4, len(m.entries) / 2, len(m.entries) - 1}
		for _, idx := range testPositions {
			if idx >= len(m.entries) {
				continue
			}
			startKey := m.entries[idx].K
			t.Run(fmt.Sprintf("start_at_position_%d", idx), func(t *testing.T) {
				entries := []Entry[*myKey, int]{}
				for k, v := range m.tree.From(startKey) {
					entries = append(entries, Entry[*myKey, int]{k, v})
				}
				expected := m.entries[idx:]
				require.Equal(t, expected, entries)
			})
		}

		// Test starting from a key that doesn't exist but is between entries
		if len(m.entries) >= 2 {
			for i := 0; i < len(m.entries)-1; i++ {
				if m.entries[i].K.Less(m.entries[i+1].K) {
					// Create a key between these two entries
					val1 := int(*(m.entries[i].K))
					val2 := int(*(m.entries[i+1].K))
					if val2 > val1+1 {
						midKey := intKey(val1 + 1)
						t.Run(fmt.Sprintf("start_between_%d_and_%d", val1, val2), func(t *testing.T) {
							entries := []Entry[*myKey, int]{}
							for k, v := range m.tree.From(midKey) {
								entries = append(entries, Entry[*myKey, int]{k, v})
							}
							expected := m.entries[i+1:]
							require.Equal(t, expected, entries)
						})
					}
				}
			}
		}
	})
}

func TestBackwardFrom(t *testing.T) {
	t.Run("empty_tree", func(t *testing.T) {
		var tree *OrdMap[*myKey, int]
		count := 0
		for range tree.BackwardFrom(intKey(0)) {
			count++
		}
		require.Equal(t, 0, count)
	})

	t.Run("single_element", func(t *testing.T) {
		var tree *OrdMap[*myKey, int]
		tree = tree.Insert(intKey(5), 5)

		t.Run("start_below", func(t *testing.T) {
			entries := []Entry[*myKey, int]{}
			for k, v := range tree.BackwardFrom(intKey(0)) {
				entries = append(entries, Entry[*myKey, int]{k, v})
			}
			require.Empty(t, entries)
		})

		t.Run("start_at", func(t *testing.T) {
			entries := []Entry[*myKey, int]{}
			for k, v := range tree.BackwardFrom(intKey(5)) {
				entries = append(entries, Entry[*myKey, int]{k, v})
			}
			require.Equal(t, []Entry[*myKey, int]{{intKey(5), 5}}, entries)
		})

		t.Run("start_above", func(t *testing.T) {
			entries := []Entry[*myKey, int]{}
			for k, v := range tree.BackwardFrom(intKey(10)) {
				entries = append(entries, Entry[*myKey, int]{k, v})
			}
			require.Equal(t, []Entry[*myKey, int]{{intKey(5), 5}}, entries)
		})
	})

	t.Run("multiple_sizes", func(t *testing.T) {
		sizes := []int{1, 2, 3, 4, 5, 7, 8, 9, 11, 12, 13, 20, 30, 100}
		for _, N := range sizes {
			t.Run(fmt.Sprintf("size_%03d", N), func(t *testing.T) {
				m := NewModel(t)

				// Build a tree with consecutive keys
				for i := 0; i < N; i++ {
					m.Insert(intKey(i), i)
				}

				// Test starting from each key in the tree
				for startKey := 0; startKey < N; startKey++ {
					t.Run(fmt.Sprintf("start_at_%d", startKey), func(t *testing.T) {
						entries := []Entry[*myKey, int]{}
						for k, v := range m.tree.BackwardFrom(intKey(startKey)) {
							entries = append(entries, Entry[*myKey, int]{k, v})
						}
						expected := make([]Entry[*myKey, int], 0, startKey+1)
						for i := startKey; i >= 0; i-- {
							expected = append(expected, Entry[*myKey, int]{intKey(i), i})
						}
						require.Equal(t, expected, entries)
					})
				}

				// Test starting from keys between consecutive entries
				for startKey := 0; startKey < N-1; startKey++ {
					t.Run(fmt.Sprintf("start_between_%d_and_%d", startKey, startKey+1), func(t *testing.T) {
						entries := []Entry[*myKey, int]{}
						for k, v := range m.tree.BackwardFrom(intKey(startKey + 1)) {
							entries = append(entries, Entry[*myKey, int]{k, v})
						}
						expected := make([]Entry[*myKey, int], 0, startKey+2)
						for i := startKey + 1; i >= 0; i-- {
							expected = append(expected, Entry[*myKey, int]{intKey(i), i})
						}
						require.Equal(t, expected, entries)
					})
				}

				// Test starting before all entries
				t.Run("start_before_all", func(t *testing.T) {
					entries := []Entry[*myKey, int]{}
					for k, v := range m.tree.BackwardFrom(intKey(-1)) {
						entries = append(entries, Entry[*myKey, int]{k, v})
					}
					require.Empty(t, entries)
				})

				// Test starting after all entries
				t.Run("start_after_all", func(t *testing.T) {
					entries := []Entry[*myKey, int]{}
					for k, v := range m.tree.BackwardFrom(intKey(N + 1)) {
						entries = append(entries, Entry[*myKey, int]{k, v})
					}
					expected := make([]Entry[*myKey, int], 0, N)
					for i := N - 1; i >= 0; i-- {
						expected = append(expected, Entry[*myKey, int]{intKey(i), i})
					}
					require.Equal(t, expected, entries)
				})
			})
		}
	})

	t.Run("random_operations", func(t *testing.T) {
		N := 200
		m := NewModel(t)
		for i := 0; i < N; i++ {
			if rand.Float64() < 0.7 {
				k := m.r.Intn(N)
				v := m.r.Intn(N)
				m.Insert(intKey(k), v)
			} else {
				k := m.r.Intn(N)
				m.Delete(intKey(k))
			}
		}

		if len(m.entries) == 0 {
			return
		}

		// Test BackwardFrom from various positions
		testPositions := []int{0, 1, len(m.entries) / 4, len(m.entries) / 2, len(m.entries) - 1}
		for _, idx := range testPositions {
			if idx >= len(m.entries) {
				continue
			}
			startKey := m.entries[idx].K
			t.Run(fmt.Sprintf("start_at_position_%d", idx), func(t *testing.T) {
				entries := []Entry[*myKey, int]{}
				for k, v := range m.tree.BackwardFrom(startKey) {
					entries = append(entries, Entry[*myKey, int]{k, v})
				}
				expected := make([]Entry[*myKey, int], idx+1)
				for i := idx; i >= 0; i-- {
					expected[idx-i] = m.entries[i]
				}
				require.Equal(t, expected, entries)
			})
		}

		// Test starting from a key that doesn't exist but is between entries
		if len(m.entries) >= 2 {
			for i := 0; i < len(m.entries)-1; i++ {
				if m.entries[i].K.Less(m.entries[i+1].K) {
					// Create a key between these two entries
					val1 := int(*(m.entries[i].K))
					val2 := int(*(m.entries[i+1].K))
					if val2 > val1+1 {
						midKey := intKey(val1 + 1)
						t.Run(fmt.Sprintf("start_between_%d_and_%d", val1, val2), func(t *testing.T) {
							entries := []Entry[*myKey, int]{}
							for k, v := range m.tree.BackwardFrom(midKey) {
								entries = append(entries, Entry[*myKey, int]{k, v})
							}
							expected := make([]Entry[*myKey, int], i+1)
							for j := i; j >= 0; j-- {
								expected[i-j] = m.entries[j]
							}
							require.Equal(t, expected, entries)
						})
					}
				}
			}
		}
	})
}
