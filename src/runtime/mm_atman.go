package runtime

import "unsafe"

// pageFrameMap maps frames to pages.
//
// The pages are expected to be stored contiguously in memory,
// starting at the address provided.
type pageFrameMap uintptr

// Get returns the page structure for the frame f.
//
// If f is out-of-bounds, the return value is undefined.
func (m pageFrameMap) Get(f frame) *page {
	return (*page)(unsafe.Pointer(uintptr(m) + uintptr(f)*unsafe.Sizeof(page{})))
}

var (
	errNoMem = errorString("mm: no free page")
)

// frame is a pseudo-physical frame number.
type frame uintptr

// frameAllocator allocates contiguous frames
// starting from Next, until Available is 0.
//
// It's primary use is for allocating the structures required
// by the memory manager during kernel bootstrap.
type frameAllocator struct {
	Next      frame
	Available int
}

// Alloc returns the number of the next free page,
// or an error if there are no free pages.
func (a *frameAllocator) Alloc() (frame, error) {
	if a.Available == 0 {
		return 0, errNoMem
	}

	f := a.Next

	a.Next++
	a.Available--

	return f, nil
}

type page struct {
	frame

	// Next is used to form pages into a linked list.
	Next *page
}

// pageQueue is a FIFO queue of pages.
type pageQueue struct {
	head *page
	tail *page
}

// Push adds p to the end of q.
func (q *pageQueue) Push(p *page) {
	if q.tail == nil {
		q.head = p
	} else {
		q.tail.Next = p
	}

	q.tail = p
}

// Pop returns the next available page.
func (q *pageQueue) Pop() *page {
	if q.head == nil {
		return nil
	}

	p := q.head

	q.head = p.Next
	if p.Next == nil {
		q.tail = nil
	}

	return p
}

// pageList is a linked list of pages.
type pageList struct {
	head *page
}

// Add adds p to l.
func (l *pageList) Add(p *page) {
	p.Next = l.head
	l.head = p
}

// Remove removes p from l, and returns true if successful.
func (l *pageList) Remove(p *page) bool {
	if l.head == nil {
		return false
	}

	if l.head == p {
		l.head = p.Next
		p.Next = nil
		return true
	}

	prev := l.head

	for e := prev.Next; e != nil; e = prev.Next {
		if e != p {
			prev = p
			continue
		}

		prev.Next = e.Next
		p.Next = nil
		return true
	}

	return false
}

// pageAllocator allocates frames from an LRU free list
// of pages.
type pageAllocator struct {
	free pageQueue
	used pageList
}

// Alloc returns the next available page,
// or an error if there are no free pages.
func (a *pageAllocator) Alloc() (*page, error) {
	p := a.free.Pop()
	if p == nil {
		return nil, errNoMem
	}
	a.used.Add(p)
	return p, nil
}

// Free adds p to the list of available pages.
//
// Attempting to free an unused page will cause a panic.
func (a *pageAllocator) Free(p *page) {
	if !a.used.Remove(p) {
		panic("attempt to free unknown page")
	}
	a.free.Push(p)
}
