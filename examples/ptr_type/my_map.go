// DO NOT EDIT tis code was generated using go-ordmap code generation
// go run github.com/edofic/go-ordmap/cmd/gen -name MyMap -key int -less < -value *MyValue -target ./my_map.go
package main

type MyMapEntry struct {
	K int
	V *MyValue
}

type MyMap struct {
	MyMapEntry MyMapEntry
	h          int
	len        int
	children   [2]*MyMap
}

func (n *MyMap) Height() int {
	if n == nil {
		return 0
	}
	return n.h
}

// suffix MyMap is needed because this will get specialised in codegen
func combinedDepthMyMap(n1, n2 *MyMap) int {
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

// suffix MyMap is needed because this will get specialised in codegen
func mkMyMap(entry MyMapEntry, left *MyMap, right *MyMap) *MyMap {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &MyMap{
		MyMapEntry: entry,
		h:          combinedDepthMyMap(left, right),
		len:        len,
		children:   [2]*MyMap{left, right},
	}
}

func (node *MyMap) Get(key int) (value *MyValue, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if key < (finger.MyMapEntry.K) {
			finger = finger.children[0]
		} else if finger.MyMapEntry.K < (key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.MyMapEntry.V, true
		}
	}
}

func (node *MyMap) Insert(key int, value *MyValue) *MyMap {
	if node == nil {
		return mkMyMap(MyMapEntry{key, value}, nil, nil)
	}
	entry, left, right := node.MyMapEntry, node.children[0], node.children[1]
	if node.MyMapEntry.K < (key) {
		right = right.Insert(key, value)
	} else if key < (node.MyMapEntry.K) {
		left = left.Insert(key, value)
	} else { // equals
		entry = MyMapEntry{key, value}
	}
	return rotateMyMap(entry, left, right)
}

func (node *MyMap) Remove(key int) *MyMap {
	if node == nil {
		return nil
	}
	entry, left, right := node.MyMapEntry, node.children[0], node.children[1]
	if node.MyMapEntry.K < (key) {
		right = right.Remove(key)
	} else if key < (node.MyMapEntry.K) {
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
	return rotateMyMap(entry, left, right)
}

// suffix MyMap is needed because this will get specialised in codegen
func rotateMyMap(entry MyMapEntry, left *MyMap, right *MyMap) *MyMap {
	if right.Height()-left.Height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepthMyMap(left, rl)-rr.Height() > 1 {
			// double rotation
			return mkMyMap(
				rl.MyMapEntry,
				mkMyMap(entry, left, rl.children[0]),
				mkMyMap(right.MyMapEntry, rl.children[1], rr),
			)
		}
		return mkMyMap(right.MyMapEntry, mkMyMap(entry, left, rl), rr)
	}
	if left.Height()-right.Height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepthMyMap(right, lr)-ll.Height() > 1 {
			// double rotation
			return mkMyMap(
				lr.MyMapEntry,
				mkMyMap(left.MyMapEntry, ll, lr.children[0]),
				mkMyMap(entry, lr.children[1], right),
			)
		}
		return mkMyMap(left.MyMapEntry, ll, mkMyMap(entry, lr, right))
	}
	return mkMyMap(entry, left, right)
}

func (node *MyMap) Len() int {
	if node == nil {
		return 0
	}
	return node.len
}

func (node *MyMap) Entries() []MyMapEntry {
	elems := make([]MyMapEntry, 0, node.Len())
	var step func(n *MyMap)
	step = func(n *MyMap) {
		if n == nil {
			return
		}
		step(n.children[0])
		elems = append(elems, n.MyMapEntry)
		step(n.children[1])
	}
	step(node)
	return elems
}

func (node *MyMap) extreme(dir int) *MyMapEntry {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.MyMapEntry
}

func (node *MyMap) Min() *MyMapEntry {
	return node.extreme(0)
}

func (node *MyMap) Max() *MyMapEntry {
	return node.extreme(1)
}

func (node *MyMap) Iterate() MyMapIterator {
	return newIteratorMyMap(node, 0, nil)
}

func (node *MyMap) IterateFrom(k int) MyMapIterator {
	return newIteratorMyMap(node, 0, &k)
}

func (node *MyMap) IterateReverse() MyMapIterator {
	return newIteratorMyMap(node, 1, nil)
}

func (node *MyMap) IterateReverseFrom(k int) MyMapIterator {
	return newIteratorMyMap(node, 1, &k)
}

type MyMapIteratorStackFrame struct {
	node  *MyMap
	state int8
}

type MyMapIterator struct {
	direction    int
	stack        []MyMapIteratorStackFrame
	currentEntry MyMapEntry
}

// suffix MyMap is needed because this will get specialised in codegen
func newIteratorMyMap(node *MyMap, direction int, startFrom *int) MyMapIterator {
	if node == nil {
		return MyMapIterator{}
	}
	stack := make([]MyMapIteratorStackFrame, 1, node.Height())
	stack[0] = MyMapIteratorStackFrame{node: node, state: 0}
	iter := MyMapIterator{direction: direction, stack: stack}
	if startFrom != nil {
		stack[0].state = 2
		iter.seek(*startFrom)
	} else {
		iter.Next()
	}
	return iter
}

func (i *MyMapIterator) Done() bool {
	return len(i.stack) == 0
}

func (i *MyMapIterator) GetKey() int {
	return i.currentEntry.K
}

func (i *MyMapIterator) GetValue() *MyValue {
	return i.currentEntry.V
}

func (i *MyMapIterator) Next() {
	for len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		switch frame.state {
		case 0:
			if frame.node == nil {
				last := len(i.stack) - 1
				i.stack[last] = MyMapIteratorStackFrame{} // zero out
				i.stack = i.stack[:last]                  // pop
			} else {
				frame.state = 1
			}
		case 1:
			i.stack = append(i.stack, MyMapIteratorStackFrame{node: frame.node.children[i.direction], state: 0})
			frame.state = 2
		case 2:
			i.currentEntry = frame.node.MyMapEntry
			frame.state = 3
			return
		case 3:
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = MyMapIteratorStackFrame{node: frame.node.children[1-i.direction], state: 0}
		}

	}
}

func (i *MyMapIterator) seek(k int) {
LOOP:
	for {
		frame := &i.stack[len(i.stack)-1]
		if frame.node == nil {
			last := len(i.stack) - 1
			i.stack[last] = MyMapIteratorStackFrame{} // zero out
			i.stack = i.stack[:last]                  // pop
			break LOOP
		}
		if (i.direction == 0 && !(frame.node.MyMapEntry.K < (k))) || (i.direction == 1 && !(k < (frame.node.MyMapEntry.K))) {
			i.stack = append(i.stack, MyMapIteratorStackFrame{node: frame.node.children[i.direction], state: 2})
		} else {
			// override frame - tail call optimisation
			i.stack[len(i.stack)-1] = MyMapIteratorStackFrame{node: frame.node.children[1-i.direction], state: 2}
		}
	}
	if len(i.stack) > 0 {
		frame := &i.stack[len(i.stack)-1]
		i.currentEntry = frame.node.MyMapEntry
		frame.state = 3
	}
}
