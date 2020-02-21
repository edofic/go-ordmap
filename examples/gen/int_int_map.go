// DO NOT EDIT tis code was generated using go-ordmap code generation
// go run github.com/edofic/go-ordmap/cmd/gen -name IntIntMap -key int -less < -value int -target ./int_int_map.go
package main

type IntIntMapEntry struct {
	K int
	V int
}

type IntIntMap struct {
	IntIntMapEntry IntIntMapEntry
	h              int
	len            int
	children       [2]*IntIntMap
}

func (n *IntIntMap) Height() int {
	if n == nil {
		return 0
	}
	return n.h
}

// suffix IntIntMap is needed because this will get specialised in codegen
func combinedDepthIntIntMap(n1, n2 *IntIntMap) int {
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

// suffix IntIntMap is needed because this will get specialised in codegen
func mkIntIntMap(entry IntIntMapEntry, left *IntIntMap, right *IntIntMap) *IntIntMap {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &IntIntMap{
		IntIntMapEntry: entry,
		h:              combinedDepthIntIntMap(left, right),
		len:            len,
		children:       [2]*IntIntMap{left, right},
	}
}

func (node *IntIntMap) Get(key int) (value int, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if key < (finger.IntIntMapEntry.K) {
			finger = finger.children[0]
		} else if finger.IntIntMapEntry.K < (key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.IntIntMapEntry.V, true
		}
	}
}

func (node *IntIntMap) Insert(key int, value int) *IntIntMap {
	if node == nil {
		return mkIntIntMap(IntIntMapEntry{key, value}, nil, nil)
	}
	entry, left, right := node.IntIntMapEntry, node.children[0], node.children[1]
	if node.IntIntMapEntry.K < (key) {
		right = right.Insert(key, value)
	} else if key < (node.IntIntMapEntry.K) {
		left = left.Insert(key, value)
	} else { // equals
		entry = IntIntMapEntry{key, value}
	}
	return rotateIntIntMap(entry, left, right)
}

func (node *IntIntMap) Remove(key int) *IntIntMap {
	if node == nil {
		return nil
	}
	entry, left, right := node.IntIntMapEntry, node.children[0], node.children[1]
	if node.IntIntMapEntry.K < (key) {
		right = right.Remove(key)
	} else if key < (node.IntIntMapEntry.K) {
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
	return rotateIntIntMap(entry, left, right)
}

// suffix IntIntMap is needed because this will get specialised in codegen
func rotateIntIntMap(entry IntIntMapEntry, left *IntIntMap, right *IntIntMap) *IntIntMap {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepthIntIntMap(left, rl)-rr.Height() > 1 {
			// double rotation
			return mkIntIntMap(
				rl.IntIntMapEntry,
				mkIntIntMap(entry, left, rl.children[0]),
				mkIntIntMap(right.IntIntMapEntry, rl.children[1], rr),
			)
		}
		return mkIntIntMap(right.IntIntMapEntry, mkIntIntMap(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepthIntIntMap(right, lr)-ll.Height() > 1 {
			// double rotation
			return mkIntIntMap(
				lr.IntIntMapEntry,
				mkIntIntMap(left.IntIntMapEntry, ll, lr.children[0]),
				mkIntIntMap(entry, lr.children[1], right),
			)
		}
		return mkIntIntMap(left.IntIntMapEntry, ll, mkIntIntMap(entry, lr, right))
	}
	return mkIntIntMap(entry, left, right)
}

func (node *IntIntMap) Len() int {
	if node == nil {
		return 0
	}
	return node.len
}

func (node *IntIntMap) Entries() []IntIntMapEntry {
	elems := make([]IntIntMapEntry, 0)
	var step func(n *IntIntMap)
	step = func(n *IntIntMap) {
		if n == nil {
			return
		}
		step(n.children[0])
		elems = append(elems, n.IntIntMapEntry)
		step(n.children[1])
	}
	step(node)
	return elems
}

func (node *IntIntMap) extreme(dir int) *IntIntMapEntry {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.IntIntMapEntry
}

func (node *IntIntMap) Min() *IntIntMapEntry {
	return node.extreme(0)
}

func (node *IntIntMap) Max() *IntIntMapEntry {
	return node.extreme(1)
}

func (node *IntIntMap) Iterate() IntIntMapIterator {
	return newIteratorIntIntMap(node, 0, nil)
}

func (node *IntIntMap) IterateFrom(k int) IntIntMapIterator {
	return newIteratorIntIntMap(node, 0, &k)
}

func (node *IntIntMap) IterateReverse() IntIntMapIterator {
	return newIteratorIntIntMap(node, 1, nil)
}

func (node *IntIntMap) IterateReverseFrom(k int) IntIntMapIterator {
	return newIteratorIntIntMap(node, 1, &k)
}

type IntIntMapIteratorStackFrame struct {
	node  *IntIntMap
	state int8
}

type IntIntMapIterator struct {
	direction    int
	stack        []IntIntMapIteratorStackFrame
	currentEntry IntIntMapEntry
}

// suffix IntIntMap is needed because this will get specialised in codegen
func newIteratorIntIntMap(node *IntIntMap, direction int, startFrom *int) IntIntMapIterator {
	if node == nil {
		return IntIntMapIterator{}
	}
	stack := make([]IntIntMapIteratorStackFrame, 1, node.Height())
	stack[0] = IntIntMapIteratorStackFrame{node: node, state: 0}
	iter := IntIntMapIterator{direction: direction, stack: stack}
	if startFrom != nil {
		stack[0].state = 2
		iter.seek(*startFrom)
	} else {
		iter.Next()
	}
	return iter
}

func (i *IntIntMapIterator) Done() bool {
	return len(i.stack) == 0
}

func (i *IntIntMapIterator) GetKey() int {
	return i.currentEntry.K
}

func (i *IntIntMapIterator) GetValue() int {
	return i.currentEntry.V
}

func (i *IntIntMapIterator) Next() {
	for len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		switch frame.state {
		case 0:
			if frame.node == nil {
				last := len(i.stack) - 1
				i.stack[last] = IntIntMapIteratorStackFrame{} // zero out
				i.stack = i.stack[:last]                      // pop
			} else {
				frame.state = 1
			}
		case 1:
			i.stack = append(i.stack, IntIntMapIteratorStackFrame{node: frame.node.children[i.direction], state: 0})
			frame.state = 2
		case 2:
			i.currentEntry = frame.node.IntIntMapEntry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = IntIntMapIteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		}

	}
}

func (i *IntIntMapIterator) seek(k int) {
LOOP:
	for {
		frame := &i.stack[len(i.stack)-1]
		if frame.node == nil {
			last := len(i.stack) - 1
			i.stack[last] = IntIntMapIteratorStackFrame{} // zero out
			i.stack = i.stack[:last]                      // pop
			break LOOP
		}
		if (i.direction == 0 && !(frame.node.IntIntMapEntry.K < (k))) || (i.direction == 1 && !(k < (frame.node.IntIntMapEntry.K))) {
			i.stack = append(i.stack, IntIntMapIteratorStackFrame{node: frame.node.children[i.direction], state: 2})
		} else {
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = IntIntMapIteratorStackFrame{node: frame.node.children[1-i.direction], state: 2}
		}
	}
	if len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		i.currentEntry = frame.node.IntIntMapEntry
		frame.state = 3
	}
}
