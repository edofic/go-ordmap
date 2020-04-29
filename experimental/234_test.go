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
	elems := []int{}
	for i := 0; i < 100; i++ {
		elems = append(elems, i)
		n = n.Insert(i)
		fmt.Println(n.visual())
	}
	require.Equal(t, elems, n.Keys())
}
