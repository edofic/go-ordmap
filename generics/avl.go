package main

type OrdMap[K, V any] struct {
	compare func(K, K) bool
	root    *node[K, V]
}

func NewOrdMap[K, V any](compare func(K, K) bool) OrdMap[K, V] {
	return OrdMap[K, V]{compare, nil}
}

func (o OrdMap[K, V]) Get(key K) (value V, ok bool) {
	return o.root.get(key, o.compare)
}

func (o OrdMap[K, V]) Insert(key K, value V) OrdMap[K, V] {
	return OrdMap[K, V]{o.compare, o.root.insert(key, value, o.compare)}
}

func (o OrdMap[K, V]) Remove(key K) OrdMap[K, V] {
	return OrdMap[K, V]{o.compare, o.root.remove(key, o.compare)}
}

func (o OrdMap[K, V]) Entries() []Entry[K, V] {
	return o.root.entries()
}

type Entry[K, V any] struct {
	K K
	V V
}

type node[K, V any] struct {
	entry    Entry[K, V]
	h        int
	len      int
	children [2]*node[K, V]
}

func (node *node[K, V]) height() int {
	if node == nil {
		return 0
	}
	return node.h
}

func combinedDepth[K, V any](n1, n2 *node[K, V]) int {
	d1 := n1.height()
	d2 := n2.height()
	var d int
	if d1 > d2 {
		d = d1
	} else {
		d = d2
	}
	return d + 1
}

func mk_OrdMap[K, V any](entry Entry[K, V], left, right *node[K, V]) *node[K, V] {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &node[K, V]{
		entry:    entry,
		h:        combinedDepth(left, right),
		len:      len,
		children: [2]*node[K, V]{left, right},
	}
}

func (node *node[K, V]) get(key K, compare func(K, K) bool) (value V, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if compare(key, finger.entry.K) {
			finger = finger.children[0]
		} else if compare(finger.entry.K, key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.entry.V, true
		}
	}
}

func (node *node[K, V]) insert(key K, value V, compare func(K, K) bool) *node[K, V] {
	if node == nil {
		return mk_OrdMap(Entry[K, V]{key, value}, nil, nil)
	}
	entry, left, right := node.entry, node.children[0], node.children[1]
	if compare(node.entry.K, key) {
		right = right.insert(key, value, compare)
	} else if compare(key, node.entry.K) {
		left = left.insert(key, value, compare)
	} else { // equals
		entry = Entry[K, V]{key, value}
	}
	return rotate(entry, left, right)
}

func (node *node[K, V]) remove(key K, compare func(K, K) bool) *node[K, V] {
	if node == nil {
		return nil
	}
	entry, left, right := node.entry, node.children[0], node.children[1]
	if compare(node.entry.K, key) {
		right = right.remove(key, compare)
	} else if compare(key, node.entry.K) {
		left = left.remove(key, compare)
	} else { // equals
		max := left.Max()
		if max == nil {
			return right
		} else {
			left = left.remove(max.K, compare)
			entry = *max
		}
	}
	return rotate(entry, left, right)
}

func rotate[K, V any](entry Entry[K, V], left, right *node[K, V]) *node[K, V] {
	if right.height()-left.height() > 1 { // implies right != nil
		// single left
		rl := right.children[0]
		rr := right.children[1]
		if combinedDepth(left, rl)-rr.height() > 1 {
			// double rotation
			return mk_OrdMap(
				rl.entry,
				mk_OrdMap(entry, left, rl.children[0]),
				mk_OrdMap(right.entry, rl.children[1], rr),
			)
		}
		return mk_OrdMap(right.entry, mk_OrdMap(entry, left, rl), rr)
	}
	if left.height()-right.height() > 1 { // implies left != nil
		// single right
		ll := left.children[0]
		lr := left.children[1]
		if combinedDepth(right, lr)-ll.height() > 1 {
			// double rotation
			return mk_OrdMap(
				lr.entry,
				mk_OrdMap(left.entry, ll, lr.children[0]),
				mk_OrdMap(entry, lr.children[1], right),
			)
		}
		return mk_OrdMap(left.entry, ll, mk_OrdMap(entry, lr, right))
	}
	return mk_OrdMap(entry, left, right)
}

func (node *node[K, V]) length() int {
	if node == nil {
		return 0
	}
	return node.len
}

type entriesFrame[K, V any] struct {
	node     *node[K, V]
	leftDone bool
}

func (node *node[K, V]) entries() []Entry[K, V] {
	elems := make([]Entry[K, V], 0, node.length())
	if node == nil {
		return elems
	}
	var preallocated [20]entriesFrame[K, V] // preallocate on stack for common case
	stack := preallocated[:0]
	stack = append(stack, entriesFrame[K, V]{node, false})
	for len(stack) > 0 {
		top := &stack[len(stack)-1]

		if !top.leftDone {
			if top.node.children[0] != nil {
				stack = append(stack, entriesFrame[K, V]{top.node.children[0], false})
			}
			top.leftDone = true
		} else {
			stack = stack[:len(stack)-1] // pop
			elems = append(elems, top.node.entry)
			if top.node.children[1] != nil {
				stack = append(stack, entriesFrame[K, V]{top.node.children[1], false})
			}
		}
	}
	return elems
}

func (node *node[K, V]) extreme(dir int) *Entry[K, V] {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.entry
}

func (node *node[K, V]) Min() *Entry[K, V] {
	return node.extreme(0)
}

func (node *node[K, V]) Max() *Entry[K, V] {
	return node.extreme(1)
}
