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

// WorldPacketSink captures one world view. Its methods deliberately match
// world.PacketSink and the region's optional delivery interfaces without an
// import from this package back into world.
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
	return p.worldEpoch == sink.epoch
}

func (sink *WorldSink) ReserveChunk(cx, cz int32) bool {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	key := [2]int32{cx, cz}
	if p.worldEpoch != sink.epoch {
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
	if p.worldEpoch == sink.epoch {
		p.SendPacket(id, payload)
	}
}

func (sink *WorldSink) SendChunkData(payload []byte) {
	p := sink.player
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.worldEpoch != sink.epoch {
		return
	}
	p.SendPacket(proto.SChunkBatchStart, nil)
	p.SendPacket(proto.SChunkData, payload)
	p.SendPacket(proto.SChunkBatchFinished, proto.NewWriter().VarInt(1).Payload())
}

func (sink *WorldSink) SendRequestedChunk(cx, cz int32, payload []byte) bool {
	p := sink.player
	p.chunkSendMu.Lock()
	defer p.chunkSendMu.Unlock()
	p.mu.Lock()
	key := [2]int32{cx, cz}
	if _, wanted := p.requestedChunks[key]; !wanted || p.worldEpoch != sink.epoch {
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

func (sink *WorldSink) IsChunkSent(cx, cz int32) bool {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.sentChunks[[2]int32{cx, cz}]
	return ok && p.worldEpoch == sink.epoch
}
func (sink *WorldSink) WantsChunk(cx, cz int32) bool {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.requestedChunks[[2]int32{cx, cz}]
	return ok && p.worldEpoch == sink.epoch
}

// A failed request releases only its own world's reservation. A late failure
// from the previous dimension must not cancel a new request at the same coords.
func (sink *WorldSink) UnmarkChunkSent(cx, cz int32) {
	p := sink.player
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.worldEpoch == sink.epoch {
		delete(p.sentChunks, [2]int32{cx, cz})
		delete(p.requestedChunks, [2]int32{cx, cz})
	}
}
