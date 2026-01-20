package main

import (
	"fmt"

	"github.com/edofic/go-ordmap/v2/generational"
)

// Define a custom comparable type.
// ordmap.Builtin is internal/helper oriented and has unexported fields,
// so for generic usage, it's best to define your own domain types.
type Int int

func (i Int) Less(other Int) bool {
	return i < other
}

func main() {
	// Create a generational map with a small limit for the "young" generation (buffer).
	// In a real scenario, this limit would be much larger (e.g., thousands).
	// When the young generation exceeds this limit, it is merged (flushed) into the old generation.
	m := generational.New[Int, string](3)

	fmt.Println("--- Insertions ---")
	m = m.Insert(1, "one")
	m = m.Insert(2, "two")
	m = m.Insert(3, "three") // Triggers flush (size 3 >= 3)
	m = m.Insert(4, "four")  // Starts filling new young gen
	
	fmt.Println("Inserted 1, 2, 3 (flushed), and 4.")

	// Retrieve items
	fmt.Println("\n--- Retrieval ---")
	if v, ok := m.Get(1); ok {
		fmt.Println("Get(1):", v) // From Old generation
	}
	if v, ok := m.Get(4); ok {
		fmt.Println("Get(4):", v) // From Young generation
	}

	// Update (Shadowing)
	fmt.Println("\n--- Shadowing (Update) ---")
	m = m.Insert(2, "two-updated") // Writes to Young, shadows Old
	if v, ok := m.Get(2); ok {
		fmt.Println("Get(2):", v)
	}

	// Deletion (Tombstone)
	fmt.Println("\n--- Deletion ---")
	m = m.Remove(1) // Writes Tombstone to Young
	if _, ok := m.Get(1); !ok {
		fmt.Println("Get(1): Not Found (Correctly deleted via tombstone)")
	}

	// Iteration (Live Merge)
	fmt.Println("\n--- Iteration (All) ---")
	// Should see: 2:two-updated, 3:three, 4:four. (1 is deleted)
	for k, v := range m.All() {
		fmt.Printf("Key: %d, Value: %s\n", k, v)
	}

	// Backward Iteration
	fmt.Println("\n--- Backward Iteration ---")
	for k, v := range m.Backward() {
		fmt.Printf("Key: %d, Value: %s\n", k, v)
	}

	// Partial Iteration
	fmt.Println("\n--- Partial Iteration (From 3) ---")
	for k, v := range m.From(3) {
		fmt.Printf("Key: %d, Value: %s\n", k, v)
	}
}
