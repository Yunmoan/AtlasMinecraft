package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/registry"
	"math"
)

// sendToViewers delivers a packet to everyone who can see block (x, z),
// except the acting player.
func sendToViewers(h *hub.Hub, x, z int32, id int32, payload []byte, except *player.Player) {
	for _, v := range h.ViewersOf(x, z) {
		if pp, ok := v.(*player.Player); ok && pp.EID == except.EID {
			continue
		}
		v.SendPacket(id, payload)
	}
}

// sendTracked requires the entity to have been spawned and its current chunk
// delivered. entityMu serializes this with visibility and position publication.
func (s *Server) sendTracked(subject *player.Player, id int32, payload []byte) {
	cx, cz := subject.ChunkPos()
	s.Hub.Others(subject, func(viewer *player.Player) {
		if viewer.PlayReady() && viewer.IsVisible(subject.EID) && viewer.IsChunkSent(cx, cz) {
			viewer.SendPacket(id, payload)
		}
	})
}

func (s *Server) syncPosition(p *player.Player) {
	x, y, z, yaw, pitch, _ := p.Pos()
	id := s.nextTeleportID.Add(1)
	p.BeginTeleport(id)
	p.SendPacket(proto.SPositionSync, proto.NewWriter().F64(x).F64(y).F64(z).F32(yaw).F32(pitch).U8(0).VarInt(id).Payload())
}

// Vanilla encodes the difference of absolute fixed-point coordinates. Rounding
// each displacement independently would accumulate error, especially at x<0.
func entityMovementPacket(eid int32, from, to [3]float64, yaw, pitch float32, moved, rotated, ground bool) (int32, []byte) {
	w := proto.NewWriter().VarInt(eid)
	var delta [3]int64
	for i := range delta {
		delta[i] = int64(math.Floor(to[i]*4096)) - int64(math.Floor(from[i]*4096))
		if delta[i] < -32768 || delta[i] > 32767 {
			return proto.SEntityTeleport, w.F64(to[0]).F64(to[1]).F64(to[2]).I8(angleI8(yaw)).I8(angleI8(pitch)).Bool(ground).Payload()
		}
	}
	if moved || !rotated {
		w.I16(int16(delta[0])).I16(int16(delta[1])).I16(int16(delta[2]))
		if rotated {
			return proto.SEntityMoveLook, w.I8(angleI8(yaw)).I8(angleI8(pitch)).Bool(ground).Payload()
		}
		return proto.SRelEntityMove, w.Bool(ground).Payload()
	}
	return proto.SEntityLook, w.I8(angleI8(yaw)).I8(angleI8(pitch)).Bool(ground).Payload()
}

func playerPoseMetadata(p *player.Player) []byte {
	flags := byte(0)
	pose := int32(0)
	if p.Crouching() {
		flags |= 0x02
		pose = 5
	}
	if p.Sprinting() {
		flags |= 0x08
	}
	return proto.NewWriter().VarInt(p.EID).U8(0).VarInt(0).U8(flags).U8(6).VarInt(21).VarInt(pose).U8(0xff).Payload()
}

// angleI8 converts degrees to the protocol's byte angle representation.
func angleI8(a float32) int8 { return int8(a / 360 * 256) }

// spawnEntityPayload builds a Spawn Entity packet for a player entity.
func spawnEntityPayload(q *player.Player) []byte {
	x, y, z, yaw, pitch, _ := q.Pos()
	w := proto.NewWriter()
	w.VarInt(q.EID).UUID(q.UUID).VarInt(registry.EntityPlayerID)
	w.F64(x).F64(y).F64(z)
	w.I8(angleI8(pitch)).I8(angleI8(yaw)).I8(angleI8(yaw)) // pitch, yaw, head yaw
	w.VarInt(0)                                            // data
	w.I16(0).I16(0).I16(0)                                 // velocity
	return w.Payload()
}

// Player metadata index 17 is the 1.21.1 displayed skin-parts byte. Omitting
// it leaves other clients with the entity's zero-valued outer-layer mask.
func playerSkinMetadata(p *player.Player) []byte {
	return proto.NewWriter().VarInt(p.EID).
		U8(17).VarInt(0).U8(p.SkinParts()). // metadata type 0 = byte
		U8(0xff).Payload()                  // end of metadata entries
}

// destroyPayload builds an Entity Destroy packet for the given entity ids.
func destroyPayload(eids []int32) []byte {
	w := proto.NewWriter().VarInt(int32(len(eids)))
	for _, e := range eids {
		w.VarInt(e)
	}
	return w.Payload()
}

// playerInfoAdd builds a Player Info Update adding the player to the tab
// list with the add_player, update_game_mode, update_listed and
// update_latency actions. Note: add_player itself carries ONLY the name and
// property list in 1.19.3+; gamemode/listed/latency are separate actions.
func playerInfoAdd(p *player.Player) []byte {
	w := proto.NewWriter()
	w.U8(0x01 | 0x04 | 0x08 | 0x10) // action bitfield
	w.VarInt(1)                     // one entry
	w.UUID(p.UUID)
	// add_player: name + properties only
	w.String(p.Name)
	w.VarInt(int32(len(p.ProfileProperties)))
	for _, property := range p.ProfileProperties {
		w.String(property.Name).String(property.Value).Bool(property.Signature != "")
		if property.Signature != "" {
			w.String(property.Signature)
		}
	}
	// update_game_mode
	w.VarInt(1) // creative
	// update_listed
	w.VarInt(1)
	// update_latency
	w.VarInt(0)
	return w.Payload()
}

// declareCommandsPayload exposes executable literals and greedy arguments so
// a vanilla client accepts and suggests the commands Atlas actually handles.
func declareCommandsPayload(operator bool) []byte {
	if operator {
		return declareCommandsPayloadForLevel(4)
	}
	return declareCommandsPayloadForLevel(0)
}

func declareCommandsPayloadForLevel(level int) []byte {
	names := []string{"help", "list", "where"}
	if level >= 2 {
		names = append(names, "tps", "mspt", "status", "sysinfo", "seed", "time", "weather", "tp", "setblock", "save")
	}
	if level >= 3 {
		names = append(names, "op", "deop")
	}
	withArgs := map[string]bool{"help": true, "op": true, "deop": true, "time": true, "weather": true, "tp": true, "setblock": true}
	type node struct {
		name     string
		children []int32
		argument bool
	}
	nodes := []node{{}}
	for _, name := range names {
		idx := int32(len(nodes))
		nodes[0].children = append(nodes[0].children, idx)
		nodes = append(nodes, node{name: name})
		if withArgs[name] {
			nodes[idx].children = []int32{int32(len(nodes))}
			nodes = append(nodes, node{name: "args", argument: true})
		}
	}
	w := proto.NewWriter()
	w.VarInt(int32(len(nodes)))
	for i, n := range nodes {
		if i == 0 {
			w.U8(0)
		} else if n.argument {
			w.U8(0x06) // argument + executable
		} else {
			w.U8(0x05) // literal + executable
		}
		w.VarInt(int32(len(n.children)))
		for _, child := range n.children {
			w.VarInt(child)
		}
		if i > 0 {
			w.String(n.name)
			if n.argument {
				w.VarInt(5).VarInt(2) // brigadier:string, greedy phrase
			}
		}
	}
	w.VarInt(0)
	return w.Payload()
}

// faceOffset maps a placement direction (0..5) to the neighbor block delta.
// 0=-Y, 1=+Y, 2=-Z, 3=+Z, 4=-X, 5=+X.
func faceOffset(dir int32) (dx, dy, dz int32) {
	switch dir {
	case 0:
		return 0, -1, 0
	case 1:
		return 0, 1, 0
	case 2:
		return 0, 0, -1
	case 3:
		return 0, 0, 1
	case 4:
		return -1, 0, 0
	default:
		return 1, 0, 0
	}
}
