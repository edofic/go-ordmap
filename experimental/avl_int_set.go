// DO NOT EDIT tis code was generated using go-ordmap code generation
// go run github.com/edofic/go-ordmap/cmd/gen -name AvlIntSet -key int -less < -value struct{} -target ./avl_int_set.go -pkg ordmap
package ordmap

type AvlIntSetEntry struct {
	K int
	V struct{}
}

type AvlIntSet struct {
	AvlIntSetEntry AvlIntSetEntry
	h              int
	len            int
	children       [2]*AvlIntSet
}

func (n *AvlIntSet) Height() int {
	if n == nil {
		return 0
	}
	return n.h
}

// suffix AvlIntSet is needed because this will get specialised in codegen
func combinedDepthAvlIntSet(n1, n2 *AvlIntSet) int {
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

// suffix AvlIntSet is needed because this will get specialised in codegen
func mkAvlIntSet(entry AvlIntSetEntry, left *AvlIntSet, right *AvlIntSet) *AvlIntSet {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &AvlIntSet{
		AvlIntSetEntry: entry,
		h:              combinedDepthAvlIntSet(left, right),
		len:            len,
		children:       [2]*AvlIntSet{left, right},
	}
}

func (node *AvlIntSet) Get(key int) (value struct{}, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if key < (finger.AvlIntSetEntry.K) {
			finger = finger.children[0]
		} else if finger.AvlIntSetEntry.K < (key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.AvlIntSetEntry.V, true
		}
	}
}

func (node *AvlIntSet) Insert(key int, value struct{}) *AvlIntSet {
	if node == nil {
		return mkAvlIntSet(AvlIntSetEntry{key, value}, nil, nil)
	}
	entry, left, right := node.AvlIntSetEntry, node.children[0], node.children[1]
	if node.AvlIntSetEntry.K < (key) {
		right = right.Insert(key, value)
	} else if key < (node.AvlIntSetEntry.K) {
		left = left.Insert(key, value)
	} else { // equals
		entry = AvlIntSetEntry{key, value}
	}
	return rotateAvlIntSet(entry, left, right)
}

func (node *AvlIntSet) Remove(key int) *AvlIntSet {
	if node == nil {
		return nil
	}
	entry, left, right := node.AvlIntSetEntry, node.children[0], node.children[1]
	if node.AvlIntSetEntry.K < (key) {
		right = right.Remove(key)
	} else if key < (node.AvlIntSetEntry.K) {
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
	return rotateAvlIntSet(entry, left, right)
}

// suffix AvlIntSet is needed because this will get specialised in codegen
func rotateAvlIntSet(entry AvlIntSetEntry, left *AvlIntSet, right *AvlIntSet) *AvlIntSet {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepthAvlIntSet(left, rl)-rr.Height() > 1 {
			// double rotation
			return mkAvlIntSet(
				rl.AvlIntSetEntry,
				mkAvlIntSet(entry, left, rl.children[0]),
				mkAvlIntSet(right.AvlIntSetEntry, rl.children[1], rr),
			)
		}
		return mkAvlIntSet(right.AvlIntSetEntry, mkAvlIntSet(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepthAvlIntSet(right, lr)-ll.Height() > 1 {
			// double rotation
			return mkAvlIntSet(
				lr.AvlIntSetEntry,
				mkAvlIntSet(left.AvlIntSetEntry, ll, lr.children[0]),
				mkAvlIntSet(entry, lr.children[1], right),
			)
		}
		return mkAvlIntSet(left.AvlIntSetEntry, ll, mkAvlIntSet(entry, lr, right))
	}
	return mkAvlIntSet(entry, left, right)
}

func (node *AvlIntSet) Len() int {
	if node == nil {
		return 0
	}
	return node.len
}

func (node *AvlIntSet) Entries() []AvlIntSetEntry {
	elems := make([]AvlIntSetEntry, 0)
	var step func(n *AvlIntSet)
	step = func(n *AvlIntSet) {
		if n == nil {
			return
		}
		step(n.children[0])
		elems = append(elems, n.AvlIntSetEntry)
		step(n.children[1])
	}
	step(node)
	return elems
}

func (node *AvlIntSet) extreme(dir int) *AvlIntSetEntry {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.AvlIntSetEntry
}

func (node *AvlIntSet) Min() *AvlIntSetEntry {
	return node.extreme(0)
}

func (node *AvlIntSet) Max() *AvlIntSetEntry {
	return node.extreme(1)
}

func (node *AvlIntSet) Iterate() AvlIntSetIterator {
	return newIteratorAvlIntSet(node, 0, nil)
}

func (node *AvlIntSet) IterateFrom(k int) AvlIntSetIterator {
	return newIteratorAvlIntSet(node, 0, &k)
}

func (node *AvlIntSet) IterateReverse() AvlIntSetIterator {
	return newIteratorAvlIntSet(node, 1, nil)
}

func (node *AvlIntSet) IterateReverseFrom(k int) AvlIntSetIterator {
	return newIteratorAvlIntSet(node, 1, &k)
}

type AvlIntSetIteratorStackFrame struct {
	node  *AvlIntSet
	state int8
}

type AvlIntSetIterator struct {
	direction    int
	stack        []AvlIntSetIteratorStackFrame
	currentEntry AvlIntSetEntry
}

// suffix AvlIntSet is needed because this will get specialised in codegen
func newIteratorAvlIntSet(node *AvlIntSet, direction int, startFrom *int) AvlIntSetIterator {
	if node == nil {
		return AvlIntSetIterator{}
	}
	stack := make([]AvlIntSetIteratorStackFrame, 1, node.Height())
	stack[0] = AvlIntSetIteratorStackFrame{node: node, state: 0}
	iter := AvlIntSetIterator{direction: direction, stack: stack}
	if startFrom != nil {
		stack[0].state = 2
		iter.seek(*startFrom)
	} else {
		iter.Next()
	}
	return iter
}

func (i *AvlIntSetIterator) Done() bool {
	return len(i.stack) == 0
}

func (i *AvlIntSetIterator) GetKey() int {
	return i.currentEntry.K
}

func (i *AvlIntSetIterator) GetValue() struct{} {
	return i.currentEntry.V
}

func (i *AvlIntSetIterator) Next() {
	for len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		switch frame.state {
		case 0:
			if frame.node == nil {
				last := len(i.stack) - 1
				i.stack[last] = AvlIntSetIteratorStackFrame{} // zero out
				i.stack = i.stack[:last]                      // pop
			} else {
				frame.state = 1
			}
		case 1:
			i.stack = append(i.stack, AvlIntSetIteratorStackFrame{node: frame.node.children[i.direction], state: 0})
			frame.state = 2
		case 2:
			i.currentEntry = frame.node.AvlIntSetEntry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = AvlIntSetIteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		}

	}
}

func (i *AvlIntSetIterator) seek(k int) {
LOOP:
	for {
		frame := &i.stack[len(i.stack)-1]
		if frame.node == nil {
			last := len(i.stack) - 1
			i.stack[last] = AvlIntSetIteratorStackFrame{} // zero out
			i.stack = i.stack[:last]                      // pop
			break LOOP
		}
		if (i.direction == 0 && !(frame.node.AvlIntSetEntry.K < (k))) || (i.direction == 1 && !(k < (frame.node.AvlIntSetEntry.K))) {
			i.stack = append(i.stack, AvlIntSetIteratorStackFrame{node: frame.node.children[i.direction], state: 2})
		} else {
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = AvlIntSetIteratorStackFrame{node: frame.node.children[1-i.direction], state: 2}
		}
	}
	if len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		i.currentEntry = frame.node.AvlIntSetEntry
		frame.state = 3
	}
}
