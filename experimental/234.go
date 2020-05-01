package ordmap

import (
	"sort"
	"strconv"
)

const MAX = 3 // must be odd

type Node234 struct {
	order    uint8 // 1..MAX
	leaf     bool
	keys     [MAX]int
	subtrees [MAX + 1]*Node234
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
	finger := n
OUTER:
	for finger != nil {
		for i := 0; i < int(finger.order); i++ {
			k := finger.keys[i]
			if k == key {
				return true
			}
			if key < k {
				finger = finger.subtrees[i]
				continue OUTER
			}
		}
		finger = finger.subtrees[finger.order]
	}
	return false
}

func (n *Node234) Insert(key int) *Node234 {
	if n == nil {
		return &Node234{order: 1, leaf: true, keys: [MAX]int{key, 0, 0}}
	}
	if n.order == MAX { // full root, need to split
		left, k2, right := n.split()
		n = &Node234{
			order: 1,
			leaf:  false,
			keys:  [MAX]int{k2, 0, 0},
		}
		n.subtrees[0] = left
		n.subtrees[1] = right
	} else {
		n = n.dup()
	}
	n.insertNonFullMut(key)
	return n
}

func (n *Node234) Remove(key int) *Node234 {
	if !n.Contains(key) {
		return n
	}
	return n.removeStep(key)
}

func (n *Node234) removeStep(key int) *Node234 {
	if n.leaf {
		for i := 0; i < int(n.order); i++ {
			if n.keys[i] == key {
				n = n.dup()
				copy(n.keys[i:], n.keys[i+1:])
				n.keys[n.order-1] = 0
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
				index = n.ensureChildNotMinimal(i + 1)
				if n.order == 0 { // degenerated, need to drop a level
					return n.subtrees[0].removeStep(key)
				}
				if index != i+1 { // merge happened
					return n.removeStep(key) // easiest to try again
				}
				if n.keys[i] != key { // rotaiton happened
					return n.removeStep(key) // easiest to try again
				}
				child, min := n.subtrees[index].popMin()
				n.subtrees[i+1] = child
				n.keys[i] = min
				return n
			}
			if key < n.keys[i] {
				index = i
				break
			}
		}
		index = n.ensureChildNotMinimal(index)
		if n.order == 0 { // degenerated, need to drop a level
			return n.subtrees[0].removeStep(key)
		}
		n.subtrees[index] = n.subtrees[index].removeStep(key)
		return n
	}
}

func (n *Node234) insertNonFullMut(key int) {
OUTER:
	for {
		for i := 0; i < int(n.order); i++ {
			if n.keys[i] == key {
				return
			}
		}
		if n.leaf {
			keys := n.keys
			keys[n.order] = key
			sort.Ints(keys[:n.order+1])
			n.order += 1
			n.keys = keys
			return
		}
		index := 0
		for i := 0; i < int(n.order); i++ {
			if key > n.keys[i] {
				index = i + 1
			}
		}
		child := n.subtrees[index]
		if child.order == MAX { // full, need to split before entering
			left, key1, right := child.split()
			for i := int(n.order); i > index; i-- {
				n.keys[i] = n.keys[i-1]
			}
			n.keys[index] = key1
			for i := int(n.order); i > index; i-- {
				n.subtrees[i+1] = n.subtrees[i]
			}
			n.subtrees[index] = left
			n.subtrees[index+1] = right
			n.order += 1
			if key < key1 {
				n = left
				continue OUTER
			} else if key == key1 {
				// nothing to do
				return
			} else {
				n = right
				continue OUTER
			}
		} else {
			n.subtrees[index] = n.subtrees[index].dup()
			n = n.subtrees[index]
			continue OUTER
		}
	}
}

func (n *Node234) ensureChildNotMinimal(index int) int {
	if n.subtrees[index].order > 1 {
		return index
	}
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
			neighbour.keys[neighbour.order] = 0
			n.subtrees[0] = child
			n.subtrees[1] = neighbour
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
			copy(newChild.subtrees[:], child.subtrees[:child.order+1])
			copy(newChild.subtrees[child.order+1:], neighbour.subtrees[:neighbour.order+1])
			n.subtrees[index] = newChild
			copy(n.subtrees[1:], n.subtrees[2:])
			copy(n.keys[0:], n.keys[1:])
			n.subtrees[n.order] = nil
			n.order -= 1
			n.keys[n.order] = 0
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
			neighbour.keys[neighbour.order] = 0
		} else {
			newChild := &Node234{
				order: child.order + neighbour.order + 1,
				leaf:  child.leaf, // == neighbour.leaf
			}
			copy(newChild.keys[:], neighbour.keys[:neighbour.order])
			newChild.keys[neighbour.order] = n.keys[index-1]
			copy(newChild.keys[neighbour.order+1:], child.keys[:child.order])
			copy(newChild.subtrees[:], neighbour.subtrees[:neighbour.order+1])
			copy(newChild.subtrees[neighbour.order+1:], child.subtrees[:child.order+1])
			copy(n.subtrees[index-1:], n.subtrees[index:])
			n.subtrees[n.order] = nil
			n.subtrees[index-1] = newChild
			copy(n.keys[index-1:], n.keys[index:])
			n.order -= 1
			n.keys[n.order] = 0
			index -= 1
		}
	}
	return index
}

func (n *Node234) split() (left *Node234, key int, right *Node234) {
	key = n.keys[(MAX-1)/2]
	left = &Node234{
		order: (MAX - 1) / 2,
		leaf:  n.leaf,
	}
	for i := 0; i < (MAX-1)/2; i++ {
		left.keys[i] = n.keys[i]
	}
	for i := 0; i <= (MAX-1)/2; i++ {
		left.subtrees[i] = n.subtrees[i]
	}
	right = &Node234{
		order: (MAX - 1) / 2,
		leaf:  n.leaf,
	}
	for i := (MAX + 1) / 2; i < MAX; i++ {
		right.keys[i-(MAX+1)/2] = n.keys[i]
	}
	for i := (MAX + 1) / 2; i <= MAX; i++ {
		right.subtrees[i-(MAX+1)/2] = n.subtrees[i]
	}
	return
}

func (n *Node234) popMin() (nn *Node234, res int) {
	if n.order == 1 {
		panic("popping from minimal")
	}
	n = n.dup()
	if n.leaf {
		k := n.keys[0]
		copy(n.keys[:], n.keys[1:])
		n.order -= 1
		n.keys[n.order] = 0
		return n, k
	}
	_ = n.ensureChildNotMinimal(0)
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
