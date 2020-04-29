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
	return n.removeStep(key, true)
}

func (n *Node234) removeStep(key int, allowMinimal bool) *Node234 {
	if !allowMinimal && n.order <= 1 {
		panic("remove called on a minimal node: " + n.visual())
	}
	if n.leaf {
		for i := 0; i < int(n.order); i++ {
			if n.keys[i] == key {
				n = n.dup()
				copy(n.keys[i:], n.keys[i+1:])
				n.order -= 1
				if n.order == 0 {
					return nil
				}
				return n
			}
		}
		return n
	} else {
		n = n.dup()
		index := int(n.order)
		for i := 0; i < int(n.order); i++ {
			if n.keys[i] == key {
				child, min := n.subtrees[i+1].popMin()
				n.subtrees[i+1] = child
				n.keys[i] = min
				return n
			}
			if key < n.keys[i] {
				index = i
				break
			}
		}
		if n.subtrees[index].order == 1 {
			if index == 0 { // grab from the right
				if n.subtrees[1].order > 1 {
					child := n.subtrees[index].dup()
					neighbour := n.subtrees[1].dup()
					nk := neighbour.keys[0]
					child.keys[1] = n.keys[index]
					n.keys[index] = nk
					child.subtrees[2] = neighbour.subtrees[0]
					copy(neighbour.keys[:], neighbour.keys[1:])
					copy(neighbour.subtrees[:], neighbour.subtrees[1:])
					child.order += 1
					neighbour.order -= 1
					n.subtrees[0] = child
					n.subtrees[1] = neighbour
					return n.removeStep(key, allowMinimal)
				} else { // right neighbour is minimal
					child := n.subtrees[index]
					neighbour := n.subtrees[1]
					newChild := &Node234{
						order: child.order + neighbour.order + 1,
						leaf:  child.leaf, // == neighbour.leaf
					}
					copy(newChild.keys[:], child.keys[:child.order])
					newChild.keys[child.order] = n.keys[index]
					copy(newChild.keys[child.order+1:], neighbour.keys[:neighbour.order])
					n.subtrees[index] = newChild
					copy(n.subtrees[1:], n.subtrees[2:])
					copy(n.keys[0:], n.keys[1:])
					n.order -= 1
				}
			} else {
				child := n.subtrees[index]
				neighbour := n.subtrees[index-1]
				if neighbour.order > 1 {
					child = child.dup()
					neighbour = neighbour.dup()
					n.subtrees[index] = child
					n.subtrees[index-1] = neighbour
					copy(child.keys[1:], child.keys[:child.order])
					copy(child.subtrees[1:], child.subtrees[:child.order+1])
					child.order += 1
					child.keys[0] = n.keys[index-1]
					child.subtrees[0] = neighbour.subtrees[neighbour.order]
					n.keys[index-1] = neighbour.keys[neighbour.order-1]
					neighbour.order -= 1
				} else {
					newChild := &Node234{
						order: child.order + neighbour.order + 1,
						leaf:  child.leaf, // == neighbour.leaf
					}
					copy(newChild.keys[:], neighbour.keys[:neighbour.order])
					newChild.keys[neighbour.order] = n.keys[index-1]
					copy(newChild.keys[neighbour.order+1:], child.keys[:child.order])
					copy(n.subtrees[:index-2], n.subtrees[:index])
					n.subtrees[index-1] = newChild
					n.order -= 1
					index -= 1
				}
			}
		}
		n.subtrees[index] = n.subtrees[index].removeStep(key, false)
		return n
	}
}

func (n *Node234) insertNonFull(key int) *Node234 {
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

func (n *Node234) popMin() (*Node234, int) {
	if n.order == 1 {
		panic("popping from minimal")
	}
	n = n.dup()
	if n.leaf {
		k := n.keys[0]
		copy(n.keys[:], n.keys[1:])
		n.order -= 1
		return n, k
	}
	// TODO merge
	child, min := n.subtrees[0].popMin()
	n.subtrees[0] = child
	return n, min
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
