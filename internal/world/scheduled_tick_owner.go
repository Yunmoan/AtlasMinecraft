package world

import (
	"atlas/internal/nbt"
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// scheduledTickOwner owns both scheduled block and fluid queues. Their levels
// are separate but share the world's sub-tick counter and activation lifetime.
// Its lock is only held for in-memory queue operations; tick callbacks and
// actor requests run without it. World.blockUpdateMu serializes a complete
// tick pass against edits, saves and unloads. Actors never acquire that lock.
type scheduledTickOwner struct {
	mu           sync.Mutex
	level        *tickqueue.Level[worldgen.FluidType]
	blocks       *tickqueue.Level[string]
	chunks       map[uint64]*chunkScheduledTicks
	loadOrder    []*chunkScheduledTicks
	now, order   int64
	current      *tickqueue.Scheduled[worldgen.FluidType]
	currentBlock *tickqueue.Scheduled[string]
	// Optional read-only diagnostic sink. Install under blockUpdateMu; callbacks
	// run without owner.mu and must not edit the world or invoke ticking again.
	observeDispatch func(time int64, kind string, position tickqueue.Position, name string)
}

// This handle is immutable except under owner.mu. The region owns its Chunk,
// while this handle lets persistence request a detached queue snapshot.
type chunkScheduledTicks struct {
	owner                   *scheduledTickOwner
	queue                   *tickqueue.Chunk[worldgen.FluidType]
	blocks                  *tickqueue.Chunk[string]
	x, z                    int32
	active                  bool
	post                    []worldgen.FeaturePosition
	postIndex               int
	postSections            int
	revision, savedRevision uint64
	savedTime               int64
}

func newScheduledTickOwner(time int64) *scheduledTickOwner {
	return &scheduledTickOwner{level: tickqueue.NewLevel[worldgen.FluidType](), blocks: tickqueue.NewLevel[string](), chunks: make(map[uint64]*chunkScheduledTicks), now: time}
}

// ChunkSerializer uses the optional registry lookup, including air. It accepts
// absent/default namespaces but never substitutes air for an unknown block.
func parseScheduledBlock(name string) (string, bool) {
	if strings.HasPrefix(name, ":") {
		name = "minecraft" + name
	} else if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	_, err := BlockStateID(name, nil)
	return name, err == nil
}

func parseScheduledFluid(name string) (worldgen.FluidType, bool) {
	for kind := worldgen.FluidEmpty; kind <= worldgen.FluidLava; kind++ {
		canonical := kind.String()
		short := canonical[len("minecraft:"):]
		if name == canonical || name == short || name == ":"+short {
			return kind, true
		}
	}
	return worldgen.FluidEmpty, false
}

func (o *scheduledTickOwner) register(c *Chunk) *chunkScheduledTicks {
	root, _ := c.anvilRoot.Object()
	h := &chunkScheduledTicks{owner: o, x: c.cx, z: c.cz, queue: tickqueue.NewChunk(tickqueue.DecodeSavedNBT(root["fluid_ticks"], c.cx, c.cz, parseScheduledFluid))}
	h.blocks = tickqueue.NewChunk(tickqueue.DecodeSavedNBT(root["block_ticks"], c.cx, c.cz, parseScheduledBlock))
	sections, _ := root["PostProcessing"].Elements()
	h.postSections = len(sections)
	for section, list := range sections {
		positions, _ := list.Elements()
		for _, tag := range positions {
			packed, ok := tag.Number()
			if !ok {
				continue
			}
			h.post = append(h.post, worldgen.FeaturePosition{X: c.cx*16 + int32(packed&15), Y: c.minY + int32(section)*16 + int32((packed>>4)&15), Z: c.cz*16 + int32((packed>>8)&15)})
		}
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	k := tickqueue.ChunkKey(c.cx, c.cz)
	if previous := o.chunks[k]; previous != nil {
		panic("duplicate loaded scheduled tick owner")
	}
	o.chunks[k] = h
	o.loadOrder = append(o.loadOrder, h)
	o.level.Add(c.cx, c.cz, h.queue)
	o.blocks.Add(c.cx, c.cz, h.blocks)
	return h
}

func (h *chunkScheduledTicks) snapshot() (nbt.Tag, nbt.Tag, nbt.Tag, uint64, int64) {
	o := h.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	// Existing actor operations persist synchronously. If they save while a
	// tick callback is running, retain both that callback and the rest of the
	// collected batch. At normal tick boundaries this list is empty, leaving
	// precisely the native heap save order. A crash can replay a current tick,
	// but must not silently lose every unexecuted request in the batch.
	saved := snapshotScheduledQueue(o.level, h.queue, o.current, h.x, h.z, o.now)
	blocks := snapshotScheduledQueue(o.blocks, h.blocks, o.currentBlock, h.x, h.z, o.now)
	var post nbt.Tag
	if h.active {
		sections := make([]nbt.Tag, h.postSections)
		for i := range sections {
			sections[i] = nbt.Tag{Type: nbt.List, ListType: nbt.End, Value: []nbt.Tag{}}
		}
		elementType := byte(nbt.List)
		if len(sections) == 0 {
			elementType = nbt.End
		}
		post = nbt.Tag{Type: nbt.List, ListType: elementType, Value: sections}
	}
	return tickqueue.EncodeSavedNBT(saved, func(kind worldgen.FluidType) string { return kind.String() }), tickqueue.EncodeSavedNBT(blocks, func(name string) string { return name }), post, h.revision, o.now
}

// Called with the owner's lock held; all returned entries are detached.
func snapshotScheduledQueue[T comparable](level *tickqueue.Level[T], queue *tickqueue.Chunk[T], current *tickqueue.Scheduled[T], x, z int32, now int64) []tickqueue.Saved[T] {
	var saved []tickqueue.Saved[T]
	appendPending := func(t tickqueue.Scheduled[T]) {
		if t.Position.X>>4 == x && t.Position.Z>>4 == z {
			saved = append(saved, tickqueue.Saved[T]{Type: t.Type, Position: t.Position, Delay: int32(t.Trigger - now), Priority: t.Priority})
		}
	}
	if current != nil {
		appendPending(*current)
	}
	for _, t := range level.Pending() {
		appendPending(t)
	}
	return append(saved, queue.Save(now)...)
}

func (h *chunkScheduledTicks) needsSave() bool {
	o := h.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	return h.revision != h.savedRevision || h.active && (h.queue.Len() > 0 || h.blocks.Len() > 0) && h.savedTime != o.now
}
func (h *chunkScheduledTicks) saved(revision uint64, time int64) {
	o := h.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	h.savedRevision = revision
	h.savedTime = time
}

func (h *chunkScheduledTicks) unregister() {
	o := h.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	k := tickqueue.ChunkKey(h.x, h.z)
	if o.chunks[k] != h {
		panic("stale scheduled tick owner")
	}
	o.level.Remove(h.x, h.z)
	o.blocks.Remove(h.x, h.z)
	delete(o.chunks, k)
	for i, entry := range o.loadOrder {
		if entry == h {
			o.loadOrder = append(o.loadOrder[:i], o.loadOrder[i+1:]...)
			break
		}
	}
}

func (o *scheduledTickOwner) setTime(time int64) { o.mu.Lock(); o.now = time; o.mu.Unlock() }

func (o *scheduledTickOwner) schedule(p worldgen.FeaturePosition, kind worldgen.FluidType, delay int32) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.chunks[tickqueue.ChunkKey(p.X>>4, p.Z>>4)] == nil {
		return ErrChunkNotLoaded
	}
	if o.level.Schedule(tickqueue.Scheduled[worldgen.FluidType]{Type: kind, Position: tickqueue.Position{X: p.X, Y: p.Y, Z: p.Z}, Trigger: o.now + int64(delay), Order: o.order}) {
		o.chunks[tickqueue.ChunkKey(p.X>>4, p.Z>>4)].revision++
	}
	o.order++
	return nil
}

func (o *scheduledTickOwner) scheduleBlock(p worldgen.FeaturePosition, name string, delay int32, priority tickqueue.Priority) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	h := o.chunks[tickqueue.ChunkKey(p.X>>4, p.Z>>4)]
	if h == nil {
		return ErrChunkNotLoaded
	}
	if o.blocks.Schedule(tickqueue.Scheduled[string]{Type: name, Position: tickqueue.Position{X: p.X, Y: p.Y, Z: p.Z}, Trigger: o.now + int64(delay), Priority: priority, Order: o.order}) {
		h.revision++
	}
	o.order++
	return nil
}

func (o *scheduledTickOwner) ready(h *chunkScheduledTicks) bool {
	for x := h.x - 1; x <= h.x+1; x++ {
		for z := h.z - 1; z <= h.z+1; z++ {
			if o.chunks[tickqueue.ChunkKey(x, z)] == nil {
				return false
			}
		}
	}
	return true
}

// activate follows ChunkMap.prepareTickingChunk: a full 3x3 neighborhood,
// postProcessGeneration, then ServerLevel.startTickingChunk/unpackTicks.
// The caller holds blockUpdateMu; callbacks may safely use region messages.
func (o *scheduledTickOwner) activate(eligible func(int32, int32) bool, process func(worldgen.FeaturePosition) error) error {
	o.mu.Lock()
	var pending []*chunkScheduledTicks
	for _, h := range o.loadOrder {
		if !h.active && eligible(h.x, h.z) && o.ready(h) {
			pending = append(pending, h)
		}
	}
	o.mu.Unlock()
	for _, h := range pending {
		for h.postIndex < len(h.post) {
			if err := process(h.post[h.postIndex]); err != nil {
				return fmt.Errorf("activate scheduled chunk %d,%d: %w", h.x, h.z, err)
			}
			// postIndex is not read by actors/snapshot; only the serialized
			// activation pass owns it until the active bit is published.
			h.postIndex++
		}
		o.mu.Lock()
		h.blocks.Unpack(o.now)
		h.queue.Unpack(o.now)
		h.active = true
		h.revision++
		o.mu.Unlock()
	}
	return nil
}

func (o *scheduledTickOwner) tick(limit int, eligible func(int32, int32) bool, run func(tickqueue.Scheduled[worldgen.FluidType]) error) error {
	return tickOwnedQueue(o, o.level, &o.current, "fluid", limit, eligible, run)
}

func (o *scheduledTickOwner) tickBlocks(limit int, eligible func(int32, int32) bool, run func(tickqueue.Scheduled[string]) error) error {
	return tickOwnedQueue(o, o.blocks, &o.currentBlock, "block", limit, eligible, run)
}

func tickOwnedQueue[T comparable](o *scheduledTickOwner, level *tickqueue.Level[T], current **tickqueue.Scheduled[T], kind string, limit int, eligible func(int32, int32) bool, run func(tickqueue.Scheduled[T]) error) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	var failure error
	level.Tick(o.now, limit, func(x, z int32) bool {
		h := o.chunks[tickqueue.ChunkKey(x, z)]
		return h != nil && h.active && o.ready(h) && eligible(x, z)
	}, func(t tickqueue.Scheduled[T]) {
		if failure != nil {
			level.Schedule(t)
			o.chunks[tickqueue.ChunkKey(t.Position.X>>4, t.Position.Z>>4)].revision++
			return
		}
		h := o.chunks[tickqueue.ChunkKey(t.Position.X>>4, t.Position.Z>>4)]
		h.revision++
		*current = &t
		var err error
		observe, time := o.observeDispatch, o.now
		func() {
			o.mu.Unlock()
			defer o.mu.Lock()
			if observe != nil {
				observe(time, kind, t.Position, fmt.Sprint(t.Type))
			}
			err = run(t)
		}()
		*current = nil
		h.revision++
		if err != nil {
			level.Schedule(t)
			failure = errors.Join(failure, err)
		}
	})
	return failure
}
