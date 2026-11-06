package player

import (
	"atlas/internal/proto"
	"math"
)

func (p *Player) Dimension() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.dimension == "" {
		return "minecraft:overworld"
	}
	return p.dimension
}

func (p *Player) SetInitialDimension(name string) {
	p.mu.Lock()
	p.dimension = name
	p.mu.Unlock()
}

func (p *Player) WorldPosition() (string, float64, float64, float64, float32, float32, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	name := p.dimension
	if name == "" {
		name = "minecraft:overworld"
	}
	return name, p.x, p.y, p.z, p.yaw, p.pitch, p.flying
}

func (p *Player) WorldPositionWithPortalCooldown() (string, float64, float64, float64, float32, float32, bool, int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	name := p.dimension
	if name == "" {
		name = "minecraft:overworld"
	}
	return name, p.x, p.y, p.z, p.yaw, p.pitch, p.flying, p.portalCooldown
}

// ResetWorld atomically moves the server position and discards the old client
// world. Epoch-scoped deliveries cannot send an earlier world's chunk/light
// packet after this respawn, even if both worlds use the same chunk coords.
func (p *Player) ResetWorld(name string, pos [3]float64, yaw, pitch float32, respawn []byte) {
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.worldEpoch++
	p.portalDuration = 0
	p.dimension = name
	p.x, p.y, p.z = pos[0], pos[1], pos[2]
	p.yaw, p.pitch = yaw, pitch
	p.onGround = false
	p.cx, p.cz = int32(math.Floor(pos[0]/16)), int32(math.Floor(pos[2]/16))
	clear(p.sentChunks)
	clear(p.requestedChunks)
	clear(p.visible)
	p.SendPacket(proto.SRespawn, respawn)
}

// WorldPacketSink captures one dimension view. WorldSink is both a broadcast
// adapter and a complete world.ChunkReceiver, without importing world here.
func (p *Player) WorldPacketSink() *WorldSink {
	p.mu.Lock()
	defer p.mu.Unlock()
	return &WorldSink{player: p, epoch: p.worldEpoch}
}

func (p *Player) WorldPacketSinkInDimension(name string) *WorldSink {
	p.mu.Lock()
	defer p.mu.Unlock()
	current := p.dimension
	if current == "" {
		current = "minecraft:overworld"
	}
	if current != name {
		return nil
	}
	return &WorldSink{player: p, epoch: p.worldEpoch}
}

type WorldSink struct {
	player *Player
	epoch  uint64
}

func (sink *WorldSink) Current() bool {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	return sink.currentLocked()
}

// currentLocked requires player.mu. Position and request queries use the same
// lock so validity, reservation and distance describe one consistent view.
func (sink *WorldSink) currentLocked() bool {
	select {
	case <-sink.player.Done:
		return false
	default:
		return sink.player.worldEpoch == sink.epoch
	}
}

func (sink *WorldSink) ReserveChunk(cx, cz int32) bool {
	if sink == nil {
		return false
	}
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	key := [2]int32{cx, cz}
	if !sink.currentLocked() {
		return false
	}
	if _, exists := p.sentChunks[key]; exists {
		return false
	}
	if _, exists := p.requestedChunks[key]; exists {
		return false
	}
	p.requestedChunks[key] = struct{}{}
	return true
}

func (sink *WorldSink) SendPacket(id int32, payload []byte) {
	p := sink.player
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	defer p.mu.Unlock()
	if sink.currentLocked() {
		p.SendPacket(id, payload)
	}
}

func (sink *WorldSink) SendChunkData(payload []byte) {
	p := sink.player
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	defer p.mu.Unlock()
	if !sink.currentLocked() {
		return
	}
	p.sendChunkBatch(payload)
}

func (sink *WorldSink) SendRequestedChunk(cx, cz int32, payload []byte) bool {
	p := sink.player
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	key := [2]int32{cx, cz}
	if _, wanted := p.requestedChunks[key]; !wanted || !sink.currentLocked() {
		if p.worldEpoch == sink.epoch {
			delete(p.requestedChunks, key)
		}
		p.mu.Unlock()
		return false
	}
	delete(p.requestedChunks, key)
	p.sendChunkBatch(payload)
	if !sink.currentLocked() {
		// A saturated outbound queue disconnects the player. A partial batch
		// on that dead connection must not count as delivered terrain.
		p.mu.Unlock()
		return false
	}
	p.sentChunks[key] = struct{}{}
	p.mu.Unlock()
	select {
	case p.ChunkUpdates <- struct{}{}:
	default:
	}
	return true
}

func (sink *WorldSink) IsChunkSent(cx, cz int32) bool {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.sentChunks[[2]int32{cx, cz}]
	return ok && p.worldEpoch == sink.epoch
}
func (sink *WorldSink) WantsChunk(cx, cz int32) bool {
	_, wanted := sink.ChunkPriority(cx, cz)
	return wanted
}

// ChunkPriority combines reservation validity and live position in one read.
// Old views and disconnected players cannot keep optional generation alive.
func (sink *WorldSink) ChunkPriority(cx, cz int32) (int64, bool) {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.requestedChunks[[2]int32{cx, cz}]
	if !ok || !sink.currentLocked() {
		return 0, false
	}
	dx, dz := int64(cx)-int64(p.cx), int64(cz)-int64(p.cz)
	return dx*dx + dz*dz, true
}

// ReleaseChunk undoes a failed reservation without erasing delivered terrain.
// Failure from an old epoch cannot release the new view's same-coordinate work.
func (sink *WorldSink) ReleaseChunk(cx, cz int32) {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.worldEpoch == sink.epoch {
		delete(p.requestedChunks, [2]int32{cx, cz})
	}
}
