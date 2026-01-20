package generational

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Int int

func (i Int) Less(other Int) bool {
	return i < other
}

func TestGenerationalMap(t *testing.T) {
	// Limit 3 to force flushes
	m := New[Int, string](3)

	// 1. Insert into Young
	m = m.Insert(1, "one")
	m = m.Insert(2, "two")

	val, ok := m.Get(1)
	require.True(t, ok)
	require.Equal(t, "one", val)

	val, ok = m.Get(2)
	require.True(t, ok)
	require.Equal(t, "two", val)

	_, ok = m.Get(3)
	require.False(t, ok)

	// 2. Trigger Flush (limit 3, inserting 3rd item -> flush? No, check logic)
	// Logic: young.Len() >= limit.
	// Current young len = 2.
	// Insert 3 -> young len 3 -> flush -> young len 0, old len 3.
	m = m.Insert(3, "three")

	// Check all exist
	require.Equal(t, 0, m.young.Len())
	require.Equal(t, 3, m.old.Len())

	val, _ = m.Get(1)
	require.Equal(t, "one", val)
	val, _ = m.Get(3)
	require.Equal(t, "three", val)

	// 3. Shadowing (Update in Young, Old stays same until flush)
	m = m.Insert(1, "one-updated")
	require.Equal(t, 1, m.young.Len())

	val, _ = m.Get(1)
	require.Equal(t, "one-updated", val)

	// 4. Deletion (Tombstone in Young)
	m = m.Remove(2)                    // 2 was in Old
	require.Equal(t, 2, m.young.Len()) // {1: update, 2: delete}

	_, ok = m.Get(2)
	require.False(t, ok)

	// 5. Iterator (Live Merge)
	// Old: {1: one, 2: two, 3: three}
	// Young: {1: one-updated, 2: delete}
	// Expected: {1: one-updated, 3: three}

	var keys []int
	var values []string
	for k, v := range m.All() {
		keys = append(keys, int(k))
		values = append(values, v)
	}

	require.Equal(t, []int{1, 3}, keys)
	require.Equal(t, []string{"one-updated", "three"}, values)
}

func TestFlushLogic(t *testing.T) {
	m := New[Int, int](2)
	m = m.Insert(1, 1)
	m = m.Insert(2, 2) // Limit reached? 2 >= 2 -> Flush

	require.Equal(t, 0, m.young.Len())
	require.Equal(t, 2, m.old.Len())

	m = m.Remove(1)
	require.Equal(t, 1, m.young.Len()) // Tombstone

	m = m.Insert(3, 3) // len 2 -> flush
	require.Equal(t, 0, m.young.Len())

	// Old should have {2, 3}. 1 was removed.
	count := 0
	for range m.All() {
		count++
	}
	require.Equal(t, 2, count)

	_, ok := m.Get(1)
	require.False(t, ok)
}

func TestExtremesAndBackward(t *testing.T) {
	m := New[Int, string](5)

	// Insert 1, 2, 3, 4, 5
	m = m.Insert(1, "1")
	m = m.Insert(3, "3")
	m = m.Insert(5, "5")

	// Flush manually or via limit? limit is 5. len is 3. No flush.
	// I'll construct a scenario where some are in old, some in young.
	mOld := m.flush(m.young)
	// mOld now has empty young, old={1,3,5}

	m2 := mOld.Insert(2, "2")
	m2 = m2.Insert(4, "4")
	m2 = m2.Insert(6, "6")

	// m2: Old={1,3,5}, Young={2,4,6}
	// Merged: 1, 2, 3, 4, 5, 6

	// Test Min/Max
	min := m2.Min()
	require.NotNil(t, min)
	require.Equal(t, Int(1), min.K)

	max := m2.Max()
	require.NotNil(t, max)
	require.Equal(t, Int(6), max.K)

	// Test Backward
	var keys []int
	for k := range m2.Backward() {
		keys = append(keys, int(k))
	}
	require.Equal(t, []int{6, 5, 4, 3, 2, 1}, keys)

	// Test Min/Max with Deletions
	// Delete 1 (Min) and 6 (Max)
	m3 := m2.Remove(1)
	m3 = m3.Remove(6)

	// m3: Old={1,3,5}, Young={2,4,6, 1(del), 6(del)}
	// Expected: 2, 3, 4, 5

	min = m3.Min()
	require.NotNil(t, min)
	require.Equal(t, Int(2), min.K) // 1 is deleted

	max = m3.Max()
	require.NotNil(t, max)
	require.Equal(t, Int(5), max.K) // 6 is deleted

	keys = nil
	for k := range m3.Backward() {
		keys = append(keys, int(k))
	}
	require.Equal(t, []int{5, 4, 3, 2}, keys)
}

func TestPartialIteration(t *testing.T) {
	// Setup map with mixed content
	// Old Generation: 10, 20, 30, 40, 50
	// We force flush by making a temp map and using its flush logic logic is tricky to force without hitting limit
	// So let's just use limit 2 and insert carefully.
	m := New[Int, string](2)
	m = m.Insert(10, "10")
	m = m.Insert(20, "20")
	m = m.Insert(30, "30") // flushed. Old={10,20,30}, Young={}
	m = m.Insert(40, "40")
	m = m.Insert(50, "50") // flushed. Old={10,20,30,40,50}, Young={}

	// Young Generation adjustments:
	// Insert 25 (New)
	// Delete 40 (Tombstone)
	// Update 10 (Shadow)
	m = m.Insert(25, "25")
	m = m.Remove(40)
	m = m.Insert(10, "10-new")

	// State:
	// Old: {10: "10", 20: "20", 30: "30", 40: "40", 50: "50"}
	// Young: {10: "10-new", 25: "25", 40: DEL}

	// Effective Map:
	// 10: "10-new", 20: "20", 25: "25", 30: "30", 50: "50"

	// Test From(20) -> Should see 20, 25, 30, 50
	var keys []int
	for k := range m.From(20) {
		keys = append(keys, int(k))
	}
	require.Equal(t, []int{20, 25, 30, 50}, keys)

	// Test From(22) -> Should start at 25
	keys = nil
	for k := range m.From(22) {
		keys = append(keys, int(k))
	}
	require.Equal(t, []int{25, 30, 50}, keys)

	// Test BackwardFrom(30) -> Should see 30, 25, 20, 10
	keys = nil
	for k := range m.BackwardFrom(30) {
		keys = append(keys, int(k))
	}
	require.Equal(t, []int{30, 25, 20, 10}, keys)

	// Test BackwardFrom(45) -> Should start at 30 (40 is deleted, 50 > 45)
	keys = nil
	for k := range m.BackwardFrom(45) {
		keys = append(keys, int(k))
	}
	require.Equal(t, []int{30, 25, 20, 10}, keys)

	// Test BackwardFrom(40) -> 40 is deleted. Should start at 30.
	keys = nil
	for k := range m.BackwardFrom(40) {
		keys = append(keys, int(k))
	}
	require.Equal(t, []int{30, 25, 20, 10}, keys)
}