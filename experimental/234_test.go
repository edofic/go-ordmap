package ordmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
