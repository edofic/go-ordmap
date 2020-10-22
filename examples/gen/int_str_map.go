// DO NOT EDIT tis code was generated using go-ordmap code generation
// go run github.com/edofic/go-ordmap/cmd/gen -name IntStrMap -key intKey -value string -target ./int_str_map.go
package main

type IntStrMapEntry struct {
	K intKey
	V string
}

type IntStrMap struct {
	IntStrMapEntry IntStrMapEntry
	h              int
	len            int
	children       [2]*IntStrMap
}

func (node *IntStrMap) Height() int {
	if node == nil {
		return 0
	}
	return node.h
}

// suffix IntStrMap is needed because this will get specialised in codegen
func combinedDepthIntStrMap(n1, n2 *IntStrMap) int {
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

// suffix IntStrMap is needed because this will get specialised in codegen
func mkIntStrMap(entry IntStrMapEntry, left *IntStrMap, right *IntStrMap) *IntStrMap {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &IntStrMap{
		IntStrMapEntry: entry,
		h:              combinedDepthIntStrMap(left, right),
		len:            len,
		children:       [2]*IntStrMap{left, right},
	}
}

func (node *IntStrMap) Get(key intKey) (value string, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if key.Less(finger.IntStrMapEntry.K) {
			finger = finger.children[0]
		} else if finger.IntStrMapEntry.K.Less(key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.IntStrMapEntry.V, true
		}
	}
}

func (node *IntStrMap) Insert(key intKey, value string) *IntStrMap {
	if node == nil {
		return mkIntStrMap(IntStrMapEntry{key, value}, nil, nil)
	}
	entry, left, right := node.IntStrMapEntry, node.children[0], node.children[1]
	if node.IntStrMapEntry.K.Less(key) {
		right = right.Insert(key, value)
	} else if key.Less(node.IntStrMapEntry.K) {
		left = left.Insert(key, value)
	} else { // equals
		entry = IntStrMapEntry{key, value}
	}
	return rotateIntStrMap(entry, left, right)
}

func (node *IntStrMap) Remove(key intKey) *IntStrMap {
	if node == nil {
		return nil
	}
	entry, left, right := node.IntStrMapEntry, node.children[0], node.children[1]
	if node.IntStrMapEntry.K.Less(key) {
		right = right.Remove(key)
	} else if key.Less(node.IntStrMapEntry.K) {
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
	return rotateIntStrMap(entry, left, right)
}

// suffix IntStrMap is needed because this will get specialised in codegen
func rotateIntStrMap(entry IntStrMapEntry, left *IntStrMap, right *IntStrMap) *IntStrMap {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepthIntStrMap(left, rl)-rr.Height() > 1 {
			// double rotation
			return mkIntStrMap(
				rl.IntStrMapEntry,
				mkIntStrMap(entry, left, rl.children[0]),
				mkIntStrMap(right.IntStrMapEntry, rl.children[1], rr),
			)
		}
		return mkIntStrMap(right.IntStrMapEntry, mkIntStrMap(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepthIntStrMap(right, lr)-ll.Height() > 1 {
			// double rotation
			return mkIntStrMap(
				lr.IntStrMapEntry,
				mkIntStrMap(left.IntStrMapEntry, ll, lr.children[0]),
				mkIntStrMap(entry, lr.children[1], right),
			)
		}
		return mkIntStrMap(left.IntStrMapEntry, ll, mkIntStrMap(entry, lr, right))
	}
	return mkIntStrMap(entry, left, right)
}

func (node *IntStrMap) Len() int {
	if node == nil {
		return 0
	}
	return node.len
}

func (node *IntStrMap) Entries() []IntStrMapEntry {
	elems := make([]IntStrMapEntry, 0, node.Len())
	if node == nil {
		return elems
	}
	type frame struct {
		node     *IntStrMap
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
			elems = append(elems, top.node.IntStrMapEntry)
			if top.node.children[1] != nil {
				stack = append(stack, frame{top.node.children[1], false})
			}
		}
	}
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
	return newIteratorIntStrMap(node, 0, nil)
}

func (node *IntStrMap) IterateFrom(k intKey) IntStrMapIterator {
	return newIteratorIntStrMap(node, 0, &k)
}

func (node *IntStrMap) IterateReverse() IntStrMapIterator {
	return newIteratorIntStrMap(node, 1, nil)
}

func (node *IntStrMap) IterateReverseFrom(k intKey) IntStrMapIterator {
	return newIteratorIntStrMap(node, 1, &k)
}

type IntStrMapIteratorStackFrame struct {
	node  *IntStrMap
	state int8
}

type IntStrMapIterator struct {
	direction    int
	stack        []IntStrMapIteratorStackFrame
	currentEntry IntStrMapEntry
}

// suffix IntStrMap is needed because this will get specialised in codegen
func newIteratorIntStrMap(node *IntStrMap, direction int, startFrom *intKey) IntStrMapIterator {
	if node == nil {
		return IntStrMapIterator{}
	}
	stack := make([]IntStrMapIteratorStackFrame, 1, node.Height())
	stack[0] = IntStrMapIteratorStackFrame{node: node, state: 0}
	iter := IntStrMapIterator{direction: direction, stack: stack}
	if startFrom != nil {
		stack[0].state = 2
		iter.seek(*startFrom)
	} else {
		iter.Next()
	}
	return iter
}

func (i *IntStrMapIterator) Done() bool {
	return len(i.stack) == 0
}

func (i *IntStrMapIterator) GetKey() intKey {
	return i.currentEntry.K
}

func (i *IntStrMapIterator) GetValue() string {
	return i.currentEntry.V
}

func (i *IntStrMapIterator) Next() {
	for len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		switch frame.state {
		case 0:
			if frame.node == nil {
				last := len(i.stack) - 1
				i.stack[last] = IntStrMapIteratorStackFrame{} // zero out
				i.stack = i.stack[:last]                      // pop
			} else {
				frame.state = 1
			}
		case 1:
			i.stack = append(i.stack, IntStrMapIteratorStackFrame{node: frame.node.children[i.direction], state: 0})
			frame.state = 2
		case 2:
			i.currentEntry = frame.node.IntStrMapEntry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = IntStrMapIteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		}

	}
}

func (i *IntStrMapIterator) seek(k intKey) {
LOOP:
	for {
		frame := &i.stack[len(i.stack)-1]
		if frame.node == nil {
			last := len(i.stack) - 1
			i.stack[last] = IntStrMapIteratorStackFrame{} // zero out
			i.stack = i.stack[:last]                      // pop
			break LOOP
		}
		if (i.direction == 0 && !(frame.node.IntStrMapEntry.K.Less(k))) || (i.direction == 1 && !(k.Less(frame.node.IntStrMapEntry.K))) {
			i.stack = append(i.stack, IntStrMapIteratorStackFrame{node: frame.node.children[i.direction], state: 2})
		} else {
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = IntStrMapIteratorStackFrame{node: frame.node.children[1-i.direction], state: 2}
		}
	}
	if len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		i.currentEntry = frame.node.IntStrMapEntry
		frame.state = 3
	}
}
