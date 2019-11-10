package main

import "fmt"

type IntStrMapEntry struct {
	intKey   intKey
	string string
}

type IntStrMap struct {
	IntStrMapEntry    IntStrMapEntry
	h        int
	children [2]*IntStrMap
}

func (n *IntStrMap) Height() int {
	if n == nil {
		return 0
	}
	return n.h
}

func combinedDepth(n1, n2 *IntStrMap) int {
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

func mkNode(entry IntStrMapEntry, left *IntStrMap, right *IntStrMap) *IntStrMap {
	return &IntStrMap{
		IntStrMapEntry:    entry,
		h:        combinedDepth(left, right),
		children: [2]*IntStrMap{left, right},
	}
}

func (node *IntStrMap) Get(key intKey) (value string, ok bool) {
	if node == nil {
		ok = false
		return // using named returns so we keep the zero value for `value`
	}
	if key.Less(node.IntStrMapEntry.intKey) {
		return node.children[0].Get(key)
	}
	if node.IntStrMapEntry.intKey.Less(key) {
		return node.children[1].Get(key)
	}
	// equal
	return node.IntStrMapEntry.string, true
}

func (node *IntStrMap) Insert(key intKey, value string) *IntStrMap {
	if node == nil {
		return mkNode(IntStrMapEntry{key, value}, nil, nil)
	}
	entry, left, right := node.IntStrMapEntry, node.children[0], node.children[1]
	if node.IntStrMapEntry.intKey.Less(key) {
		right = right.Insert(key, value)
	} else if key.Less(node.IntStrMapEntry.intKey) {
		left = left.Insert(key, value)
	} else { // equals
		entry = IntStrMapEntry{key, value}
	}
	return rotate(entry, left, right)
}

func (node *IntStrMap) Remove(key intKey) *IntStrMap {
	if node == nil {
		return nil
	}
	entry, left, right := node.IntStrMapEntry, node.children[0], node.children[1]
	if node.IntStrMapEntry.intKey.Less(key) {
		right = right.Remove(key)
	} else if key.Less(node.IntStrMapEntry.intKey) {
		left = left.Remove(key)
	} else { // equals
		max := left.Max()
		if max == nil {
			return right
		} else {
			left = left.Remove(max.intKey)
			entry = *max
		}
	}
	return rotate(entry, left, right)
}

func rotate(entry IntStrMapEntry, left *IntStrMap, right *IntStrMap) *IntStrMap {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepth(left, rl)-rr.Height() > 1 {
			// double rotation
			return mkNode(
				rl.IntStrMapEntry,
				mkNode(entry, left, rl.children[0]),
				mkNode(right.IntStrMapEntry, rl.children[1], rr),
			)
		}
		return mkNode(right.IntStrMapEntry, mkNode(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepth(right, lr)-ll.Height() > 1 {
			// double rotation
			return mkNode(
				lr.IntStrMapEntry,
				mkNode(left.IntStrMapEntry, ll, lr.children[0]),
				mkNode(entry, lr.children[1], right),
			)
		}
		return mkNode(left.IntStrMapEntry, ll, mkNode(entry, lr, right))
	}
	return mkNode(entry, left, right)
}

func (node *IntStrMap) Len() int {
	if node == nil {
		return 0
	}
	return 1 + node.children[0].Len() + node.children[1].Len()
}

func (node *IntStrMap) Entries() []IntStrMapEntry {
	elems := make([]IntStrMapEntry, 0)
	var step func(n *IntStrMap)
	step = func(n *IntStrMap) {
		if n == nil {
			return
		}
		step(n.children[0])
		elems = append(elems, n.IntStrMapEntry)
		step(n.children[1])
	}
	step(node)
	return elems
}

func (node *IntStrMap) extreme(dir int) *IntStrMapEntry {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.IntStrMapEntry
}

func (node *IntStrMap) Min() *IntStrMapEntry {
	return node.extreme(0)
}

func (node *IntStrMap) Max() *IntStrMapEntry {
	return node.extreme(1)
}

func (node *IntStrMap) Iterate() IntStrMapIterator {
	return newIterator(node, 0)
}

func (node *IntStrMap) IterateReverse() IntStrMapIterator {
	return newIterator(node, 1)
}

type iteratorStackFrame struct {
	node  *IntStrMap
	state int8
}

type IntStrMapIterator struct {
	direction    int
	stack        []iteratorStackFrame
	currentEntry IntStrMapEntry
}

func newIterator(node *IntStrMap, direction int) IntStrMapIterator {
	stack := make([]iteratorStackFrame, 1, node.Height())
	stack[0] = iteratorStackFrame{node: node, state: 0}
	iter := IntStrMapIterator{direction: direction, stack: stack}
	iter.Next()
	return iter
}

func (i *IntStrMapIterator) Done() bool {
	return len(i.stack) == 0
}

func (i *IntStrMapIterator) GetKey() intKey {
	return i.currentEntry.intKey
}

func (i *IntStrMapIterator) GetValue() string {
	return i.currentEntry.string
}

func (i *IntStrMapIterator) Next() {
	for len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		switch frame.state {
		case 0:
			if frame.node == nil {
				last := len(i.stack) - 1
				i.stack[last] = iteratorStackFrame{} // zero out
				i.stack = i.stack[:last]             // pop
			} else {
				frame.state = 1
			}
		case 1:
			i.stack = append(i.stack, iteratorStackFrame{node: frame.node.children[i.direction], state: 0})
			frame.state = 2
		case 2:
			i.currentEntry = frame.node.IntStrMapEntry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = iteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		default:
			panic(fmt.Sprintf("Unknown state %v", frame.state))
		}

	}
}
