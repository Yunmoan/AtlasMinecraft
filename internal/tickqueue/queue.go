// Package tickqueue implements the 1.21.1 scheduled-tick containers. It does
// not decide which chunks may tick, advance world time, or execute block logic.
// The level owner must serialize all access, including callbacks during Tick.
package tickqueue

type Position struct{ X, Y, Z int32 }

func ChunkKey(x, z int32) uint64 { return uint64(uint32(x)) | uint64(uint32(z))<<32 }
func (p Position) chunk() uint64 { return ChunkKey(p.X>>4, p.Z>>4) }

type Priority int8

// ClampPriority is TickPriority.byValue, used when reading saved NBT values.
func ClampPriority(value int32) Priority { return Priority(max(-3, min(3, value))) }

type Saved[T comparable] struct {
	Type     T
	Position Position
	Delay    int32
	Priority Priority
}

type Scheduled[T comparable] struct {
	Type     T
	Position Position
	Trigger  int64
	Priority Priority
	Order    int64
}

type identity[T comparable] struct {
	kind     T
	position Position
}

func (t Scheduled[T]) id() identity[T] { return identity[T]{t.Type, t.Position} }
func (t Saved[T]) id() identity[T]     { return identity[T]{t.Type, t.Position} }

// Chunk retains the native heap iteration order when saving. Sorting that
// output changes the negative order numbers assigned on the next load.
type Chunk[T comparable] struct {
	queue   []Scheduled[T]
	pending []Saved[T]
	unique  map[identity[T]]struct{}
	onAdded func(Scheduled[T])
}

func NewChunk[T comparable](saved []Saved[T]) *Chunk[T] {
	c := &Chunk[T]{pending: append([]Saved[T](nil), saved...), unique: make(map[identity[T]]struct{})}
	for _, t := range saved {
		c.unique[t.id()] = struct{}{}
	}
	return c
}

func intraCompare[T comparable](a, b Scheduled[T]) int {
	if a.Priority < b.Priority {
		return -1
	}
	if a.Priority > b.Priority {
		return 1
	}
	if a.Order < b.Order {
		return -1
	}
	if a.Order > b.Order {
		return 1
	}
	return 0
}
func drainCompare[T comparable](a, b Scheduled[T]) int {
	if a.Trigger < b.Trigger {
		return -1
	}
	if a.Trigger > b.Trigger {
		return 1
	}
	return intraCompare(a, b)
}

// Java PriorityQueue keeps the inserted item in place on equality, and picks
// the left child when both children compare equal. Its array is also its save
// order, so an unstable sort or a heap with different tie handling is wrong.
func heapPush[T any](h []T, v T, compare func(T, T) int) []T {
	i := len(h)
	h = append(h, v)
	for i > 0 {
		p := (i - 1) >> 1
		if compare(v, h[p]) >= 0 {
			break
		}
		h[i] = h[p]
		i = p
	}
	h[i] = v
	return h
}
func heapPop[T any](h []T, compare func(T, T) int) (T, []T) {
	result := h[0]
	end := len(h) - 1
	v := h[end]
	var zero T
	h[end] = zero
	h = h[:end]
	if end > 0 {
		i := 0
		for i < end/2 {
			child := i*2 + 1
			if child+1 < end && compare(h[child], h[child+1]) > 0 {
				child++
			}
			if compare(v, h[child]) <= 0 {
				break
			}
			h[i] = h[child]
			i = child
		}
		h[i] = v
	}
	return result, h
}
func (c *Chunk[T]) add(t Scheduled[T]) {
	c.queue = heapPush(c.queue, t, drainCompare[T])
	if c.onAdded != nil {
		c.onAdded(t)
	}
}
func (c *Chunk[T]) Schedule(t Scheduled[T]) bool {
	if _, ok := c.unique[t.id()]; ok {
		return false
	}
	c.unique[t.id()] = struct{}{}
	c.add(t)
	return true
}
func (c *Chunk[T]) Has(p Position, kind T) bool { _, ok := c.unique[identity[T]{kind, p}]; return ok }
func (c *Chunk[T]) Len() int                    { return len(c.queue) + len(c.pending) }
func (c *Chunk[T]) peek() (Scheduled[T], bool) {
	if len(c.queue) == 0 {
		return Scheduled[T]{}, false
	}
	return c.queue[0], true
}
func (c *Chunk[T]) poll() Scheduled[T] {
	t, q := heapPop(c.queue, drainCompare[T])
	c.queue = q
	delete(c.unique, t.id())
	return t
}

// Unpack runs once when a saved chunk becomes active; its delays are relative
// to that activation time, not to when the bytes were read from disk.
func (c *Chunk[T]) Unpack(time int64) {
	order := -int64(len(c.pending))
	for _, t := range c.pending {
		c.add(Scheduled[T]{t.Type, t.Position, time + int64(t.Delay), t.Priority, order})
		order++
	}
	c.pending = nil
}
func (c *Chunk[T]) Save(time int64) []Saved[T] {
	result := make([]Saved[T], 0, c.Len())
	result = append(result, c.pending...)
	for _, t := range c.queue {
		result = append(result, Saved[T]{t.Type, t.Position, int32(t.Trigger - time), t.Priority})
	}
	return result
}

type Level[T comparable] struct {
	chunks    map[uint64]*Chunk[T]
	deadlines deadlineMap
	batch     []Scheduled[T]
	batchSet  map[identity[T]]struct{}
	running   bool
}

func NewLevel[T comparable]() *Level[T] {
	return &Level[T]{chunks: make(map[uint64]*Chunk[T]), deadlines: newDeadlineMap()}
}
func (l *Level[T]) Add(x, z int32, c *Chunk[T]) {
	k := ChunkKey(x, z)
	l.chunks[k] = c
	if head, ok := c.peek(); ok {
		l.deadlines.put(k, head.Trigger)
	}
	c.onAdded = func(t Scheduled[T]) {
		if head, ok := c.peek(); ok && head == t {
			l.deadlines.put(t.Position.chunk(), t.Trigger)
		}
	}
}
func (l *Level[T]) Remove(x, z int32) *Chunk[T] {
	k := ChunkKey(x, z)
	c := l.chunks[k]
	delete(l.chunks, k)
	l.deadlines.remove(k)
	if c != nil {
		c.onAdded = nil
	}
	return c
}

// Schedule returns false when the owner is unloaded or this (type,position)
// is already queued. Collected ticks have already left that uniqueness set.
func (l *Level[T]) Schedule(t Scheduled[T]) bool {
	c := l.chunks[t.Position.chunk()]
	return c != nil && c.Schedule(t)
}
func (l *Level[T]) Has(p Position, kind T) bool {
	c := l.chunks[p.chunk()]
	return c != nil && c.Has(p, kind)
}
func (l *Level[T]) WillTick(p Position, kind T) bool {
	if len(l.batchSet) == 0 && len(l.batch) > 0 {
		if l.batchSet == nil {
			l.batchSet = make(map[identity[T]]struct{})
		}
		for _, t := range l.batch {
			l.batchSet[t.id()] = struct{}{}
		}
	}
	_, ok := l.batchSet[identity[T]{kind, p}]
	return ok
}
func (l *Level[T]) Len() int {
	n := 0
	for _, c := range l.chunks {
		n += c.Len()
	}
	return n
}

// Pending returns the not-yet-executed portion of the current batch. Owners
// which checkpoint between callbacks must retain these requests too: they
// have already left the chunk heaps, but have not run yet.
func (l *Level[T]) Pending() []Scheduled[T] { return append([]Scheduled[T](nil), l.batch...) }

// Tick collects the whole bounded batch before executing any callback.
// Eligible must be a read-only predicate. Callbacks may schedule new work, but
// it cannot join this batch even if its requested trigger is already due.
func (l *Level[T]) Tick(time int64, limit int, eligible func(x, z int32) bool, run func(Scheduled[T])) {
	if l.running {
		panic("recursive scheduled-tick execution")
	}
	l.running = true
	defer func() { l.running = false; l.batch = nil; clear(l.batchSet) }()
	var ready []*Chunk[T]
	compare := func(a, b *Chunk[T]) int { return intraCompare(a.queue[0], b.queue[0]) }
	it := l.deadlines.iterator()
	for it.next() {
		k, deadline := it.entry()
		if deadline > time {
			continue
		}
		c := l.chunks[k]
		if c == nil {
			it.remove()
			continue
		}
		head, ok := c.peek()
		if !ok {
			it.remove()
			continue
		}
		if head.Trigger > time {
			it.setValue(head.Trigger)
			continue
		}
		if eligible != nil && !eligible(int32(k), int32(k>>32)) {
			continue
		}
		it.remove()
		ready = heapPush(ready, c, compare)
	}
	update := func(c *Chunk[T]) {
		if t, ok := c.peek(); ok {
			l.deadlines.put(t.Position.chunk(), t.Trigger)
		}
	}
	for len(l.batch) < limit && len(ready) > 0 {
		var c *Chunk[T]
		c, ready = heapPop(ready, compare)
		l.batch = append(l.batch, c.poll())
		for len(l.batch) < limit {
			head, ok := c.peek()
			if !ok || head.Trigger > time || len(ready) > 0 && intraCompare(head, ready[0].queue[0]) > 0 {
				break
			}
			l.batch = append(l.batch, c.poll())
		}
		if head, ok := c.peek(); ok {
			if head.Trigger <= time && len(l.batch) < limit {
				ready = heapPush(ready, c, compare)
			} else {
				update(c)
			}
		}
	}
	for _, c := range ready {
		update(c)
	}
	for len(l.batch) > 0 {
		t := l.batch[0]
		l.batch = l.batch[1:]
		delete(l.batchSet, t.id())
		run(t)
	}
}
