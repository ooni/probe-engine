// Package atomicx contains atomic extensions. The main reason for rolling out
// this package is to avoid potential crashes when using 32 bit devices where we
// are atomically accessing a 64 bit variable that is not aligned. The solution
// to this issue is rather crude: use a normal variable and protect it using a
// normal mutex. While this could be disappointing in general, it seems fine to
// be done in our context where we mainly use atomic semantics for counting.
package atomicx

import (
	"sync"
)

// Int64 is an int64 with atomic semantics.
type Int64 struct {
	mu sync.Mutex
	v  int64
}

// NewInt64 creates a new int64 with atomic semantics.
func NewInt64() *Int64 {
	return new(Int64)
}

// Add behaves like atomic.AddInt64
func (i64 *Int64) Add(delta int64) (newvalue int64) {
	i64.mu.Lock()
	i64.v += delta
	newvalue = i64.v
	i64.mu.Unlock()
	return
}

// Load behaves like atomic.LoadInt64
func (i64 *Int64) Load() (v int64) {
	i64.mu.Lock()
	v = i64.v
	i64.mu.Unlock()
	return
}
