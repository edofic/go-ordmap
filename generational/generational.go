package generational

import (
	"iter"

	"github.com/edofic/go-ordmap/v2"
)

type operation[V any] struct {
	value  V
	delete bool
}

type Map[K ordmap.Comparable[K], V any] struct {
	young *ordmap.Node[K, operation[V]]
	old   *ordmap.Node[K, V]
	limit int
}

func New[K ordmap.Comparable[K], V any](limit int) *Map[K, V] {
	return &Map[K, V]{
		young: ordmap.New[K, operation[V]](),
		old:   ordmap.New[K, V](),
		limit: limit,
	}
}

func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	if m == nil {
		return
	}
	op, ok := m.young.Get(key)
	if ok {
		if op.delete {
			return value, false
		}
		return op.value, true
	}
	return m.old.Get(key)
}

func (m *Map[K, V]) Insert(key K, value V) *Map[K, V] {
	op := operation[V]{value: value}
	young := m.young.Insert(key, op)
	if young.Len() >= m.limit {
		return m.flush(young)
	}
	return &Map[K, V]{
		young: young,
		old:   m.old,
		limit: m.limit,
	}
}

func (m *Map[K, V]) Remove(key K) *Map[K, V] {
	op := operation[V]{delete: true}
	young := m.young.Insert(key, op)
	if young.Len() >= m.limit {
		return m.flush(young)
	}
	return &Map[K, V]{
		young: young,
		old:   m.old,
		limit: m.limit,
	}
}

func (m *Map[K, V]) flush(young *ordmap.Node[K, operation[V]]) *Map[K, V] {
	old := m.old
	for k, op := range young.All() {
		if op.delete {
			old = old.Remove(k)
		} else {
			old = old.Insert(k, op.value)
		}
	}
	return &Map[K, V]{
		young: ordmap.New[K, operation[V]](),
		old:   old,
		limit: m.limit,
	}
}

func mergeForward[K ordmap.Comparable[K], V any](
	youngSeq iter.Seq2[K, operation[V]],
	oldSeq iter.Seq2[K, V],
) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		nextYoung, stopYoung := iter.Pull2(youngSeq)
		defer stopYoung()
		nextOld, stopOld := iter.Pull2(oldSeq)
		defer stopOld()

		kYoung, opYoung, okYoung := nextYoung()
		kOld, vOld, okOld := nextOld()

		for okYoung && okOld {
			if kYoung.Less(kOld) {
				if !opYoung.delete {
					if !yield(kYoung, opYoung.value) {
						return
					}
				}
				kYoung, opYoung, okYoung = nextYoung()
			} else if kOld.Less(kYoung) {
				if !yield(kOld, vOld) {
					return
				}
				kOld, vOld, okOld = nextOld()
			} else {
				// Equal keys: young shadows old
				if !opYoung.delete {
					if !yield(kYoung, opYoung.value) {
						return
					}
				}
				kYoung, opYoung, okYoung = nextYoung()
				kOld, vOld, okOld = nextOld()
			}
		}

		for okYoung {
			if !opYoung.delete {
				if !yield(kYoung, opYoung.value) {
					return
				}
			}
			kYoung, opYoung, okYoung = nextYoung()
		}

		for okOld {
			if !yield(kOld, vOld) {
				return
			}
			kOld, vOld, okOld = nextOld()
		}
	}
}

func mergeBackward[K ordmap.Comparable[K], V any](
	youngSeq iter.Seq2[K, operation[V]],
	oldSeq iter.Seq2[K, V],
) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		nextYoung, stopYoung := iter.Pull2(youngSeq)
		defer stopYoung()
		nextOld, stopOld := iter.Pull2(oldSeq)
		defer stopOld()

		kYoung, opYoung, okYoung := nextYoung()
		kOld, vOld, okOld := nextOld()

		for okYoung && okOld {
			// In backward traversal, we want larger keys first.
			// If kOld < kYoung, then kYoung > kOld, so process Young.
			if kOld.Less(kYoung) {
				if !opYoung.delete {
					if !yield(kYoung, opYoung.value) {
						return
					}
				}
				kYoung, opYoung, okYoung = nextYoung()
			} else if kYoung.Less(kOld) {
				// kYoung < kOld, so kOld > kYoung, process Old.
				if !yield(kOld, vOld) {
					return
				}
				kOld, vOld, okOld = nextOld()
			} else {
				// Equal
				if !opYoung.delete {
					if !yield(kYoung, opYoung.value) {
						return
					}
				}
				kYoung, opYoung, okYoung = nextYoung()
				kOld, vOld, okOld = nextOld()
			}
		}

		for okYoung {
			if !opYoung.delete {
				if !yield(kYoung, opYoung.value) {
					return
				}
			}
			kYoung, opYoung, okYoung = nextYoung()
		}

		for okOld {
			if !yield(kOld, vOld) {
				return
			}
			kOld, vOld, okOld = nextOld()
		}
	}
}

func (m *Map[K, V]) All() iter.Seq2[K, V] {
	if m == nil {
		return func(func(K, V) bool) {}
	}
	return mergeForward(m.young.All(), m.old.All())
}

func (m *Map[K, V]) Backward() iter.Seq2[K, V] {
	if m == nil {
		return func(func(K, V) bool) {}
	}
	return mergeBackward(m.young.Backward(), m.old.Backward())
}

func (m *Map[K, V]) From(k K) iter.Seq2[K, V] {
	if m == nil {
		return func(func(K, V) bool) {}
	}
	return mergeForward(m.young.From(k), m.old.From(k))
}

func (m *Map[K, V]) BackwardFrom(k K) iter.Seq2[K, V] {
	if m == nil {
		return func(func(K, V) bool) {}
	}
	return mergeBackward(m.young.BackwardFrom(k), m.old.BackwardFrom(k))
}

func (m *Map[K, V]) Min() *ordmap.Entry[K, V] {
	for k, v := range m.All() {
		return &ordmap.Entry[K, V]{K: k, V: v}
	}
	return nil
}

func (m *Map[K, V]) Max() *ordmap.Entry[K, V] {
	for k, v := range m.Backward() {
		return &ordmap.Entry[K, V]{K: k, V: v}
	}
	return nil
}
