package main

import "fmt"

type Key interface {
	Less(Key) bool
}

type Value interface{}

type Entry struct {
	Key   Key
	Value Value
}

type Node struct {
	Entry    Entry
	h        int
	children [2]*Node
}

func (n *Node) Height() int {
	if n == nil {
		return 0
	}
	return n.h
}

func combinedDepth(n1, n2 *Node) int {
	d1 := n1.Height()
	d2 := n2.Height()
	var d int
	if d1 > d2 {
		d = d1
	} else {
		d = d2
	}
	return d + 1
}

func mkNode(entry Entry, left *Node, right *Node) *Node {
	return &Node{
		Entry:    entry,
		h:        combinedDepth(left, right),
		children: [2]*Node{left, right},
	}
}

func (node *Node) Get(key Key) (value Value, ok bool) {
	if node == nil {
		return nil, false
	}
	if key.Less(node.Entry.Key) {
		return node.children[0].Get(key)
	}
	if node.Entry.Key.Less(key) {
		return node.children[1].Get(key)
	}
	// equal
	return node.Entry.Value, true
}

func (node *Node) Insert(key Key, value Value) *Node {
	if node == nil {
		return mkNode(Entry{key, value}, nil, nil)
	}
	entry, left, right := node.Entry, node.children[0], node.children[1]
	if node.Entry.Key.Less(key) {
		right = right.Insert(key, value)
	} else if key.Less(node.Entry.Key) {
		left = left.Insert(key, value)
	} else { // equals
		entry = Entry{key, value}
	}
	return rotate(entry, left, right)
}

func (node *Node) Remove(key Key) *Node {
	if node == nil {
		return nil
	}
	entry, left, right := node.Entry, node.children[0], node.children[1]
	if node.Entry.Key.Less(key) {
		right = right.Remove(key)
	} else if key.Less(node.Entry.Key) {
		left = left.Remove(key)
	} else { // equals
		max := left.Max()
		if max == nil {
			return right
		} else {
			left = left.Remove(max.Key)
			entry = *max
		}
	}
	return rotate(entry, left, right)
}

func rotate(entry Entry, left *Node, right *Node) *Node {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepth(left, rl)-rr.Height() > 1 {
			// double rotation
			return mkNode(
				rl.Entry,
				mkNode(entry, left, rl.children[0]),
				mkNode(right.Entry, rl.children[1], rr),
			)
		}
		return mkNode(right.Entry, mkNode(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepth(right, lr)-ll.Height() > 1 {
			// double rotation
			return mkNode(
				lr.Entry,
				mkNode(left.Entry, ll, lr.children[0]),
				mkNode(entry, lr.children[1], right),
			)
		}
		return mkNode(left.Entry, ll, mkNode(entry, lr, right))
	}
	return mkNode(entry, left, right)
}

func (node *Node) Len() int {
	if node == nil {
		return 0
	}
	return 1 + node.children[0].Len() + node.children[1].Len()
}

func (node *Node) Entries() []Entry {
	elems := make([]Entry, 0)
	var step func(n *Node)
	step = func(n *Node) {
		if n == nil {
			return
		}
		step(n.children[0])
		elems = append(elems, n.Entry)
		step(n.children[1])
	}
	step(node)
	return elems
}

func (node *Node) extreme(dir int) *Entry {
	if node == nil {
		return nil
	}
	child := node.children[dir]
	if child != nil {
		return child.extreme(dir)
	}
	return &node.Entry
}

func (node *Node) Min() *Entry {
	return node.extreme(0)
}

func (node *Node) Max() *Entry {
	return node.extreme(1)
}

func (node *Node) Iterator() *Iterator {
	return newIterator(node, 0)
}

func (node *Node) IteratorReverse() *Iterator {
	return newIterator(node, 1)
}

type iteratorStackFrame struct {
	node  *Node
	state int8
}

type Iterator struct {
	direction    int
	stack        []*iteratorStackFrame
	currentEntry Entry
}

func newIterator(node *Node, direction int) *Iterator {
	iter := &Iterator{
		direction: direction,
		stack:     []*iteratorStackFrame{{node: node, state: 0}},
	}
	iter.Next()
	return iter
}

func (i *Iterator) Done() bool {
	return len(i.stack) == 0
}

func (i *Iterator) Key() Key {
	return i.currentEntry.Key
}

func (i *Iterator) Value() Value {
	return i.currentEntry.Value
}

func (i *Iterator) Next() {
	for len(i.stack) > 0 {
		frame := i.stack[len(i.stack)-1]
		switch frame.state {
		case 0:
			if frame.node == nil {
				i.stack = i.stack[:len(i.stack)-1] // pop
			} else {
				frame.state = 1
			}
		case 1:
			i.stack = append(i.stack, &iteratorStackFrame{node: frame.node.children[i.direction], state: 0})
			frame.state = 2
		case 2:
			i.currentEntry = frame.node.Entry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = &iteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		default:
			panic(fmt.Sprintf("Unknown state %v", frame.state))
		}

	}
}
