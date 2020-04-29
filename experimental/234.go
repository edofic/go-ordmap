package ordmap

import (
	"sort"
	"strconv"
)

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

func (n *Node234) Contains(key int) bool {
	if n == nil {
		return false
	}
	for i := 0; i < int(n.order); i++ {
		k := n.keys[i]
		if k == key {
			return true
		}
		if key < k {
			return n.subtrees[i].Contains(key)
		}
	}
	return n.subtrees[n.order].Contains(key)
}

func (n *Node234) Insert(key int) *Node234 {
	if n == nil {
		return &Node234{1, true, [3]int{key, 0, 0}, [4]*Node234{nil, nil, nil, nil}}
	}
	if n.order == 3 { // full root, need to split
		left, key, right := n.split()
		n = &Node234{
			order:    1,
			leaf:     false,
			keys:     [3]int{key, 0, 0},
			subtrees: [4]*Node234{left, right, nil, nil},
		}
	}
	return n.insertNonFull(key)
}

func (n *Node234) Remove(key int) *Node234 {
	if n == nil {
		return n
	}
	if n.leaf {
		for i := 0; i < int(n.order); i++ {
			if n.keys[i] == key {
				n = n.dup()
				copy(n.keys[i:], n.keys[i+1:])
				n.order -= 1
				return n
			}
		}
	}
	return n
}

func (n *Node234) insertNonFull(key int) *Node234 {
	if n.order == 3 {
		panic("insertNonFull called on a full node: " + n.visual())
	}
	if n.leaf {
		keys := n.keys
		keys[n.order] = key
		sort.Ints(keys[:n.order+1])
		return &Node234{n.order + 1, true, keys, [4]*Node234{nil, nil, nil, nil}}
	}
	index := 0
	for i := 0; i < int(n.order); i++ {
		if key > n.keys[i] {
			index = i + 1
		}
	}
	n = n.dup()
	child := n.subtrees[index]
	if child.order == 3 { // full, need to split before entering
		left, key1, right := child.split()
		copy(n.keys[index+1:], n.keys[index:])
		n.keys[index] = key1
		copy(n.subtrees[index+2:], n.subtrees[index+1:])
		n.order += 1
		if key < key1 {
			left = left.insertNonFull(key)
		} else {
			right = right.insertNonFull(key)
		}
		n.subtrees[index] = left
		n.subtrees[index+1] = right
	} else {
		n.subtrees[index] = n.subtrees[index].insertNonFull(key)
	}
	return n
}

func (n *Node234) split() (left *Node234, key int, right *Node234) {
	key = n.keys[1]
	left = &Node234{
		order:    1,
		leaf:     n.leaf,
		keys:     [3]int{n.keys[0], 0, 0},
		subtrees: [4]*Node234{n.subtrees[0], n.subtrees[1], nil, nil},
	}
	right = &Node234{
		order:    1,
		leaf:     n.leaf,
		keys:     [3]int{n.keys[2], 0, 0},
		subtrees: [4]*Node234{n.subtrees[2], n.subtrees[3], nil, nil},
	}
	return
}

func (n Node234) dup() *Node234 {
	return &n
}

func (n *Node234) visual() string {
	if n == nil {
		return "_"
	}
	s := "[ " + n.subtrees[0].visual()
	for i := 0; i < int(n.order); i++ {
		s += " " + strconv.Itoa(n.keys[i]) + " " + n.subtrees[i+1].visual()
	}
	s += " ]"
	return s
}
