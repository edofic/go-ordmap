package ordmap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic234(t *testing.T) {
	require.True(t, true)
	var n *Node234
	fmt.Println(n.visual())
	elems := []int{1, 2, 3, 4, 5}
	for _, e := range elems {
		n = n.Insert(e)
		fmt.Println(n.visual())
	}
	require.Equal(t, elems, n.Keys())
}
