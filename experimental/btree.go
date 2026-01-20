package ordmap

import (
	"fmt"
	"iter"
)

const MAX = 5 // must be odd

type Key interface {
	Cmp(Key) int
}

type Value interface{}

type Entry struct {
	K Key
	V Value
}

var zeroEntry Entry

type OrdMap struct {
	order    uint8 // 1..MAX
	height   uint8
	len      int
	entries  [MAX]Entry
	subtrees [MAX + 1]*OrdMap
}

func mkOrdMap(order uint8, entries [MAX]Entry, subtrees [MAX + 1]*OrdMap) *OrdMap {
	height := uint8(1)
	len := int(order)
	for i := uint8(0); i <= order; i++ {
		if subtrees[i] != nil {
			if subtrees[i].height >= height {
				height = subtrees[i].height + 1
			}
			len += subtrees[i].len
		}
	}
	return &OrdMap{
		order:    order,
		height:   height,
		len:      len,
		entries:  entries,
		subtrees: subtrees,
	}
}

func (n *OrdMap) Entries() []Entry {
	entries := make([]Entry, 0, n.Len())
	var step func(n *OrdMap)
	step = func(n *OrdMap) {
		if n == nil {
			return
		}
		step(n.subtrees[0])
		for i := uint8(0); i < n.order; i++ {
			entries = append(entries, n.entries[i])
			step(n.subtrees[i+1])
		}
	}
	step(n)
	return entries
}

func (n *OrdMap) Get(key Key) (value Value, ok bool) {
	finger := n
OUTER:
	for finger != nil {
		for i := 0; i < int(finger.order); i++ {
			entry := finger.entries[i]
			cmp := key.Cmp(entry.K)
			if cmp == 0 {
				return entry.V, true
			}
			if cmp < 0 { // key < entry.K
				finger = finger.subtrees[i]
				continue OUTER
			}
		}
		finger = finger.subtrees[finger.order]
	}
	return value, false
}

func (n *OrdMap) Insert(key Key, value Value) *OrdMap {
	if n == nil {
		var entries [MAX]Entry
		entries[0] = Entry{key, value}
		return mkOrdMap(1, entries, [MAX + 1]*OrdMap{})
	}
	if n.order == MAX { // full root, need to split
		left, entry, right := n.split()
		var entries [MAX]Entry
		entries[0] = entry
		var subtrees [MAX + 1]*OrdMap
		subtrees[0] = left
		subtrees[1] = right
		n = mkOrdMap(1, entries, subtrees)
	} else {
		n = n.dup()
	}
	n.insertNonFullMut(key, value)
	return n
}

func (n *OrdMap) Remove(key Key) *OrdMap {
	if _, ok := n.Get(key); !ok {
		return n
	}
	if n.height == 1 && n.order == 1 && n.entries[0].K.Cmp(key) == 0 {
		return nil
	}
	n = n.dup()
	n.removeStepMut(key)
	return n
}

func (n *OrdMap) Min() *Entry {
	if n == nil {
		return nil
	}
	for finger := n; ; finger = finger.subtrees[0] {
		if finger.height == 1 {
			entry := finger.entries[0] // not taking address of an inner value directly
			return &entry
		}
	}
}

func (n *OrdMap) Max() *Entry {
	if n == nil {
		return nil
	}
	for finger := n; ; finger = finger.subtrees[finger.order] {
		if finger.height == 1 {
			entry := finger.entries[finger.order-1] // not taking address of an inner value directly
			return &entry
		}
	}
}

func (n *OrdMap) Height() int {
	if n == nil {
		return 0
	}
	return int(n.height)
}

func (n *OrdMap) Len() int {
	if n == nil {
		return 0
	}
	return n.len
}

func (n *OrdMap) All() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		var step func(*OrdMap) bool
		step = func(n *OrdMap) bool {
			if n == nil {
				return true
			}
			if !step(n.subtrees[0]) {
				return false
			}
			for i := uint8(0); i < n.order; i++ {
				if !yield(n.entries[i].K, n.entries[i].V) {
					return false
				}
				if !step(n.subtrees[i+1]) {
					return false
				}
			}
			return true
		}
		step(n)
	}
}

func (n *OrdMap) Backward() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		var step func(*OrdMap) bool
		step = func(n *OrdMap) bool {
			if n == nil {
				return true
			}
			if !step(n.subtrees[n.order]) {
				return false
			}
			for i := int(n.order) - 1; i >= 0; i-- {
				if !yield(n.entries[i].K, n.entries[i].V) {
					return false
				}
				if !step(n.subtrees[uint8(i)]) {
					return false
				}
			}
			return true
		}
		step(n)
	}
}

func (n *OrdMap) AllFrom(k Key) iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		// Phase 2: Unconditional Iterator
		// Standard B-Tree traversal: LeftSub -> Entry -> RightSub
		// No key comparisons performed here.
		var iterate func(*OrdMap) bool
		iterate = func(n *OrdMap) bool {
			if n == nil {
				return true
			}
			// 1. Traverse first subtree
			if !iterate(n.subtrees[0]) {
				return false
			}
			// 2. Traverse interleaved entries and subsequent subtrees
			for i := 0; i < int(n.order); i++ {
				if !yield(n.entries[i].K, n.entries[i].V) {
					return false
				}
				if !iterate(n.subtrees[i+1]) {
					return false
				}
			}
			return true
		}

		// Phase 1: Seek
		// Skips subtrees and entries that are strictly smaller than k.
		var seek func(*OrdMap) bool
		seek = func(n *OrdMap) bool {
			if n == nil {
				return true
			}

			// Iterate through the entries in this node
			for i := 0; i < int(n.order); i++ {
				entry := n.entries[i]

				// Compare k vs entry.K
				cmp := k.Cmp(entry.K)

				// Case 1: k > entry.K
				// The entry is strictly smaller than k.
				// Implication: subtree[i] is also strictly smaller.
				// Action: Skip both subtree[i] and entry[i].
				if cmp > 0 {
					continue
				}

				// Case 2: k <= entry.K
				// We found the crossover point!
				// entry[i] is the first entry in this node that is valid.

				// A. The boundary might be inside the left subtree (subtree[i]).
				//    It might contain keys >= k.
				if !seek(n.subtrees[i]) {
					return false
				}

				// B. Yield the current valid entry.
				if !yield(entry.K, entry.V) {
					return false
				}

				// C. OPTIMIZATION SWITCH
				// Since entry[i] >= k, we know subtree[i+1] and ALL subsequent
				// entries in this node are definitely > k.
				// We switch to unconditional 'iterate' for the rest of this node.

				// C1. Iterate the immediate right subtree
				if !iterate(n.subtrees[i+1]) {
					return false
				}

				// C2. Flush the remaining entries and subtrees in this node
				for j := i + 1; j < int(n.order); j++ {
					if !yield(n.entries[j].K, n.entries[j].V) {
						return false
					}
					if !iterate(n.subtrees[j+1]) {
						return false
					}
				}

				// We have fully processed this node and its relevant children.
				return true
			}

			// Case 3: We scanned all entries and all were < k.
			// However, valid nodes might still exist in the very last subtree
			// (the one to the right of the last entry).
			return seek(n.subtrees[n.order])
		}

		seek(n)
	}
}

func (n *OrdMap) BackwardFrom(k Key) iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		// Phase 2: Unconditional Backward Iterator
		// Traverses: Subtree[i+1] -> Entry[i] -> ... -> Subtree[0]
		// No key comparisons performed here.
		var iterate func(*OrdMap) bool
		iterate = func(n *OrdMap) bool {
			if n == nil {
				return true
			}
			// Loop from last entry to first
			for i := int(n.order) - 1; i >= 0; i-- {
				// 1. Visit right child of entry i
				if !iterate(n.subtrees[i+1]) {
					return false
				}
				// 2. Visit entry i
				if !yield(n.entries[i].K, n.entries[i].V) {
					return false
				}
			}
			// 3. Visit leftmost child
			return iterate(n.subtrees[0])
		}

		// Phase 1: Seek Backward
		// Prunes entries and subtrees strictly > k
		var seek func(*OrdMap) bool
		seek = func(n *OrdMap) bool {
			if n == nil {
				return true
			}

			// Loop backwards: check largest entries first
			for i := int(n.order) - 1; i >= 0; i-- {
				entry := n.entries[i]
				cmp := k.Cmp(entry.K)

				// Case 1: k < entry.K
				// This entry is too big.
				// Implication: The subtree to its right (subtrees[i+1]) is even bigger.
				// Action: Skip both. Continue loop to check smaller entries.
				if cmp < 0 {
					continue
				}

				// Case 2: k >= entry.K
				// This entry is valid (<= k).

				// A. The boundary is likely inside the right child (subtrees[i+1]).
				//    (It may contain values > entry.K but <= k)
				if !seek(n.subtrees[i+1]) {
					return false
				}

				// B. Yield the current valid entry
				if !yield(entry.K, entry.V) {
					return false
				}

				// C. OPTIMIZATION SWITCH
				// Since entry[i] <= k, everything to the LEFT of this entry
				// (subtree[i], entry[i-1]...) is strictly smaller than k.
				// Switch to unconditional 'iterate' for the rest of this node.

				// C1. Process remaining pairs (Subtree -> Entry) to the left
				for j := i - 1; j >= 0; j-- {
					if !iterate(n.subtrees[j+1]) { // Logic matches subtrees[i] relative to entry[i] from outer loop context
						return false
					}
					if !yield(n.entries[j].K, n.entries[j].V) {
						return false
					}
				}

				// C2. Process the final leftmost child
				return iterate(n.subtrees[0])
			}

			// Case 3: We scanned all entries and they were all > k (too big).
			// However, the very first subtree (subtrees[0]) contains values
			// smaller than entry[0], so it might contain valid items.
			return seek(n.subtrees[0])
		}

		seek(n)
	}
}

func (n *OrdMap) removeStepMut(key Key) {
OUTER:
	for {
		if n.height == 1 {
			for i := 0; i < int(n.order); i++ {
				if n.entries[i].K.Cmp(key) == 0 {
					top := int(n.order) - 1
					for j := i; j < top; j++ {
						n.entries[j] = n.entries[j+1]
					}
					n.entries[n.order-1] = zeroEntry
					n.order -= 1
					return
				}
			}
			return
		} else {
			index := int(n.order)
			for i := 0; i < int(n.order); i++ {
				if n.entries[i].K.Cmp(key) == 0 { // inner delete
					index = n.ensureChildNotMinimal(i + 1)
					if n.order == 0 { // degenerated, need to drop a level
						*n = *n.subtrees[0]
						continue OUTER
					}
					if index != i+1 || n.entries[i].K.Cmp(key) != 0 { // merge OR rotation
						continue OUTER // easiest to try again
					}
					child := n.subtrees[index].dup()
					min := child.popMinMut()
					n.subtrees[index] = child
					n.entries[i] = min
					return
				}
				if key.Cmp(n.entries[i].K) < 0 {
					index = i
					break
				}
			}
			index = n.ensureChildNotMinimal(index)
			if n.order == 0 { // degenerated, need to drop a level
				*n = *n.subtrees[0]
				continue OUTER
			}
			n.subtrees[index] = n.subtrees[index].dup()
			n = n.subtrees[index]
			continue OUTER
		}
	}
}

func (n *OrdMap) insertNonFullMut(key Key, value Value) {
OUTER:
	for {
		for i := 0; i < int(n.order); i++ {
			if n.entries[i].K.Cmp(key) == 0 {
				n.entries[i].V = value
				return
			}
		}
		if n.height == 1 {
			n.entries[n.order] = Entry{key, value}
			n.order += 1
			for i := int(n.order) - 1; i > 0; i-- {
				if n.entries[i].K.Cmp(n.entries[i-1].K) < 0 {
					n.entries[i], n.entries[i-1] = n.entries[i-1], n.entries[i]
				} else {
					break
				}
			}
			return
		}
		index := 0
		for i := 0; i < int(n.order); i++ {
			if key.Cmp(n.entries[i].K) > 0 {
				index = i + 1
			}
		}
		child := n.subtrees[index]
		if child.order == MAX { // full, need to split before entering
			left, entry, right := child.split()
			for i := int(n.order); i > index; i-- {
				n.entries[i] = n.entries[i-1]
			}
			n.entries[index] = entry
			for i := int(n.order); i > index; i-- {
				n.subtrees[i+1] = n.subtrees[i]
			}
			n.subtrees[index] = left
			n.subtrees[index+1] = right
			n.order += 1
			cmp := key.Cmp(entry.K)
			if cmp < 0 {
				n = left
				continue OUTER
			} else if cmp == 0 {
				n.entries[index].V = value
				return
			} else {
				n = right
				continue OUTER
			}
		} else {
			n.subtrees[index] = n.subtrees[index].dup()
			n = n.subtrees[index]
			continue OUTER
		}
	}
}

func (n *OrdMap) ensureChildNotMinimal(index int) int {
	if n.subtrees[index].order > 1 {
		return index
	}
	if index == 0 { // grab from the right
		if n.subtrees[1].order > 1 {
			child := n.subtrees[index].dup()
			neighbour := n.subtrees[1].dup()
			ne := neighbour.entries[0]
			child.entries[1] = n.entries[index]
			n.entries[index] = ne
			child.subtrees[2] = neighbour.subtrees[0]
			copy(neighbour.entries[:], neighbour.entries[1:])
			copy(neighbour.subtrees[:], neighbour.subtrees[1:])
			child.order += 1
			neighbour.order -= 1
			neighbour.entries[neighbour.order] = zeroEntry
			n.subtrees[0] = child
			n.subtrees[1] = neighbour
		} else { // right neighbour is minimal
			child := n.subtrees[index]
			neighbour := n.subtrees[1]
			var entries [MAX]Entry
			copy(entries[:], child.entries[:child.order])
			entries[child.order] = n.entries[index]
			copy(entries[child.order+1:], neighbour.entries[:neighbour.order])
			var subtrees [MAX + 1]*OrdMap
			copy(subtrees[:], child.subtrees[:child.order+1])
			copy(subtrees[child.order+1:], neighbour.subtrees[:neighbour.order+1])
			newChild := mkOrdMap(child.order+neighbour.order+1, entries, subtrees)
			n.subtrees[index] = newChild
			copy(n.subtrees[1:], n.subtrees[2:])
			copy(n.entries[0:], n.entries[1:])
			n.subtrees[n.order] = nil
			n.order -= 1
			n.entries[n.order] = zeroEntry
		}
	} else {
		child := n.subtrees[index]
		neighbour := n.subtrees[index-1]
		if neighbour.order > 1 {
			child = child.dup()
			neighbour = neighbour.dup()
			n.subtrees[index] = child
			n.subtrees[index-1] = neighbour
			copy(child.entries[1:], child.entries[:child.order])
			copy(child.subtrees[1:], child.subtrees[:child.order+1])
			child.order += 1
			child.entries[0] = n.entries[index-1]
			child.subtrees[0] = neighbour.subtrees[neighbour.order]
			n.entries[index-1] = neighbour.entries[neighbour.order-1]
			neighbour.subtrees[neighbour.order] = nil
			neighbour.order -= 1
			neighbour.entries[neighbour.order] = zeroEntry
		} else {
			var entries [MAX]Entry
			copy(entries[:], neighbour.entries[:neighbour.order])
			entries[neighbour.order] = n.entries[index-1]
			copy(entries[neighbour.order+1:], child.entries[:child.order])
			var subtrees [MAX + 1]*OrdMap
			copy(subtrees[:], neighbour.subtrees[:neighbour.order+1])
			copy(subtrees[neighbour.order+1:], child.subtrees[:child.order+1])
			newChild := mkOrdMap(child.order+neighbour.order+1, entries, subtrees)
			copy(n.subtrees[index-1:], n.subtrees[index:])
			n.subtrees[n.order] = nil
			n.subtrees[index-1] = newChild
			copy(n.entries[index-1:], n.entries[index:])
			n.order -= 1
			n.entries[n.order] = zeroEntry
			index -= 1
		}
	}
	return index
}

func (n *OrdMap) split() (left *OrdMap, entry Entry, right *OrdMap) {
	entry = n.entries[(MAX-1)/2]
	var leftEntries [MAX]Entry
	for i := 0; i < (MAX-1)/2; i++ {
		leftEntries[i] = n.entries[i]
	}
	var leftSubtrees [MAX + 1]*OrdMap
	for i := 0; i <= (MAX-1)/2; i++ {
		leftSubtrees[i] = n.subtrees[i]
	}
	left = mkOrdMap((MAX-1)/2, leftEntries, leftSubtrees)
	var rightEntries [MAX]Entry
	for i := (MAX + 1) / 2; i < MAX; i++ {
		rightEntries[i-(MAX+1)/2] = n.entries[i]
	}
	var rightSubtrees [MAX + 1]*OrdMap
	for i := (MAX + 1) / 2; i <= MAX; i++ {
		rightSubtrees[i-(MAX+1)/2] = n.subtrees[i]
	}
	right = mkOrdMap((MAX-1)/2, rightEntries, rightSubtrees)
	return
}

func (n *OrdMap) popMinMut() Entry {
OUTER:
	for {
		if n.height == 1 {
			e := n.entries[0]
			for i := 1; i < int(n.order); i++ {
				n.entries[i-1] = n.entries[i]
			}
			n.order -= 1
			n.entries[n.order] = zeroEntry
			return e
		}
		_ = n.ensureChildNotMinimal(0)
		n.subtrees[0] = n.subtrees[0].dup()
		n = n.subtrees[0]
		continue OUTER
	}
}

func (n OrdMap) dup() *OrdMap {
	return &n
}

func (n *OrdMap) visual() string {
	if n == nil {
		return "_"
	}
	s := "[ " + n.subtrees[0].visual()
	for i := 0; i < int(n.order); i++ {
		s += fmt.Sprintf(" %v %v", n.entries[i], n.subtrees[i+1].visual())
	}
	s += " ]"
	return s
}
