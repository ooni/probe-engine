// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/ooni/probe-engine/internal/oopsi/github.com/cheekybits/genny

package utils

// Linked list implementation from the Go standard library.

// PacketIntervalElement is an element of a linked list.
type PacketIntervalElement struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *PacketIntervalElement

	// The list to which this element belongs.
	list *PacketIntervalList

	// The value stored with this element.
	Value PacketInterval
}

// Next returns the next list element or nil.
func (e *PacketIntervalElement) Next() *PacketIntervalElement {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *PacketIntervalElement) Prev() *PacketIntervalElement {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// PacketIntervalList is a linked list of PacketIntervals.
type PacketIntervalList struct {
	root PacketIntervalElement // sentinel list element, only &root, root.prev, and root.next are used
	len  int                   // current list length excluding (this) sentinel element
}

// Init initializes or clears list l.
func (l *PacketIntervalList) Init() *PacketIntervalList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// NewPacketIntervalList returns an initialized list.
func NewPacketIntervalList() *PacketIntervalList { return new(PacketIntervalList).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *PacketIntervalList) Len() int { return l.len }

// Front returns the first element of list l or nil if the list is empty.
func (l *PacketIntervalList) Front() *PacketIntervalElement {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *PacketIntervalList) Back() *PacketIntervalElement {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero List value.
func (l *PacketIntervalList) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *PacketIntervalList) insert(e, at *PacketIntervalElement) *PacketIntervalElement {
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *PacketIntervalList) insertValue(v PacketInterval, at *PacketIntervalElement) *PacketIntervalElement {
	return l.insert(&PacketIntervalElement{Value: v}, at)
}

// remove removes e from its list, decrements l.len, and returns e.
func (l *PacketIntervalList) remove(e *PacketIntervalElement) *PacketIntervalElement {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.len--
	return e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *PacketIntervalList) Remove(e *PacketIntervalElement) PacketInterval {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *PacketIntervalList) PushFront(v PacketInterval) *PacketIntervalElement {
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *PacketIntervalList) PushBack(v PacketInterval) *PacketIntervalElement {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *PacketIntervalList) InsertBefore(v PacketInterval, mark *PacketIntervalElement) *PacketIntervalElement {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *PacketIntervalList) InsertAfter(v PacketInterval, mark *PacketIntervalElement) *PacketIntervalElement {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *PacketIntervalList) MoveToFront(e *PacketIntervalElement) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.insert(l.remove(e), &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *PacketIntervalList) MoveToBack(e *PacketIntervalElement) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.insert(l.remove(e), l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *PacketIntervalList) MoveBefore(e, mark *PacketIntervalElement) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.insert(l.remove(e), mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *PacketIntervalList) MoveAfter(e, mark *PacketIntervalElement) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.insert(l.remove(e), mark)
}

// PushBackList inserts a copy of an other list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *PacketIntervalList) PushBackList(other *PacketIntervalList) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of an other list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *PacketIntervalList) PushFrontList(other *PacketIntervalList) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}
