package ordmap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic234(t *testing.T) {
	N := 5
	require.True(t, true)
	var n *Node234
	fmt.Println(n.visual())
	elems := []int{}
	for i := 0; i < N; i++ {
		elems = append(elems, i)
		n = n.Insert(i)
		//fmt.Println(n.visual())
	}
	require.Equal(t, elems, n.Keys())
	for i := -N / 2; i < 3*N/2; i++ {
		shouldContain := i >= 0 && i < N
		require.Equal(t, shouldContain, n.Contains(i), i)
	}

	toDelete := []int{4}
	for _, e := range toDelete {
		for i := 0; i < len(elems); i++ {
			if elems[i] == e {
				copy(elems[i:], elems[i+1:])
				elems = elems[:len(elems)-1]
				break
			}
		}
		n = n.Remove(e)
		fmt.Println(n.visual())
		require.Equal(t, elems, n.Keys(), e)
	}
}
