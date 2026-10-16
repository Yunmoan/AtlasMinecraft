package world

import (
	"atlas/internal/light"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// PacketSink is anything that can accept an encoded clientbound packet.
// *player.Player implements it; defined here to avoid an import cycle.
type PacketSink interface {
	SendPacket(id int32, payload []byte)
	// SendChunkData delivers one chunk-data payload, wrapped in the
	// client's chunk-batch framing by the sink.
	SendChunkData(payload []byte)
}

func deliverChunk(to PacketSink, cx, cz int32, payload []byte) bool {
	if receiver, ok := to.(interface {
		SendRequestedChunk(int32, int32, []byte) bool
	}); ok {
		return receiver.SendRequestedChunk(cx, cz, payload)
	}
	to.SendChunkData(payload)
	return true
}

func chunkDelivered(to PacketSink, cx, cz int32) bool {
	if receiver, ok := to.(interface{ IsChunkSent(int32, int32) bool }); ok {
		return receiver.IsChunkSent(cx, cz)
	}
	return true
}

// HubView lets a region find the players who can see a block position.
type HubView interface {
	ViewersOf(x, z int32) []PacketSink
}

const (
	RegionChunkBits = 5
	RegionChunks    = 1 << RegionChunkBits // 32x32 chunks owned by each region actor
)

// chunkKey identifies a chunk column.
type chunkKey struct{ x, z int32 }

// Region is an actor goroutine owning a 32x32-chunk slice of the world.
// All chunk state is touched only by the region's goroutine — no locks —
// and cross-region interaction happens purely through the Hub's viewer
// lookup, which is coordinate-based and lock-free per message.
type Region struct {
	world *World

	chunks          map[chunkKey]*Chunk
	packets         chunkPacketCache
	pending         map[chunkKey][]PacketSink // sinks waiting for a chunk to generate
	pendingLoads    map[chunkKey][]chan chunkResult
	pendingReady    map[chunkKey][]chan error
	pendingSurfaces map[chunkKey][]surfaceQuery
	genQueued       map[chunkKey]bool
	unavailable     map[chunkKey]error
	pendingEdits    map[chunkKey][]setBlockMsg

	mailbox chan any // chunkRequest | chunkGenerated | setBlockMsg
}

type chunkRequest struct {
	cx, cz int32
	to     PacketSink
}

type chunkGenerated struct {
	cx, cz int32
	chunk  *Chunk
	err    error
}

type generationCanceled struct{ cx, cz int32 }

type chunkResult struct {
	payload []byte
	err     error
}
type surfaceResult struct {
	y   int32
	err error
}

type chunkLoad struct {
	cx, cz int32
	result chan chunkResult
}

type chunkEnsure struct {
	cx, cz int32
	result chan error
}

type surfaceQuery struct {
	x, z   int32
	result chan surfaceResult
}

type spawnScanResult struct {
	x, y, z int32
	ok      bool
}

// spawnScanQuery asks the owning actor to run the vanilla spawn column scan
// over one loaded chunk. Unloaded chunks report ok=false without generating.
type spawnScanQuery struct {
	cx, cz int32
	result chan spawnScanResult
}

type saveRegion struct{ result chan error }

type setBlockMsg struct {
	x, y, z       int32
	state         uint32
	expected      *uint32
	upper         *uint32
	upperExpected *uint32
	changed       chan []BlockPosition
	result        chan error
}

func newRegion(w *World, rx, rz int32) *Region {
	return &Region{
		world:           w,
		chunks:          make(map[chunkKey]*Chunk),
		pending:         make(map[chunkKey][]PacketSink),
		pendingLoads:    make(map[chunkKey][]chan chunkResult),
		pendingReady:    make(map[chunkKey][]chan error),
		pendingSurfaces: make(map[chunkKey][]surfaceQuery),
		genQueued:       make(map[chunkKey]bool),
		unavailable:     make(map[chunkKey]error),
		pendingEdits:    make(map[chunkKey][]setBlockMsg),
		mailbox:         make(chan any, 4096),
	}
}

// run is the region actor's event loop.
func (r *Region) run() {
	defer r.world.wg.Done()
	for {
		var m any
		select {
		case <-r.world.done:
			return
		case m = <-r.mailbox:
		}
		select {
		case <-r.world.done:
			return
		default:
		}
		switch msg := m.(type) {
		case containerQuery:
			r.containerQuery(msg)
		case entityQuery:
			r.entitiesInChunk(msg)
		case armorStandQuery:
			r.interactArmorStand(msg)
		case generationCanceled:
			k := chunkKey{msg.cx, msg.cz}
			delete(r.genQueued, k)
			var wanted []PacketSink
			for _, sink := range r.pending[k] {
				if chunkWanted(sink, k.x, k.z) {
					wanted = append(wanted, sink)
				}
			}
			r.pending[k] = wanted
			if len(wanted) > 0 || len(r.pendingLoads[k])+len(r.pendingReady[k])+len(r.pendingEdits[k])+len(r.pendingSurfaces[k]) > 0 {
				r.queueGeneration(k)
			} else {
				delete(r.pending, k)
			}
		case lightChanged:
			if ch := r.chunks[chunkKey{msg.cx, msg.cz}]; ch != nil {
				// A queued notification may belong to an earlier load of this
				// coordinate. Only publish the engine's current snapshot.
				data := msg.data
				if r.world.lighting != nil {
					data = r.world.lighting.Snapshot(light.Position{X: msg.cx, Z: msg.cz})
				}
				r.applyLight(ch, data, true)
			}
		case unloadRegion:
			count, err := r.unloadUnused(msg.cutoff)
			msg.result <- unloadResult{count, err}
		case blockRead:
			states := make([]uint32, len(msg.Positions))
			var err error
			for i, p := range msg.Positions {
				ch := r.chunks[chunkKey{p.X >> 4, p.Z >> 4}]
				if ch == nil {
					err = ErrChunkNotLoaded
					break
				}
				states[i] = ch.GetBlock(int(p.X&15), int(p.Z&15), p.Y)
				ch.lastAccess = time.Now()
			}
			msg.Result <- blockReadResult{States: states, Err: err}
		case portalChunkQuery:
			msg.result <- r.portalChunkSnapshot(msg)
		case saveRegion:
			msg.result <- r.saveChunks()
		case chunkEnsure:
			k := chunkKey{msg.cx, msg.cz}
			if err := r.unavailable[k]; err != nil {
				msg.result <- err
				continue
			}
			if r.chunks[k] != nil {
				r.chunks[k].lastAccess = time.Now()
				msg.result <- nil
				continue
			}
			r.pendingReady[k] = append(r.pendingReady[k], msg.result)
			r.queueGeneration(k)
		case chunkRequest:
			if !chunkWanted(msg.to, msg.cx, msg.cz) {
				continue
			}
			k := chunkKey{msg.cx, msg.cz}
			if r.unavailable[k] != nil {
				if pending, ok := msg.to.(interface{ UnmarkChunkSent(int32, int32) }); ok {
					pending.UnmarkChunkSent(msg.cx, msg.cz)
				}
				continue
			}
			if ch := r.chunks[k]; ch != nil {
				ch.lastAccess = time.Now()
				deliverChunk(msg.to, msg.cx, msg.cz, r.packets.packet(ch))
				continue
			}
			r.pending[k] = append(r.pending[k], msg.to)
			r.queueGeneration(k)

		case chunkLoad:
			k := chunkKey{msg.cx, msg.cz}
			if err := r.unavailable[k]; err != nil {
				msg.result <- chunkResult{err: err}
				continue
			}
			if ch := r.chunks[k]; ch != nil {
				ch.lastAccess = time.Now()
				msg.result <- chunkResult{payload: r.packets.packet(ch)}
				continue
			}
			r.pendingLoads[k] = append(r.pendingLoads[k], msg.result)
			r.queueGeneration(k)

		case surfaceQuery:
			k := chunkKey{msg.x >> 4, msg.z >> 4}
			if err := r.unavailable[k]; err != nil {
				msg.result <- surfaceResult{err: err}
				continue
			}
			if ch := r.chunks[k]; ch != nil {
				ch.lastAccess = time.Now()
				msg.result <- surfaceResult{y: ch.Height(int(msg.x&15), int(msg.z&15))}
				continue
			}
			r.pendingSurfaces[k] = append(r.pendingSurfaces[k], msg)
			r.queueGeneration(k)

		case spawnScanQuery:
			k := chunkKey{msg.cx, msg.cz}
			if ch := r.chunks[k]; ch != nil {
				ch.lastAccess = time.Now()
				x, y, z, ok := findSafeSpawnColumn(ch)
				msg.result <- spawnScanResult{x: x, y: y, z: z, ok: ok}
				continue
			}
			msg.result <- spawnScanResult{}

		case chunkGenerated:
			k := chunkKey{msg.cx, msg.cz}
			if msg.err != nil {
				if errors.Is(msg.err, ErrVanillaGenerationUnavailable) {
					r.unavailable[k] = msg.err
					slog.Warn("vanilla chunk cannot be generated", "x", msg.cx, "z", msg.cz, "error", msg.err)
				} else {
					slog.Warn("chunk unavailable", "x", msg.cx, "z", msg.cz, "error", msg.err)
				}
				for _, result := range r.pendingLoads[k] {
					result <- chunkResult{err: msg.err}
				}
				for _, result := range r.pendingReady[k] {
					result <- msg.err
				}
				for _, query := range r.pendingSurfaces[k] {
					query.result <- surfaceResult{err: msg.err}
				}
				for _, edit := range r.pendingEdits[k] {
					edit.result <- msg.err
				}
				for _, sink := range r.pending[k] {
					if pending, ok := sink.(interface{ UnmarkChunkSent(int32, int32) }); ok {
						pending.UnmarkChunkSent(msg.cx, msg.cz)
					}
				}
				delete(r.pending, k)
				delete(r.pendingLoads, k)
				delete(r.pendingReady, k)
				delete(r.pendingSurfaces, k)
				delete(r.pendingEdits, k)
				delete(r.genQueued, k)
				continue
			}
			if r.chunks[k] == nil {
				r.world.loadedChunks.Add(1)
			}
			r.chunks[k] = msg.chunk
			msg.chunk.lastAccess = time.Now()
			if r.world.lighting != nil {
				r.applyLight(msg.chunk, r.world.lighting.Snapshot(light.Position{X: msg.cx, Z: msg.cz}), false)
			}
			delete(r.unavailable, k)
			delete(r.genQueued, k)
			for _, edit := range r.pendingEdits[k] {
				edit.result <- r.applyBlockEdit(msg.chunk, edit)
			}
			delete(r.pendingEdits, k)
			var payload []byte
			if len(r.pending[k]) > 0 || len(r.pendingLoads[k]) > 0 {
				payload = r.packets.packet(msg.chunk)
			}
			for _, result := range r.pendingReady[k] {
				result <- nil
			}
			for _, sink := range r.pending[k] {
				deliverChunk(sink, msg.cx, msg.cz, payload)
			}
			for _, result := range r.pendingLoads[k] {
				result <- chunkResult{payload: payload}
			}
			for _, query := range r.pendingSurfaces[k] {
				query.result <- surfaceResult{y: msg.chunk.Height(int(query.x&15), int(query.z&15))}
			}
			delete(r.pending, k)
			delete(r.pendingLoads, k)
			delete(r.pendingReady, k)
			delete(r.pendingSurfaces, k)

		case setBlockMsg:
			k := chunkKey{msg.x >> 4, msg.z >> 4}
			if err := r.unavailable[k]; err != nil {
				msg.result <- err
				continue
			}
			ch := r.chunks[k]
			if ch == nil {
				r.pendingEdits[k] = append(r.pendingEdits[k], msg)
				r.queueGeneration(k)
				continue
			}
			msg.result <- r.applyBlockEdit(ch, msg)
			ch.lastAccess = time.Now()
		}
	}
}

func (r *Region) saveChunks() error {
	var combined error
	if r.world.store != nil {
		for key, ch := range r.chunks {
			r.world.store.vanilla.refreshStrongholdFlags(ch)
			if r.world.lighting != nil {
				r.applyLight(ch, r.world.lighting.Snapshot(light.Position{X: key.x, Z: key.z}), false)
			}
			if ch.dirty {
				if err := r.world.store.vanilla.persistChunk(ch); err != nil {
					combined = errors.Join(combined, fmt.Errorf("chunk %d,%d: %w", key.x, key.z, err))
				}
			}
		}
	}
	return combined
}

func (r *Region) applyBlockEdit(ch *Chunk, msg setBlockMsg) error {
	return r.applyColumnEdit(ch, msg)
}

func (r *Region) queueGeneration(k chunkKey) {
	create := !r.genQueued[k]
	r.genQueued[k] = true
	job := genJob{cx: k.x, cz: k.z, region: r}
	required := len(r.pendingLoads[k])+len(r.pendingReady[k])+len(r.pendingEdits[k])+len(r.pendingSurfaces[k]) > 0
	r.world.generation.enqueue(job, r.pending[k], required, create)
}
