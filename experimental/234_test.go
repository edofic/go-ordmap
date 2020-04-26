package ordmap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBasic234(t *testing.T) {
	require.True(t, true)
	var n *Node234
	elems := []int{1, 2, 3}
	for _, e := range elems {
		n = n.Insert(e)
	}
	require.ElementsMatch(t, elems, n.Keys())
}
