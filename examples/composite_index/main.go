//go:generate go run github.com/edofic/go-ordmap/cmd/gen -name Index -key CompositeKey -value bool -target ./composite.go
package main

import (
	"fmt"
)

type CompositeKey struct {
	User       int
	Preference int
}

func (c CompositeKey) Less(c2 CompositeKey) bool {
	if c.User < c2.User {
		return true
	}
	if c.User > c2.User {
		return false
	}
	return c.Preference < c2.Preference
}

func main() {
	// can be used as a sql table (user, preference, is_set) with a composite index on (user, preference)
	var preferences *Index
	preferences = preferences.Insert(CompositeKey{2, 3}, true)
	preferences = preferences.Insert(CompositeKey{1, 1}, true)
	preferences = preferences.Insert(CompositeKey{2, 1}, false)
	preferences = preferences.Insert(CompositeKey{1, 3}, false)
	preferences = preferences.Insert(CompositeKey{2, 2}, false)
	preferences = preferences.Insert(CompositeKey{1, 2}, false)
	fmt.Println(preferences.Entries())               // can get everything
	fmt.Println(preferences.Get(CompositeKey{1, 1})) // can read as usual
	fmt.Println(preferences.Get(CompositeKey{1, 4}))

	// but with a bit of cunningnes you can also query a prefix
	// e.g all preferences for user 2
	targetUser := 2
	for i := preferences.IterateFrom(CompositeKey{targetUser, 0}); !i.Done() && i.GetKey().User == targetUser; i.Next() {
		fmt.Println(i.GetKey(), i.GetValue())
	}
}
