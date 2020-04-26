// DO NOT EDIT tis code was generated using go-ordmap code generation
// go run github.com/edofic/go-ordmap/cmd/gen -name Index -key CompositeKey -value bool -target ./composite.go
package main

type IndexEntry struct {
	K CompositeKey
	V bool
}

type Index struct {
	IndexEntry IndexEntry
	h          int
	len        int
	children   [2]*Index
}

func (n *Index) Height() int {
	if n == nil {
		return 0
	}
	return n.h
}

// suffix Index is needed because this will get specialised in codegen
func combinedDepthIndex(n1, n2 *Index) int {
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

// suffix Index is needed because this will get specialised in codegen
func mkIndex(entry IndexEntry, left *Index, right *Index) *Index {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &Index{
		IndexEntry: entry,
		h:          combinedDepthIndex(left, right),
		len:        len,
		children:   [2]*Index{left, right},
	}
}

func (node *Index) Get(key CompositeKey) (value bool, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if key.Less(finger.IndexEntry.K) {
			finger = finger.children[0]
		} else if finger.IndexEntry.K.Less(key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.IndexEntry.V, true
		}
	}
}

func (node *Index) Insert(key CompositeKey, value bool) *Index {
	if node == nil {
		return mkIndex(IndexEntry{key, value}, nil, nil)
	}
	entry, left, right := node.IndexEntry, node.children[0], node.children[1]
	if node.IndexEntry.K.Less(key) {
		right = right.Insert(key, value)
	} else if key.Less(node.IndexEntry.K) {
		left = left.Insert(key, value)
	} else { // equals
		entry = IndexEntry{key, value}
	}
	return rotateIndex(entry, left, right)
}

func (node *Index) Remove(key CompositeKey) *Index {
	if node == nil {
		return nil
	}
	entry, left, right := node.IndexEntry, node.children[0], node.children[1]
	if node.IndexEntry.K.Less(key) {
		right = right.Remove(key)
	} else if key.Less(node.IndexEntry.K) {
		left = left.Remove(key)
	} else { // equals
		max := left.Max()
		if max == nil {
			return right
		} else {
			left = left.Remove(max.K)
			entry = *max
		}
	}
	return rotateIndex(entry, left, right)
}

// suffix Index is needed because this will get specialised in codegen
func rotateIndex(entry IndexEntry, left *Index, right *Index) *Index {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepthIndex(left, rl)-rr.Height() > 1 {
			// double rotation
			return mkIndex(
				rl.IndexEntry,
				mkIndex(entry, left, rl.children[0]),
				mkIndex(right.IndexEntry, rl.children[1], rr),
			)
		}
		return mkIndex(right.IndexEntry, mkIndex(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepthIndex(right, lr)-ll.Height() > 1 {
			// double rotation
			return mkIndex(
				lr.IndexEntry,
				mkIndex(left.IndexEntry, ll, lr.children[0]),
				mkIndex(entry, lr.children[1], right),
			)
		}
		return mkIndex(left.IndexEntry, ll, mkIndex(entry, lr, right))
	}
	return mkIndex(entry, left, right)
}

func (node *Index) Len() int {
	if node == nil {
		return 0
	}
	return node.len
}

func (node *Index) Entries() []IndexEntry {
	elems := make([]IndexEntry, 0, node.Len())
	if node == nil {
		return elems
	}
	type frame struct {
		node     *Index
		leftDone bool
	}
	var preallocated [20]frame // preallocate on stack for common case
	stack := preallocated[:0]
	stack = append(stack, frame{node, false})
	for len(stack) > 0 {
		top := &stack[len(stack)-1]

		if !top.leftDone {
			if top.node.children[0] != nil {
				stack = append(stack, frame{top.node.children[0], false})
			}
			top.leftDone = true
		} else {
			stack = stack[:len(stack)-1] // pop
			elems = append(elems, top.node.IndexEntry)
			if top.node.children[1] != nil {
				stack = append(stack, frame{top.node.children[1], false})
			}
		}
	}
	return elems
}

func (node *Index) extreme(dir int) *IndexEntry {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.IndexEntry
}

func (node *Index) Min() *IndexEntry {
	return node.extreme(0)
}

func (node *Index) Max() *IndexEntry {
	return node.extreme(1)
}

func (node *Index) Iterate() IndexIterator {
	return newIteratorIndex(node, 0, nil)
}

func (node *Index) IterateFrom(k CompositeKey) IndexIterator {
	return newIteratorIndex(node, 0, &k)
}

func (node *Index) IterateReverse() IndexIterator {
	return newIteratorIndex(node, 1, nil)
}

func (node *Index) IterateReverseFrom(k CompositeKey) IndexIterator {
	return newIteratorIndex(node, 1, &k)
}

type IndexIteratorStackFrame struct {
	node  *Index
	state int8
}

type IndexIterator struct {
	direction    int
	stack        []IndexIteratorStackFrame
	currentEntry IndexEntry
}

// suffix Index is needed because this will get specialised in codegen
func newIteratorIndex(node *Index, direction int, startFrom *CompositeKey) IndexIterator {
	if node == nil {
		return IndexIterator{}
	}
	stack := make([]IndexIteratorStackFrame, 1, node.Height())
	stack[0] = IndexIteratorStackFrame{node: node, state: 0}
	iter := IndexIterator{direction: direction, stack: stack}
	if startFrom != nil {
		stack[0].state = 2
		iter.seek(*startFrom)
	} else {
		iter.Next()
	}
	return iter
}

func (i *IndexIterator) Done() bool {
	return len(i.stack) == 0
}

func (i *IndexIterator) GetKey() CompositeKey {
	return i.currentEntry.K
}

func (i *IndexIterator) GetValue() bool {
	return i.currentEntry.V
}

func (i *IndexIterator) Next() {
	for len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		switch frame.state {
		case 0:
			if frame.node == nil {
				last := len(i.stack) - 1
				i.stack[last] = IndexIteratorStackFrame{} // zero out
				i.stack = i.stack[:last]                  // pop
			} else {
				frame.state = 1
			}
		case 1:
			i.stack = append(i.stack, IndexIteratorStackFrame{node: frame.node.children[i.direction], state: 0})
			frame.state = 2
		case 2:
			i.currentEntry = frame.node.IndexEntry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = IndexIteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		}

	}
}

func (i *IndexIterator) seek(k CompositeKey) {
LOOP:
	for {
		frame := &i.stack[len(i.stack)-1]
		if frame.node == nil {
			last := len(i.stack) - 1
			i.stack[last] = IndexIteratorStackFrame{} // zero out
			i.stack = i.stack[:last]                  // pop
			break LOOP
		}
		if (i.direction == 0 && !(frame.node.IndexEntry.K.Less(k))) || (i.direction == 1 && !(k.Less(frame.node.IndexEntry.K))) {
			i.stack = append(i.stack, IndexIteratorStackFrame{node: frame.node.children[i.direction], state: 2})
		} else {
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = IndexIteratorStackFrame{node: frame.node.children[1-i.direction], state: 2}
		}
	}
	if len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		i.currentEntry = frame.node.IndexEntry
		frame.state = 3
	}
}
