package ordmap

type Entry struct {
	Key   Key
	Value Value
}

type OrdMap struct {
	Entry    Entry
	h        int
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
	return &OrdMap{
		Entry:    entry,
		h:        combinedDepth_OrdMap(left, right),
		children: [2]*OrdMap{left, right},
	}
}

func (node *OrdMap) Get(key Key) (value Value, ok bool) {
	if node == nil {
		ok = false
		return // using named returns so we keep the zero value for `value`
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

func (node *OrdMap) Insert(key Key, value Value) *OrdMap {
	if node == nil {
		return mk_OrdMap(Entry{key, value}, nil, nil)
	}
	entry, left, right := node.Entry, node.children[0], node.children[1]
	if node.Entry.Key.Less(key) {
		right = right.Insert(key, value)
	} else if key.Less(node.Entry.Key) {
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
	return 1 + node.children[0].Len() + node.children[1].Len()
}

func (node *OrdMap) Entries() []Entry {
	elems := make([]Entry, 0)
	var step func(n *OrdMap)
	step = func(n *OrdMap) {
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
	return newIterator_OrdMap(node, 0)
}

func (node *OrdMap) IterateReverse() Iterator {
	return newIterator_OrdMap(node, 1)
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
func newIterator_OrdMap(node *OrdMap, direction int) Iterator {
	stack := make([]iteratorStackFrame, 1, node.Height())
	stack[0] = iteratorStackFrame{node: node, state: 0}
	iter := Iterator{direction: direction, stack: stack}
	iter.Next()
	return iter
}

func (i *Iterator) Done() bool {
	return len(i.stack) == 0
}

func (i *Iterator) GetKey() Key {
	return i.currentEntry.Key
}

func (i *Iterator) GetValue() Value {
	return i.currentEntry.Value
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
