package ordmap

import "fmt"

const MAX = 5 // must be odd

type Key interface {
	Cmp(Key) int
}

type Value interface{}

type Entry struct {
	K Key
	V Value
}

var zeroEntry Entry

type Node struct {
	order    uint8 // 1..MAX
	height   uint8
	entries  [MAX]Entry
	subtrees [MAX + 1]*Node
}

func (n *Node) Entries() []Entry {
	entries := make([]Entry, 0) // TODO preallocate when Len is implemented
	var step func(n *Node)
	step = func(n *Node) {
		if n == nil {
			return
		}
		step(n.subtrees[0])
		for i := uint8(0); i < n.order; i++ {
			entries = append(entries, n.entries[i])
			step(n.subtrees[i+1])
		}
	}
	step(n)
	return entries
}

func (n *Node) Get(key Key) (value Value, ok bool) {
	finger := n
OUTER:
	for finger != nil {
		for i := 0; i < int(finger.order); i++ {
			entry := finger.entries[i]
			cmp := key.Cmp(entry.K)
			if cmp == 0 {
				return entry.V, true
			}
			if cmp < 0 { // key < entry.K
				finger = finger.subtrees[i]
				continue OUTER
			}
		}
		finger = finger.subtrees[finger.order]
	}
	return value, false
}

func (n *Node) Insert(key Key, value Value) *Node {
	if n == nil {
		n = &Node{order: 1, height: 1}
		n.entries[0] = Entry{key, value}
		return n
	}
	if n.order == MAX { // full root, need to split
		left, entry, right := n.split()
		n = &Node{
			order:  1,
			height: left.height + 1,
		}
		n.entries[0] = entry
		n.subtrees[0] = left
		n.subtrees[1] = right
	} else {
		n = n.dup()
	}
	n.insertNonFullMut(key, value)
	return n
}

func (n *Node) Remove(key Key) *Node {
	if _, ok := n.Get(key); !ok {
		return n
	}
	if n.height == 1 && n.order == 1 && n.entries[0].K.Cmp(key) == 0 {
		return nil
	}
	n = n.dup()
	n.removeStepMut(key)
	return n
}

func (n *Node) Min() *Entry {
	if n == nil {
		return nil
	}
	for finger := n; ; finger = finger.subtrees[0] {
		if finger.height == 1 {
			entry := finger.entries[0] // not taking address of an inner value directly
			return &entry
		}
	}
}

func (n *Node) Max() *Entry {
	if n == nil {
		return nil
	}
	for finger := n; ; finger = finger.subtrees[finger.order] {
		if finger.height == 1 {
			entry := finger.entries[finger.order-1] // not taking address of an inner value directly
			return &entry
		}
	}
}

func (n *Node) Height() int {
	if n == nil {
		return 0
	}
	return int(n.height)
}

func (n *Node) Iterate() Iterator {
	i := Iterator{
		stack: []iteratorStackFrame{{n, 0, 0}},
	}
	i.Next()
	return i
}

type iteratorStackFrame struct {
	n     *Node
	state uint8 // 0 start, 1 leftmost done, 2 i-th entry done, 3 done
	i     uint8
}
type Iterator struct {
	stack []iteratorStackFrame
	done  bool
	Entry Entry
}

func (i *Iterator) Done() bool {
	return i.done
}

func (i *Iterator) Next() {
	i.done = true
LOOP:
	for len(i.stack) > 0 {
		f := &i.stack[len(i.stack)-1]
		switch f.state {
		case 0:
			if f.n == nil {
				i.stack = i.stack[:len(i.stack)-1]
				continue LOOP
			}
			f.state = 1
			i.stack = append(i.stack, iteratorStackFrame{f.n.subtrees[0], 0, 0})
		case 1:
			if f.i < f.n.order {
				f.state = 2
				i.Entry = f.n.entries[f.i]
				i.done = false
				break LOOP
			} else {
				f.state = 3
			}
		case 2:
			f.i += 1
			f.state = 1
			i.stack = append(i.stack, iteratorStackFrame{f.n.subtrees[f.i], 0, 0})
		default:
			i.stack = i.stack[:len(i.stack)-1]
		}
	}
}

// TODO func (n *Node) Len() int {}
// TODO func (n *Node) IterateFrom(k Key) Iterator {}
// TODO func (n *Node) IterateReverse() Iterator {}
// TODO func (n *Node) IterateReverseFrom(k Key) Iterator {}

func (n *Node) removeStepMut(key Key) {
OUTER:
	for {
		if n.height == 1 {
			for i := 0; i < int(n.order); i++ {
				if n.entries[i].K.Cmp(key) == 0 {
					top := int(n.order) - 1
					for j := i; j < top; j++ {
						n.entries[j] = n.entries[j+1]
					}
					n.entries[n.order-1] = zeroEntry
					n.order -= 1
					return
				}
			}
			return
		} else {
			index := int(n.order)
			for i := 0; i < int(n.order); i++ {
				if n.entries[i].K.Cmp(key) == 0 { // inner delete
					index = n.ensureChildNotMinimal(i + 1)
					if n.order == 0 { // degenerated, need to drop a level
						*n = *n.subtrees[0]
						continue OUTER
					}
					if index != i+1 || n.entries[i].K.Cmp(key) != 0 { // merge OR rotation
						continue OUTER // easiest to try again
					}
					child := n.subtrees[index].dup()
					min := child.popMinMut()
					n.subtrees[index] = child
					n.entries[i] = min
					return
				}
				if key.Cmp(n.entries[i].K) < 0 {
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

func (n *Node) insertNonFullMut(key Key, value Value) {
OUTER:
	for {
		for i := 0; i < int(n.order); i++ {
			if n.entries[i].K.Cmp(key) == 0 {
				n.entries[i].V = value
				return
			}
		}
		if n.height == 1 {
			n.entries[n.order] = Entry{key, value}
			n.order += 1
			for i := int(n.order) - 1; i > 0; i-- {
				if n.entries[i].K.Cmp(n.entries[i-1].K) < 0 {
					n.entries[i], n.entries[i-1] = n.entries[i-1], n.entries[i]
				} else {
					break
				}
			}
			return
		}
		index := 0
		for i := 0; i < int(n.order); i++ {
			if key.Cmp(n.entries[i].K) > 0 {
				index = i + 1
			}
		}
		child := n.subtrees[index]
		if child.order == MAX { // full, need to split before entering
			left, entry, right := child.split()
			for i := int(n.order); i > index; i-- {
				n.entries[i] = n.entries[i-1]
			}
			n.entries[index] = entry
			for i := int(n.order); i > index; i-- {
				n.subtrees[i+1] = n.subtrees[i]
			}
			n.subtrees[index] = left
			n.subtrees[index+1] = right
			n.order += 1
			cmp := key.Cmp(entry.K)
			if cmp < 0 {
				n = left
				continue OUTER
			} else if cmp == 0 {
				n.entries[index].V = value
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
			ne := neighbour.entries[0]
			child.entries[1] = n.entries[index]
			n.entries[index] = ne
			child.subtrees[2] = neighbour.subtrees[0]
			copy(neighbour.entries[:], neighbour.entries[1:])
			copy(neighbour.subtrees[:], neighbour.subtrees[1:])
			child.order += 1
			neighbour.order -= 1
			neighbour.entries[neighbour.order] = zeroEntry
			n.subtrees[0] = child
			n.subtrees[1] = neighbour
		} else { // right neighbour is minimal
			child := n.subtrees[index]
			neighbour := n.subtrees[1]
			newChild := &Node{
				order:  child.order + neighbour.order + 1,
				height: child.height, // == neighbour.height
			}
			copy(newChild.entries[:], child.entries[:child.order])
			newChild.entries[child.order] = n.entries[index]
			copy(newChild.entries[child.order+1:], neighbour.entries[:neighbour.order])
			copy(newChild.subtrees[:], child.subtrees[:child.order+1])
			copy(newChild.subtrees[child.order+1:], neighbour.subtrees[:neighbour.order+1])
			n.subtrees[index] = newChild
			copy(n.subtrees[1:], n.subtrees[2:])
			copy(n.entries[0:], n.entries[1:])
			n.subtrees[n.order] = nil
			n.order -= 1
			n.entries[n.order] = zeroEntry
		}
	} else {
		child := n.subtrees[index]
		neighbour := n.subtrees[index-1]
		if neighbour.order > 1 {
			child = child.dup()
			neighbour = neighbour.dup()
			n.subtrees[index] = child
			n.subtrees[index-1] = neighbour
			copy(child.entries[1:], child.entries[:child.order])
			copy(child.subtrees[1:], child.subtrees[:child.order+1])
			child.order += 1
			child.entries[0] = n.entries[index-1]
			child.subtrees[0] = neighbour.subtrees[neighbour.order]
			n.entries[index-1] = neighbour.entries[neighbour.order-1]
			neighbour.subtrees[neighbour.order] = nil
			neighbour.order -= 1
			neighbour.entries[neighbour.order] = zeroEntry
		} else {
			newChild := &Node{
				order:  child.order + neighbour.order + 1,
				height: child.height, // == neighbour.height
			}
			copy(newChild.entries[:], neighbour.entries[:neighbour.order])
			newChild.entries[neighbour.order] = n.entries[index-1]
			copy(newChild.entries[neighbour.order+1:], child.entries[:child.order])
			copy(newChild.subtrees[:], neighbour.subtrees[:neighbour.order+1])
			copy(newChild.subtrees[neighbour.order+1:], child.subtrees[:child.order+1])
			copy(n.subtrees[index-1:], n.subtrees[index:])
			n.subtrees[n.order] = nil
			n.subtrees[index-1] = newChild
			copy(n.entries[index-1:], n.entries[index:])
			n.order -= 1
			n.entries[n.order] = zeroEntry
			index -= 1
		}
	}
	return index
}

func (n *Node) split() (left *Node, entry Entry, right *Node) {
	entry = n.entries[(MAX-1)/2]
	left = &Node{
		order:  (MAX - 1) / 2,
		height: n.height,
	}
	for i := 0; i < (MAX-1)/2; i++ {
		left.entries[i] = n.entries[i]
	}
	for i := 0; i <= (MAX-1)/2; i++ {
		left.subtrees[i] = n.subtrees[i]
	}
	right = &Node{
		order:  (MAX - 1) / 2,
		height: n.height,
	}
	for i := (MAX + 1) / 2; i < MAX; i++ {
		right.entries[i-(MAX+1)/2] = n.entries[i]
	}
	for i := (MAX + 1) / 2; i <= MAX; i++ {
		right.subtrees[i-(MAX+1)/2] = n.subtrees[i]
	}
	return
}

func (n *Node) popMinMut() Entry {
OUTER:
	for {
		if n.height == 1 {
			e := n.entries[0]
			for i := 1; i < int(n.order); i++ {
				n.entries[i-1] = n.entries[i]
			}
			n.order -= 1
			n.entries[n.order] = zeroEntry
			return e
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
		s += fmt.Sprintf(" %v %v", n.entries[i], n.subtrees[i+1].visual())
	}
	s += " ]"
	return s
}
