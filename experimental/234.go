package ordmap

import "sort"

type Node234 struct {
	order    uint8 // 1, 2, 3
	leaf     bool
	keys     [3]int
	subtrees [4]*Node234
}

func (n *Node234) Keys() []int {
	keys := make([]int, 0)
	var step func(n *Node234)
	step = func(n *Node234) {
		if n == nil {
			return
		}
		step(n.subtrees[0])
		for i := uint8(0); i < n.order; i++ {
			keys = append(keys, n.keys[i])
			step(n.subtrees[i+1])
		}
	}
	step(n)
	return keys
}

func (n *Node234) Insert(key int) *Node234 {
	if n == nil {
		return &Node234{1, true, [3]int{key, 0, 0}, [4]*Node234{nil, nil, nil, nil}}
	}
	if n.leaf && n.order < 3 {
		keys := n.keys
		keys[n.order] = key
		sort.Ints(keys[:n.order+1])
		return &Node234{n.order + 1, true, keys, [4]*Node234{nil, nil, nil, nil}}
	}
	panic("TODO")
}
