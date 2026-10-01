// Package player holds the per-connection player state. Each player owns a
// buffered outbound queue drained by a dedicated writer goroutine, so game
// goroutines never block on network I/O; a slow client is disconnected
// instead of stalling the world.
package player

import (
	"atlas/internal/item"
	"atlas/internal/proto"
	"math"
	"sync"
)

// OutPacket is one encoded clientbound packet awaiting delivery.
type OutPacket struct {
	ID      int32
	Payload []byte
}

// ProfileProperty comes from Mojang's authenticated GameProfile in online mode.
type ProfileProperty struct {
	Name      string
	Value     string
	Signature string
}

// Player is the server-side view of one connected client (play phase).
type Player struct {
	Name              string
	UUID              [16]byte
	EID               int32
	ProfileProperties []ProfileProperty // set before joining the hub, then immutable

	Conn         *proto.FramedConn
	Out          chan OutPacket
	Done         chan struct{} // closed when the connection dies
	ChunkUpdates chan struct{} // coalesced delivery notifications; actors never wait for entity tracking
	MovementMu   sync.Mutex    // serialize movement with console teleports
	SaveMu       sync.Mutex    // serialize autosave/logout snapshots and writes

	closeOnce sync.Once

	// mu guards the mutable game state below. The connection's own read
	// loop writes it; other goroutines read it through the accessors.
	mu              sync.Mutex
	x, y, z         float64
	yaw, pitch      float32
	onGround        bool
	flying          bool
	crouching       bool
	sprinting       bool
	playReady       bool
	teleportID      int32
	teleportPending bool
	cx, cz          int32
	viewDist        int32
	skinParts       byte
	hotbar          [9]uint32 // creative hotbar: block states for slots 36..44
	heldSlot        int32
	inventory       item.Inventory
	inventoryState  int32
	dragKind        int
	dragSlots       []int
	lastKeep        int64
	chunkSendMu     sync.Mutex            // serialize complete batches from independent region actors
	visible         map[int32]struct{}    // EIDs spawned client-side for this player
	sentChunks      map[[2]int32]struct{} // chunk packets already delivered
	requestedChunks map[[2]int32]struct{}
}

// New creates a player with sane defaults (creative mode, view distance 8).
func New(name string, uuid [16]byte, eid int32, fc *proto.FramedConn) *Player {
	return &Player{
		Name:            name,
		UUID:            uuid,
		EID:             eid,
		Conn:            fc,
		Out:             make(chan OutPacket, 8192),
		Done:            make(chan struct{}),
		ChunkUpdates:    make(chan struct{}, 1),
		viewDist:        8,
		skinParts:       0x7f,
		visible:         make(map[int32]struct{}),
		sentChunks:      make(map[[2]int32]struct{}),
		requestedChunks: make(map[[2]int32]struct{}),
	}
}

// SkinParts is the client's seven displayed-skin-layer bits (cape through
// hat). The server forwards this value in player entity metadata.
func (p *Player) SkinParts() byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.skinParts
}

func (p *Player) SetSkinParts(parts byte) bool {
	parts &= 0x7f
	p.mu.Lock()
	defer p.mu.Unlock()
	changed := p.skinParts != parts
	p.skinParts = parts
	return changed
}

// SendPacket queues a packet. Never blocks: a full queue means the client
// cannot keep up, and it is disconnected rather than allowed to stall the
// world.
func (p *Player) SendPacket(id int32, payload []byte) {
	select {
	case <-p.Done:
		return
	default:
	}
	select {
	case p.Out <- OutPacket{ID: id, Payload: payload}:
	default:
		p.Disconnect()
	}
}

// SendChunkData sends a complete one-chunk batch. Noise generation can take
// seconds between chunks, so waiting to fill a 16-chunk batch leaves a client
// waiting indefinitely for batch_finished. Serialize region actors so their
// start/data/finished packets cannot form nested batches.
func (p *Player) SendChunkData(payload []byte) {
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.SendPacket(proto.SChunkBatchStart, nil)
	p.SendPacket(proto.SChunkData, payload)
	p.SendPacket(proto.SChunkBatchFinished, proto.NewWriter().VarInt(1).Payload())
}

// SendRequestedChunk queues a reserved chunk only while it is still wanted.
// Delivery and pruning share the batch lock so an unload cannot overtake data.
func (p *Player) SendRequestedChunk(cx, cz int32, payload []byte) bool {
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	key := [2]int32{cx, cz}
	_, requested := p.requestedChunks[key]
	if !requested {
		p.mu.Unlock()
		return false
	}
	delete(p.requestedChunks, key)
	p.SendPacket(proto.SChunkBatchStart, nil)
	p.SendPacket(proto.SChunkData, payload)
	p.SendPacket(proto.SChunkBatchFinished, proto.NewWriter().VarInt(1).Payload())
	p.sentChunks[key] = struct{}{}
	p.mu.Unlock()
	select {
	case p.ChunkUpdates <- struct{}{}:
	default:
	}
	return true
}

func (p *Player) ReserveChunk(cx, cz int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	k := [2]int32{cx, cz}
	if _, ok := p.sentChunks[k]; ok {
		return false
	}
	if _, ok := p.requestedChunks[k]; ok {
		return false
	}
	p.requestedChunks[k] = struct{}{}
	return true
}

func (p *Player) WantsChunk(cx, cz int32) bool {
	select {
	case <-p.Done:
		return false
	default:
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	_, wanted := p.requestedChunks[[2]int32{cx, cz}]
	return wanted
}

// CancelPendingChunks stops old view requests before a teleport's blocking
// destination preparation. Already delivered ground stays visible until the
// destination is ready and the client cache is re-centered.
func (p *Player) CancelPendingChunks() {
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	clear(p.requestedChunks)
	p.mu.Unlock()
}

func (p *Player) SetPlayReady(v bool) { p.mu.Lock(); p.playReady = v; p.mu.Unlock() }
func (p *Player) PlayReady() bool     { p.mu.Lock(); defer p.mu.Unlock(); return p.playReady }
func (p *Player) BeginTeleport(id int32) {
	p.mu.Lock()
	p.teleportID = id
	p.teleportPending = true
	p.mu.Unlock()
}
func (p *Player) ConfirmTeleport(id int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.teleportPending || p.teleportID != id {
		return false
	}
	p.teleportPending = false
	return true
}
func (p *Player) TeleportPending() bool { p.mu.Lock(); defer p.mu.Unlock(); return p.teleportPending }
func (p *Player) SetSprinting(v bool)   { p.mu.Lock(); p.sprinting = v; p.mu.Unlock() }
func (p *Player) Sprinting() bool       { p.mu.Lock(); defer p.mu.Unlock(); return p.sprinting }

// WriterLoop drains the outbound queue. It is the only writer to the
// connection after handoff, so packet writes need no locking.
func (p *Player) WriterLoop() {
	for {
		select {
		case pm := <-p.Out:
			if err := p.Conn.WritePacket(pm.ID, pm.Payload); err != nil {
				p.Disconnect()
				return
			}
		case <-p.Done:
			return
		}
	}
}

// Disconnect tears the connection down. Safe from any goroutine.
func (p *Player) Disconnect() {
	p.closeOnce.Do(func() {
		close(p.Done)
		if p.Conn != nil {
			p.Conn.Conn.Close()
		}
	})
}

// --- position & orientation ---

// SetPos updates position and orientation; reports whether the player
// crossed into a new chunk.
func (p *Player) SetPos(x, y, z float64, yaw, pitch float32, ground bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.x, p.y, p.z = x, y, z
	p.yaw, p.pitch = yaw, pitch
	p.onGround = ground
	ncx, ncz := int32(math.Floor(x/16)), int32(math.Floor(z/16))
	changed := ncx != p.cx || ncz != p.cz
	p.cx, p.cz = ncx, ncz
	return changed
}

// Pos returns a snapshot of the player's position and orientation.
func (p *Player) Pos() (x, y, z float64, yaw, pitch float32, ground bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.x, p.y, p.z, p.yaw, p.pitch, p.onGround
}

func (p *Player) SetFlying(v bool) {
	p.mu.Lock()
	p.flying = v
	p.mu.Unlock()
}

func (p *Player) Flying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.flying
}

func (p *Player) SetCrouching(v bool) { p.mu.Lock(); p.crouching = v; p.mu.Unlock() }
func (p *Player) Crouching() bool     { p.mu.Lock(); defer p.mu.Unlock(); return p.crouching }

// ChunkPos returns the player's current chunk coordinates.
func (p *Player) ChunkPos() (cx, cz int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cx, p.cz
}

// --- view distance ---

func (p *Player) ViewDistance() int32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.viewDist
}

func (p *Player) SetViewDistance(v int32) {
	if v < 2 {
		v = 2
	}
	if v > 16 {
		// Server-side cap: 32 would flood ~3400 chunk payloads (~200MB) on
		// join. 16 (~900 chunks) is already generous for this build.
		v = 16
	}
	p.mu.Lock()
	p.viewDist = v
	p.mu.Unlock()
}

// --- creative hotbar ---

// SetHotbarSlot records a creative-inventory update for slots 36..44.
func (p *Player) SetHotbarSlot(slot int32, state uint32) {
	if slot < 36 || slot > 44 {
		return
	}
	p.mu.Lock()
	p.hotbar[slot-36] = state
	p.mu.Unlock()
}

// SetHeld records the selected hotbar slot (0..8 from CHeldItemSlot).
func (p *Player) SetHeld(slot int32) {
	if slot < 0 || slot > 8 {
		return
	}
	p.mu.Lock()
	p.heldSlot = slot
	p.inventory.Selected = slot
	p.mu.Unlock()
}

// HeldBlock returns the block state in the selected hotbar slot.
func (p *Player) HeldBlock() uint32 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hotbar[p.heldSlot]
}

// --- keepalive bookkeeping ---

func (p *Player) SetKeepAliveID(id int64) {
	p.mu.Lock()
	p.lastKeep = id
	p.mu.Unlock()
}

func (p *Player) KeepAliveID() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastKeep
}

// --- entity visibility (per-player client-side entity list) ---

func (p *Player) MarkVisible(eid int32) {
	p.mu.Lock()
	p.visible[eid] = struct{}{}
	p.mu.Unlock()
}

func (p *Player) MarkInvisible(eid int32) {
	p.mu.Lock()
	delete(p.visible, eid)
	p.mu.Unlock()
}

func (p *Player) IsVisible(eid int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.visible[eid]
	return ok
}

// --- sent chunk tracking ---

func (p *Player) MarkChunkSent(cx, cz int32) {
	p.mu.Lock()
	p.sentChunks[[2]int32{cx, cz}] = struct{}{}
	p.mu.Unlock()
}

func (p *Player) IsChunkSent(cx, cz int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.sentChunks[[2]int32{cx, cz}]
	return ok
}

func (p *Player) UnmarkChunkSent(cx, cz int32) {
	p.mu.Lock()
	delete(p.sentChunks, [2]int32{cx, cz})
	delete(p.requestedChunks, [2]int32{cx, cz})
	p.mu.Unlock()
}

// ForEachSentChunk calls fn for every chunk delivered to this player.
// Returning false from fn stops the iteration.
func (p *Player) ForEachSentChunk(fn func(cx, cz int32) bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k := range p.sentChunks {
		if !fn(k[0], k[1]) {
			return
		}
	}
}

// PruneChunks removes every sent chunk farther than maxD2 (squared chunk
// distance) from (curCx, curCz) and returns the removed coordinates.
func (p *Player) PruneChunks(curCx, curCz, maxD2 int32) [][2]int32 {
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	defer p.mu.Unlock()
	var pruned [][2]int32
	for k := range p.sentChunks {
		dx, dz := int64(k[0])-int64(curCx), int64(k[1])-int64(curCz)
		if dx*dx+dz*dz > int64(maxD2) {
			delete(p.sentChunks, k)
			pruned = append(pruned, k)
			p.SendPacket(proto.SUnloadChunk, proto.NewWriter().I32(k[1]).I32(k[0]).Payload())
		}
	}
	for k := range p.requestedChunks {
		dx, dz := int64(k[0])-int64(curCx), int64(k[1])-int64(curCz)
		if dx*dx+dz*dz > int64(maxD2) {
			delete(p.requestedChunks, k)
		}
	}
	return pruned
}
