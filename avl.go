package ordmap

type Entry struct {
	K Key
	V Value
}

type OrdMap struct {
	Entry    Entry
	h        int
	len      int
	children [2]*OrdMap
}

func (n *OrdMap) Height() int {
	if n == nil {
		return 0
	}
	return n.h
}

// suffix _OrdMap is needed because this will get specialised in codegen
func combinedDepth_OrdMap(n1, n2 *OrdMap) int {
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

// suffix _OrdMap is needed because this will get specialised in codegen
func mk_OrdMap(entry Entry, left *OrdMap, right *OrdMap) *OrdMap {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &OrdMap{
		Entry:    entry,
		h:        combinedDepth_OrdMap(left, right),
		len:      len,
		children: [2]*OrdMap{left, right},
	}
}

func (node *OrdMap) Get(key Key) (value Value, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if key.Less(finger.Entry.K) {
			finger = finger.children[0]
		} else if finger.Entry.K.Less(key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.Entry.V, true
		}
	}
}

func (node *OrdMap) Insert(key Key, value Value) *OrdMap {
	if node == nil {
		return mk_OrdMap(Entry{key, value}, nil, nil)
	}
	entry, left, right := node.Entry, node.children[0], node.children[1]
	if node.Entry.K.Less(key) {
		right = right.Insert(key, value)
	} else if key.Less(node.Entry.K) {
		left = left.Insert(key, value)
	} else { // equals
		entry = Entry{key, value}
	}
	return rotate_OrdMap(entry, left, right)
}

func (node *OrdMap) Remove(key Key) *OrdMap {
	if node == nil {
		return nil
	}
	entry, left, right := node.Entry, node.children[0], node.children[1]
	if node.Entry.K.Less(key) {
		right = right.Remove(key)
	} else if key.Less(node.Entry.K) {
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
	return rotate_OrdMap(entry, left, right)
}

// suffix _OrdMap is needed because this will get specialised in codegen
func rotate_OrdMap(entry Entry, left *OrdMap, right *OrdMap) *OrdMap {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepth_OrdMap(left, rl)-rr.Height() > 1 {
			// double rotation
			return mk_OrdMap(
				rl.Entry,
				mk_OrdMap(entry, left, rl.children[0]),
				mk_OrdMap(right.Entry, rl.children[1], rr),
			)
		}
		return mk_OrdMap(right.Entry, mk_OrdMap(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepth_OrdMap(right, lr)-ll.Height() > 1 {
			// double rotation
			return mk_OrdMap(
				lr.Entry,
				mk_OrdMap(left.Entry, ll, lr.children[0]),
				mk_OrdMap(entry, lr.children[1], right),
			)
		}
		return mk_OrdMap(left.Entry, ll, mk_OrdMap(entry, lr, right))
	}
	return mk_OrdMap(entry, left, right)
}

func (node *OrdMap) Len() int {
	if node == nil {
		return 0
	}
	return node.len
}

func (node *OrdMap) Entries() []Entry {
	elems := make([]Entry, 0, node.Len())
	if node == nil {
		return elems
	}
	type frame struct {
		node     *OrdMap
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
			elems = append(elems, top.node.Entry)
			if top.node.children[1] != nil {
				stack = append(stack, frame{top.node.children[1], false})
			}
		}
	}
	return elems
}

func (node *OrdMap) extreme(dir int) *Entry {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.Entry
}

func (node *OrdMap) Min() *Entry {
	return node.extreme(0)
}

func (node *OrdMap) Max() *Entry {
	return node.extreme(1)
}

func (node *OrdMap) Iterate() Iterator {
	return newIterator_OrdMap(node, 0, nil)
}

func (node *OrdMap) IterateFrom(k Key) Iterator {
	return newIterator_OrdMap(node, 0, &k)
}

func (node *OrdMap) IterateReverse() Iterator {
	return newIterator_OrdMap(node, 1, nil)
}

func (node *OrdMap) IterateReverseFrom(k Key) Iterator {
	return newIterator_OrdMap(node, 1, &k)
}

type iteratorStackFrame struct {
	node  *OrdMap
	state int8
}

type Iterator struct {
	direction    int
	stack        []iteratorStackFrame
	currentEntry Entry
}

// suffix _OrdMap is needed because this will get specialised in codegen
func newIterator_OrdMap(node *OrdMap, direction int, startFrom *Key) Iterator {
	if node == nil {
		return Iterator{}
	}
	stack := make([]iteratorStackFrame, 1, node.Height())
	stack[0] = iteratorStackFrame{node: node, state: 0}
	iter := Iterator{direction: direction, stack: stack}
	if startFrom != nil {
		stack[0].state = 2
		iter.seek(*startFrom)
	} else {
		iter.Next()
	}
	return iter
}

func (i *Iterator) Done() bool {
	return len(i.stack) == 0
}

func (i *Iterator) GetKey() Key {
	return i.currentEntry.K
}

func (i *Iterator) GetValue() Value {
	return i.currentEntry.V
}

func (i *Iterator) Next() {
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
			i.currentEntry = frame.node.Entry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = iteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		}

	}
}

func (i *Iterator) seek(k Key) {
LOOP:
	for {
		frame := &i.stack[len(i.stack)-1]
		if frame.node == nil {
			last := len(i.stack) - 1
			i.stack[last] = iteratorStackFrame{} // zero out
			i.stack = i.stack[:last]             // pop
			break LOOP
		}
		if (i.direction == 0 && !(frame.node.Entry.K.Less(k))) || (i.direction == 1 && !(k.Less(frame.node.Entry.K))) {
			i.stack = append(i.stack, iteratorStackFrame{node: frame.node.children[i.direction], state: 2})
		} else {
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = iteratorStackFrame{node: frame.node.children[1-i.direction], state: 2}
		}
	}
	if len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		i.currentEntry = frame.node.Entry
		frame.state = 3
	}
}
