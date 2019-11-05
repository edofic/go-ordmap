package main

import "fmt"

type Key interface {
	Less(Key) bool
	Eq(Key) bool
}

type Value interface{}

type Entry struct {
	Key   Key
	Value Value
}

type Node struct {
	Entry    Entry
	balance  int8
	children [2]*Node
}

func (n Node) dup() *Node {
	return &n
}

func (n *Node) depth() int {
	if n == nil {
		return 0
	}
	depthLeft := n.children[0].depth()
	depthRight := n.children[1].depth()
	var depth int
	if depthLeft > depthRight {
		depth = depthLeft
	} else {
		depth = depthRight
	}
	return depth + 1
}

func (node Node) singleRot(dir int) *Node {
	tmp := node.children[1-dir].dup()
	node.children[1-dir] = tmp.children[dir]
	tmp.children[dir] = &node
	return tmp
}

func (node Node) doubleRot(dir int) *Node {
	tmp := node.children[1-dir].children[dir].dup()

	node.children[1-dir] = node.children[1-dir].dup()
	node.children[1-dir].children[dir] = tmp.children[1-dir]
	tmp.children[1-dir] = node.children[1-dir]
	node.children[1-dir] = tmp

	tmp = node.children[1-dir].dup()
	node.children[1-dir] = tmp.children[dir]
	tmp.children[dir] = &node
	return tmp
}

func (node *Node) adjustBalance(dir int, bal int8) {
	n := node.children[dir].dup()
	node.children[dir] = n
	nn := n.children[1-dir].dup()
	n.children[1-dir] = nn
	switch nn.balance {
	case 0:
		node.balance = 0
		n.balance = 0
	case bal:
		node.balance = -bal
		n.balance = 0
	default:
		node.balance = 0
		n.balance = bal
	}
	nn.balance = 0
}

func (node *Node) Get(key Key) (value Value, ok bool) {
	if node == nil {
		return nil, false
	}
	if node.Entry.Key.Eq(key) {
		return node.Entry.Value, true
	}
	if key.Less(node.Entry.Key) {
		return node.children[0].Get(key)
	}
	return node.children[1].Get(key)
}

func (node *Node) Insert(key Key, value Value) *Node {
	entry := Entry{key, value}
	var step func(*Node) (*Node, bool)
	step = func(node *Node) (*Node, bool) {
		if node == nil {
			return &Node{Entry: entry}, false
		}
		node = node.dup()
		if node.Entry.Key.Eq(key) {
			node.Entry = Entry{key, value}
			return node, true
		}
		dir := 0
		if node.Entry.Key.Less(key) {
			dir = 1
		}
		var done bool
		node.children[dir], done = step(node.children[dir])
		if done {
			return node, true
		}
		node.balance += int8(2*dir - 1)
		switch node.balance {
		case 0:
			return node, true
		case 1, -1:
			return node, false
		}
		n := node.children[dir].dup()
		node.children[dir] = n
		bal := 2*dir - 1
		if n.balance == int8(bal) {
			node.balance = 0
			n.balance = 0
			return node.singleRot(1 - dir), true
		}
		node.adjustBalance(dir, int8(bal))
		return node.doubleRot(1 - dir), true
	}
	tree, _ := step(node)
	return tree
}

func (node *Node) Remove(key Key) *Node {
	var step func(*Node) (*Node, bool)
	step = func(node *Node) (*Node, bool) {
		if node == nil {
			return nil, false
		}
		node = node.dup()
		if node.Entry.Key.Eq(key) {
			switch {
			case node.children[0] == nil:
				return node.children[1], false
			case node.children[1] == nil:
				return node.children[0], false
			}
			heir := node.children[0]
			for heir.children[1] != nil {
				heir = heir.children[1]
			}
			node.Entry = heir.Entry
			key = heir.Entry.Key
		}
		dir := 0
		if node.Entry.Key.Less(key) {
			dir = 1
		}
		var done bool
		node.children[dir], done = step(node.children[dir])
		if done {
			return node, true
		}
		node.balance += int8(1 - 2*dir)
		switch node.balance {
		case 1, -1:
			return node, true
		case 0:
			return node, false
		}
		n := node.children[1-dir].dup()
		node.children[1-dir] = n
		bal := 2*dir - 1
		switch int(n.balance) {
		case -bal:
			node.balance = 0
			n.balance = 0
			return node.singleRot(dir), false
		case bal:
			node.adjustBalance(1-dir, int8(-bal))
			return node.doubleRot(dir), false
		}
		node.balance = int8(-bal)
		n.balance = int8(bal)
		return node.singleRot(dir), true
	}
	tree, _ := step(node)
	return tree
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
LOOP:
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
			break LOOP
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = &iteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		default:
			panic(fmt.Sprintf("Unknown state %v", frame.state))
		}

	}
}
