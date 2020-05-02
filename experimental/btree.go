package ordmap

import (
	"strconv"
)

const MAX = 5 // must be odd

type Node struct {
	order    uint8 // 1..MAX
	leaf     bool
	keys     [MAX]int
	subtrees [MAX + 1]*Node
}

func (n *Node) Keys() []int {
	keys := make([]int, 0)
	var step func(n *Node)
	step = func(n *Node) {
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

func (n *Node) Contains(key int) bool {
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

func (n *Node) Insert(key int) *Node {
	if n == nil {
		return &Node{order: 1, leaf: true, keys: [MAX]int{key, 0, 0}}
	}
	if n.order == MAX { // full root, need to split
		left, k2, right := n.split()
		n = &Node{
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

func (n *Node) Remove(key int) *Node {
	if !n.Contains(key) {
		return n
	}
	if n.leaf && n.order == 1 && n.keys[0] == key {
		return nil
	}
	n = n.dup()
	n.removeStepMut(key)
	return n
}

func (n *Node) removeStepMut(key int) {
OUTER:
	for {
		if n.leaf {
			for i := 0; i < int(n.order); i++ {
				if n.keys[i] == key {
					top := int(n.order) - 1
					for j := i; j < top; j++ {
						n.keys[j] = n.keys[j+1]
					}
					n.keys[n.order-1] = 0
					n.order -= 1
					return
				}
			}
			return
		} else {
			index := int(n.order)
			for i := 0; i < int(n.order); i++ {
				if n.keys[i] == key { // inner delete
					index = n.ensureChildNotMinimal(i + 1)
					if n.order == 0 { // degenerated, need to drop a level
						*n = *n.subtrees[0]
						continue OUTER
					}
					if index != i+1 || n.keys[i] != key { // merge OR rotation
						continue OUTER // easiest to try again
					}
					child := n.subtrees[index].dup()
					min := child.popMinMut()
					n.subtrees[index] = child
					n.keys[i] = min
					return
				}
				if key < n.keys[i] {
					index = i
					break
				}
			}
			index = n.ensureChildNotMinimal(index)
			if n.order == 0 { // degenerated, need to drop a level
				*n = *n.subtrees[0]
				continue OUTER
			}
			n.subtrees[index] = n.subtrees[index].dup()
			n = n.subtrees[index]
			continue OUTER
		}
	}
}

func (n *Node) insertNonFullMut(key int) {
OUTER:
	for {
		for i := 0; i < int(n.order); i++ {
			if n.keys[i] == key {
				return
			}
		}
		if n.leaf {
			n.keys[n.order] = key
			n.order += 1
			for i := int(n.order) - 1; i > 0; i-- {
				if n.keys[i] < n.keys[i-1] {
					n.keys[i], n.keys[i-1] = n.keys[i-1], n.keys[i]
				} else {
					break
				}
			}
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

func (n *Node) ensureChildNotMinimal(index int) int {
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
			newChild := &Node{
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
			newChild := &Node{
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

func (n *Node) split() (left *Node, key int, right *Node) {
	key = n.keys[(MAX-1)/2]
	left = &Node{
		order: (MAX - 1) / 2,
		leaf:  n.leaf,
	}
	for i := 0; i < (MAX-1)/2; i++ {
		left.keys[i] = n.keys[i]
	}
	for i := 0; i <= (MAX-1)/2; i++ {
		left.subtrees[i] = n.subtrees[i]
	}
	right = &Node{
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

func (n *Node) popMinMut() int {
OUTER:
	for {
		if n.leaf {
			k := n.keys[0]
			for i := 1; i < int(n.order); i++ {
				n.keys[i-1] = n.keys[i]
			}
			n.order -= 1
			n.keys[n.order] = 0
			return k
		}
		_ = n.ensureChildNotMinimal(0)
		n.subtrees[0] = n.subtrees[0].dup()
		n = n.subtrees[0]
		continue OUTER
	}
}

func (n Node) dup() *Node {
	return &n
}

func (n *Node) visual() string {
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
