package ordmap

import "iter"

type Comparable[A any] interface {
	Less(A) bool
}

type BuiltinComparable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

type Builtin[A BuiltinComparable] struct {
	value A
}

func (b Builtin[A]) Less(b2 Builtin[A]) bool {
	return b.value < b2.value
}

func New[K Comparable[K], V any]() *Node[K, V] {
	return nil
}

func NewBuiltin[K BuiltinComparable, V any]() NodeBuiltin[K, V] {
	return NodeBuiltin[K, V]{nil}
}

type Entry[K, V any] struct {
	K K
	V V
}

type NodeBuiltin[K BuiltinComparable, V any] struct {
	n *Node[Builtin[K], V]
}

func (n NodeBuiltin[K, V]) Get(key K) (value V, ok bool) {
	return n.n.Get(Builtin[K]{key})
}

func (n NodeBuiltin[K, V]) Insert(key K, value V) NodeBuiltin[K, V] {
	return NodeBuiltin[K, V]{n.n.Insert(Builtin[K]{key}, value)}
}

func (n NodeBuiltin[K, V]) Remove(key K) NodeBuiltin[K, V] {
	return NodeBuiltin[K, V]{n.n.Remove(Builtin[K]{key})}
}

func (n NodeBuiltin[K, V]) Len() int {
	return n.n.Len()
}

func (n NodeBuiltin[K, V]) Entries() []Entry[K, V] {
	baseEntries := n.n.Entries()
	// TODO can this be unsafely cast?
	entries := make([]Entry[K, V], len(baseEntries))
	for i, e := range baseEntries {
		entries[i] = Entry[K, V]{e.K.value, e.V}
	}
	return entries
}

func (n NodeBuiltin[K, V]) Min() *Entry[K, V] {
	e := n.n.Min()
	if e == nil {
		return nil
	}
	return &Entry[K, V]{e.K.value, e.V}
}

func (n NodeBuiltin[K, V]) Max() *Entry[K, V] {
	e := n.n.Max()
	if e == nil {
		return nil
	}
	return &Entry[K, V]{e.K.value, e.V}
}

func (n NodeBuiltin[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range n.n.All() {
			if !yield(k.value, v) {
				return
			}
		}
	}
}

func (n NodeBuiltin[K, V]) Backward() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range n.n.Backward() {
			if !yield(k.value, v) {
				return
			}
		}
	}
}

func (n NodeBuiltin[K, V]) From(k K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for b, v := range n.n.From(Builtin[K]{k}) {
			if !yield(b.value, v) {
				return
			}
		}
	}
}

func (n NodeBuiltin[K, V]) BackwardFrom(k K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for b, v := range n.n.BackwardFrom(Builtin[K]{k}) {
			if !yield(b.value, v) {
				return
			}
		}
	}
}

type Node[K Comparable[K], V any] struct {
	entry    Entry[K, V]
	h        int
	len      int
	children [2]*Node[K, V]
}

func (node *Node[K, V]) height() int {
	if node == nil {
		return 0
	}
	return node.h
}

func combinedDepth[K Comparable[K], V any](n1, n2 *Node[K, V]) int {
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

func mk_OrdMap[K Comparable[K], V any](entry Entry[K, V], left, right *Node[K, V]) *Node[K, V] {
	len := 1
	if left != nil {
		len += left.len
	}
	if right != nil {
		len += right.len
	}
	return &Node[K, V]{
		entry:    entry,
		h:        combinedDepth(left, right),
		len:      len,
		children: [2]*Node[K, V]{left, right},
	}
}

func (node *Node[K, V]) Get(key K) (value V, ok bool) {
	finger := node
	for {
		if finger == nil {
			ok = false
			return // using named returns so we keep the zero value for `value`
		}
		if key.Less(finger.entry.K) {
			finger = finger.children[0]
		} else if finger.entry.K.Less(key) {
			finger = finger.children[1]
		} else {
			// equal
			return finger.entry.V, true
		}
	}
}

func (node *Node[K, V]) Insert(key K, value V) *Node[K, V] {
	if node == nil {
		return mk_OrdMap(Entry[K, V]{key, value}, nil, nil)
	}
	entry, left, right := node.entry, node.children[0], node.children[1]
	if node.entry.K.Less(key) {
		right = right.Insert(key, value)
	} else if key.Less(node.entry.K) {
		left = left.Insert(key, value)
	} else { // equals
		entry = Entry[K, V]{key, value}
	}
	return rotate(entry, left, right)
}

func (node *Node[K, V]) Remove(key K) *Node[K, V] {
	if node == nil {
		return nil
	}
	entry, left, right := node.entry, node.children[0], node.children[1]
	if node.entry.K.Less(key) {
		right = right.Remove(key)
	} else if key.Less(node.entry.K) {
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
	return rotate(entry, left, right)
}

func rotate[K Comparable[K], V any](entry Entry[K, V], left, right *Node[K, V]) *Node[K, V] {
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

func (node *Node[K, V]) Len() int {
	if node == nil {
		return 0
	}
	return node.len
}

type entriesFrame[K Comparable[K], V any] struct {
	node     *Node[K, V]
	leftDone bool
}

func (node *Node[K, V]) Entries() []Entry[K, V] {
	elems := make([]Entry[K, V], 0, node.Len())
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

func (node *Node[K, V]) extreme(dir int) *Entry[K, V] {
	if node == nil {
		return nil
	}
	finger := node
	for finger.children[dir] != nil {
		finger = finger.children[dir]
	}
	return &finger.entry
}

func (node *Node[K, V]) Min() *Entry[K, V] {
	return node.extreme(0)
}

func (node *Node[K, V]) Max() *Entry[K, V] {
	return node.extreme(1)
}

func (node *Node[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var step func(*Node[K, V]) bool
		step = func(node *Node[K, V]) bool {
			if node == nil {
				return true
			}
			if !step(node.children[0]) {
				return false
			}
			if !yield(node.entry.K, node.entry.V) {
				return false
			}
			if !step(node.children[1]) {
				return false
			}
			return true
		}
		step(node)
	}
}

func (node *Node[K, V]) Backward() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		var step func(*Node[K, V]) bool
		step = func(node *Node[K, V]) bool {
			if node == nil {
				return true
			}
			if !step(node.children[1]) {
				return false
			}
			if !yield(node.entry.K, node.entry.V) {
				return false
			}
			if !step(node.children[0]) {
				return false
			}
			return true
		}
		step(node)
	}
}

func (node *Node[K, V]) From(k K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// Phase 2: Unconditional Traversal
		// Used once we are strictly inside the valid range.
		// No comparisons against 'k' are performed here.
		var iterate func(*Node[K, V]) bool
		iterate = func(n *Node[K, V]) bool {
			if n == nil {
				return true
			}
			if !iterate(n.children[0]) {
				return false
			}
			if !yield(n.entry.K, n.entry.V) {
				return false
			}
			return iterate(n.children[1])
		}

		// Phase 1: Seeking the Boundary
		// Performs comparisons to prune left subtrees.
		var seek func(*Node[K, V]) bool
		seek = func(n *Node[K, V]) bool {
			if n == nil {
				return true
			}

			// If n < k:
			// This node and its entire left subtree are too small.
			// Skip them and continue seeking in the right subtree.
			if n.entry.K.Less(k) {
				return seek(n.children[1])
			}

			// If n >= k:
			// 1. The boundary might be in the left subtree, so we 'seek' left.
			if !seek(n.children[0]) {
				return false
			}

			// 2. This node is valid.
			if !yield(n.entry.K, n.entry.V) {
				return false
			}

			// 3. OPTIMIZATION:
			// Since n >= k, all nodes in the right subtree are > k.
			// Switch to unconditional 'iterate' (no comparisons).
			return iterate(n.children[1])
		}

		seek(node)
	}
}

func (node *Node[K, V]) BackwardFrom(k K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// Phase 2: Unconditional Reverse Traversal
		// Used for left subtrees of valid nodes (since Left < Node <= k).
		// Traverses: Right -> Node -> Left
		var iterate func(*Node[K, V]) bool
		iterate = func(n *Node[K, V]) bool {
			if n == nil {
				return true
			}
			// 1. Right (Largest)
			if !iterate(n.children[1]) {
				return false
			}
			// 2. Current
			if !yield(n.entry.K, n.entry.V) {
				return false
			}
			// 3. Left (Smallest)
			return iterate(n.children[0])
		}

		// Phase 1: Seeking the Boundary
		// Performs comparisons to prune Right subtrees.
		var seek func(*Node[K, V]) bool
		seek = func(n *Node[K, V]) bool {
			if n == nil {
				return true
			}

			// Check if Node > k.
			// (Assuming standard Less interface: k < n implies n > k)
			if k.Less(n.entry.K) {
				// This node is too big. The right subtree is even bigger.
				// Skip everything here and search the left subtree.
				return seek(n.children[0])
			}

			// If we are here, n <= k.

			// 1. The boundary might be deep in the Right subtree.
			//    (There could be nodes > n but still <= k)
			if !seek(n.children[1]) {
				return false
			}

			// 2. This node is valid.
			if !yield(n.entry.K, n.entry.V) {
				return false
			}

			// 3. OPTIMIZATION:
			// Since n <= k, all nodes in the Left subtree are < n.
			// Therefore, they are all < k.
			// Switch to unconditional 'iterate'.
			return iterate(n.children[0])
		}

		seek(node)
	}
}
