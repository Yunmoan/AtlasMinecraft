package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/registry"
	"atlas/internal/world"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"
)

// handlePlay runs the play phase until the connection dies. The caller's
// goroutine becomes this player's packet-read loop; every write funnels
// through the player's outbound queue (WriterLoop goroutine).
func (s *Server) handlePlay(fc *proto.FramedConn, name string, uuid [16]byte, viewDist int32, skinParts byte, properties []player.ProfileProperty) {
	saved, hasSaved, err := s.World.LoadPlayer(uuid)
	if err != nil {
		slog.Error("player data load failed", "player", name, "error", err)
		s.kick(fc, proto.SKickDisconnect, "Could not load player data")
		return
	}
	inventory, err := s.World.LoadInventory(uuid)
	if err != nil {
		slog.Error("player inventory load failed", "player", name, "error", err)
		s.kick(fc, proto.SKickDisconnect, "Could not load player inventory")
		return
	}
	p := player.New(name, uuid, s.nextEntityID(), fc)
	if hasSaved {
		if _, ok := s.WorldInDimension(saved.Dimension); !ok {
			s.kick(fc, proto.SKickDisconnect, "Saved player dimension is unavailable")
			return
		}
		p.SetInitialDimension(saved.Dimension)
	}
	currentWorld := s.PlayerWorld(p)
	s.setInventory(p, inventory)
	p.ProfileProperties = properties
	p.SetSkinParts(skinParts)
	if !s.Hub.Join(p) {
		s.kick(fc, proto.SKickDisconnect, "You are already connected to this server")
		return
	}
	p.MovementMu.Lock()
	joining := true
	defer func() {
		if joining {
			p.MovementMu.Unlock()
		}
	}()
	p.SetViewDistance(s.effectiveViewDistance(viewDist))
	go p.WriterLoop()

	joined := false
	defer func() {
		s.forgetChest(p)
		s.savePlayer(p)
		p.Disconnect()
		s.Hub.Leave(p)
		s.despawnFor(p)
		if joined {
			s.broadcastPlayerPresence(p.Name, false)
		}
		s.Logf("player left: %s", name)
	}()

	// New players join at the stored world spawn point, as vanilla does. The
	// Y value is kept in level.dat (the boot-time spawn scan refines the
	// climate-searched placeholder into an exact safe column), so a spawn set
	// via /setworldspawn — even inside a cave — is honored exactly.
	spawnX, spawnY, spawnZ, spawnAngle := currentWorld.SpawnPoint()
	px, py, pz := float64(spawnX)+0.5, float64(spawnY), float64(spawnZ)+0.5
	yaw, pitch := spawnAngle, float32(0)
	if hasSaved {
		px, py, pz = saved.X, saved.Y, saved.Z
		p.SetPortalCooldown(saved.PortalCooldown)
		yaw, pitch = saved.Yaw, saved.Pitch
		p.SetFlying(saved.Flying)
	}
	// Set the server-side position before any packet references it.
	p.SetPos(px, py, pz, yaw, pitch, !hasSaved)
	cx, cz := p.ChunkPos()
	send := p.SendPacket

	// --- Join sequence ---
	send(proto.SLogin, s.joinGamePayload(p))

	send(proto.SSpawnPosition, proto.NewWriter().
		I64(proto.PackPosition(spawnX, spawnY, spawnZ)).F32(spawnAngle).Payload())
	abilityFlags := int8(0x0D)
	if p.Flying() {
		abilityFlags |= 0x02
	}
	send(proto.SAbilities, proto.NewWriter().I8(abilityFlags).F32(0.05).F32(0.1).Payload())
	send(proto.SDifficulty, proto.NewWriter().Bool(false).U8(1).Payload()) // unlocked, easy
	send(proto.SHeldItemSlot, proto.NewWriter().I8(int8(inventory.Selected)).Payload())
	s.sendInventory(p)
	s.sendWorldState(p)
	// Vanilla starts the client's level-loading tracker explicitly. Receiving
	// chunks alone cannot advance it past WAITING_FOR_SERVER without this event.
	send(proto.SGameStateChange, proto.NewWriter().U8(proto.GameEventLevelChunksLoadStart).F32(0).Payload())
	send(proto.SDeclareCommands, declareCommandsPayloadForLevel(s.operatorLevel(p)))
	send(proto.SPlayerInfo, playerInfoAdd(p))
	send(proto.SEntityMetadata, playerSkinMetadata(p))
	s.Hub.Others(p, func(o *player.Player) {
		if o.PlayReady() {
			o.SendPacket(proto.SPlayerInfo, playerInfoAdd(p))
			p.SendPacket(proto.SPlayerInfo, playerInfoAdd(o))
		}
	})
	s.syncPosition(p)

	send(proto.SSetViewPosition, proto.NewWriter().VarInt(cx).VarInt(cz).Payload())
	p.SetPlayReady(true)
	go func() {
		for {
			select {
			case <-p.ChunkUpdates:
				s.updateVisibility(p)
				s.updateGeneratedEntityVisibility(p)
			case <-p.Done:
				return
			}
		}
	}()
	// The player's current chunk goes out first, including saved positions
	// away from the boot-prepared spawn area.
	p.ReserveChunk(cx, cz)
	if err := currentWorld.SpawnChunk(p.WorldPacketSinkInDimension(currentWorld.DimensionName()), cx, cz, func(cx, cz int32) {}); err != nil {
		slog.Error("player chunk unavailable", "player", p.Name, "x", cx, "z", cz, "error", err)
		return
	}
	s.requestChunksAround(p, cx, cz)

	send(proto.SSystemChat, systemChat("§eWelcome to §b§lAtlas§r§7 — a Go-powered Minecraft server."))

	if warning := currentWorld.GenerationWarning(); warning != "" {
		send(proto.SSystemChat, systemChat("World generation: "+warning))
	}

	// Background tickers: keep-alives and periodic time sync. SendPacket is
	// safe from any goroutine, so these live outside the read loop.
	keep := time.NewTicker(15 * time.Second)
	defer keep.Stop()
	clock := time.NewTicker(5 * time.Second)
	defer clock.Stop()
	go func() {
		for {
			select {
			case <-keep.C:
				id := time.Now().UnixNano()
				p.SetKeepAliveID(id)
				p.SendPacket(proto.SKeepAlive, proto.NewWriter().I64(id).Payload())
			case <-clock.C:
				s.sendTime(p)
				s.validateChestSession(p)
			case <-p.Done:
				return
			}
		}
	}()

	s.Logf("player joined: %s (eid %d)", name, p.EID)
	joined = true
	s.broadcastPlayerPresence(p.Name, true)
	joining = false
	p.MovementMu.Unlock()

	for {
		fc.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		id, payload, err := fc.ReadPacket()
		if err != nil {
			return
		}
		if s.Cfg.Debug && (s.Cfg.TracePackets || !noisyPlayPacket(id)) {
			if _, known := playPacketNames[id]; known {
				s.Dbg("play <- %s (%d bytes) from %s", playPacketNames[id], len(payload), p.Name)
			} else {
				s.Dbg("play <- UNKNOWN 0x%02X (%d bytes) from %s", id, len(payload), p.Name)
			}
		}
		if !s.handlePlayPacket(p, id, payload) {
			return
		}
		select {
		case <-p.Done:
			return
		default:
		}
	}
}

func (s *Server) broadcastPlayerPresence(name string, joined bool) {
	payload := playerPresenceChat(name, joined)
	s.Hub.Each(func(p *player.Player) {
		if p.PlayReady() {
			p.SendPacket(proto.SSystemChat, payload)
		}
	})
}

func (s *Server) effectiveViewDistance(client int32) int32 {
	server := int32(s.Cfg.ViewDistance)
	if server < 2 {
		server = 2
	}
	if server > 16 {
		server = 16
	}
	if client < 2 {
		client = 2
	}
	if client < server {
		return client
	}
	return server
}

func (s *Server) joinGamePayload(p *player.Player) []byte {
	w := proto.NewWriter()
	w.I32(p.EID).Bool(false) // entityId, hardcore
	worlds := s.Worlds()
	w.VarInt(int32(len(worlds)))
	for _, dimension := range worlds {
		w.String(dimension.DimensionName())
	}
	w.VarInt(int32(s.Cfg.MaxPlayers))
	w.VarInt(p.ViewDistance()).VarInt(int32(s.Cfg.SimulationDistance))
	w.Bool(false).Bool(true).Bool(false) // reducedDebug, respawnScreen, limitedCrafting
	currentWorld := s.PlayerWorld(p)
	w.VarInt(dimensionTypeID(currentWorld.DimensionName())).String(currentWorld.DimensionName())
	w.I64(currentWorld.HashedSeed())          // vanilla BiomeManager seed hash
	w.I8(1).U8(255)                           // gamemode creative, no previous gamemode
	w.Bool(false).Bool(currentWorld.IsFlat()) // debug, flat
	w.Bool(false)                             // no death location
	w.VarInt(0)                               // portal cooldown
	w.Bool(false)                             // enforcesSecureChat
	return w.Payload()
}

func noisyPlayPacket(id int32) bool {
	switch id {
	case proto.CPlayerPosition, proto.CPlayerPositionLook, proto.CPlayerLook,
		proto.CPlayerFlying, proto.CChunkBatchReceived, proto.CKeepAlive:
		return true
	default:
		return false
	}
}

// handlePlayPacket dispatches one serverbound play packet. Returns false to
// end the connection.
func (s *Server) handlePlayPacket(p *player.Player, id int32, payload []byte) bool {
	switch id {
	case proto.CTeleportConfirm:
		if teleportID, err := proto.NewReader(payload).VarInt(); err == nil {
			p.ConfirmTeleport(teleportID)
		}
	case proto.CCustomPayload, proto.CClientCommand:
		// acknowledged / not implemented
	case proto.CPlayerAbilities:
		if flags, err := proto.NewReader(payload).U8(); err == nil {
			p.SetFlying(flags&0x02 != 0)
		}
	case proto.CEntityAction:
		s.entityMu.Lock()
		r := proto.NewReader(payload)
		if eid, err := r.VarInt(); err == nil && eid == p.EID {
			if action, err := r.VarInt(); err == nil {
				if action == 0 {
					p.SetCrouching(true)
				}
				if action == 1 {
					p.SetCrouching(false)
				}
				if action == 3 {
					p.SetSprinting(true)
				}
				if action == 4 {
					p.SetSprinting(false)
				}
				if action == 0 || action == 1 || action == 3 || action == 4 {
					metadata := playerPoseMetadata(p)
					p.SendPacket(proto.SEntityMetadata, metadata)
					s.sendTracked(p, proto.SEntityMetadata, metadata)
				}
			}
		}
		s.entityMu.Unlock()

	case proto.CChatCommand, proto.CChatCommandSigned:
		command, err := proto.NewReader(payload).String(256)
		if err == nil {
			s.RunCommand(p, command)
		}
	case proto.CCommandSuggestions:
		s.handleCommandSuggestions(p, payload)

	case proto.CClientSettings:
		clientDistance, parts, err := clientSettings(payload)
		if err != nil {
			break
		}
		if next := s.effectiveViewDistance(clientDistance); next != p.ViewDistance() {
			p.SetViewDistance(next)
			p.SendPacket(proto.SUpdateViewDistance, proto.NewWriter().VarInt(next).Payload())
			s.handleChunkChange(p)
		}
		if p.SetSkinParts(parts) {
			metadata := playerSkinMetadata(p)
			p.SendPacket(proto.SEntityMetadata, metadata)
			s.Hub.Others(p, func(o *player.Player) {
				if o.IsVisible(p.EID) {
					o.SendPacket(proto.SEntityMetadata, metadata)
				}
			})
		}

	case proto.CChunkBatchReceived:
		// Batches are sized and closed proactively by the server; the
		// client's measured throughput is currently ignored.

	case proto.CKeepAlive:
		// Tolerate slow/late responses like vanilla does; only log.
		if v, err := proto.NewReader(payload).I64(); err == nil && v != p.KeepAliveID() {
			s.Logf("keep-alive mismatch from %s (got %d)", p.Name, v)
		}

	case proto.CPlayerPosition, proto.CPlayerPositionLook, proto.CPlayerLook, proto.CPlayerFlying:
		s.handleMovement(p, id, payload)

	case proto.CBlockDig:
		s.handleDig(p, payload)

	case proto.CBlockPlace:
		s.handlePlace(p, payload)
	case proto.CInteractEntity:
		s.handleEntityInteraction(p, payload)

	case proto.CSetCreativeSlot:
		s.handleCreativeSlot(p, payload)
	case proto.CContainerClick:
		s.handleContainerClick(p, payload)
	case proto.CCloseContainer:
		if len(payload) == 1 && !s.closeChest(p, payload[0], false) && payload[0] == 0 {
			s.closeInventory(p)
		}

	case proto.CHeldItemSlot:
		if slot, err := proto.NewReader(payload).I16(); err == nil {
			p.SaveMu.Lock()
			p.SetHeld(int32(slot))
			p.SaveMu.Unlock()
			s.publishEquipment(p)
		}

	case proto.CArmAnimation:
		anim := proto.NewWriter().VarInt(p.EID).U8(0).Payload()
		s.entityMu.Lock()
		s.sendTracked(p, proto.SAnimation, anim)
		s.entityMu.Unlock()

	case proto.CChatMessage:
		msg, err := proto.NewReader(payload).String(256)
		if err == nil {
			s.relayChat(p, msg)
		}

	default:
		// unhandled serverbound packets are acceptable at this stage
	}
	return true
}

func (s *Server) relayChat(p *player.Player, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	if strings.HasPrefix(message, "/") {
		s.RunCommand(p, strings.TrimPrefix(message, "/"))
		return
	}
	line := "<" + p.Name + "> " + message
	s.Hub.Broadcast(proto.SSystemChat, styledChat(textSpan{Text: "<" + p.Name + "> ", Color: "aqua"}, textSpan{Text: message, Color: "white"}))
	slog.Info(line)
}

// handleMovement applies a position/rotation update, broadcasts it to
// viewers, and manages chunk streaming on chunk crossings.
func (s *Server) handleMovement(p *player.Player, id int32, payload []byte) {
	p.MovementMu.Lock()
	defer p.MovementMu.Unlock()
	if p.TeleportPending() {
		return
	}
	r := proto.NewReader(payload)
	ox, oy, oz, oyaw, opitch, ground := p.Pos()
	x, y, z := ox, oy, oz
	yaw, pitch := oyaw, opitch
	var err error
	switch id {
	case proto.CPlayerPosition:
		if x, err = r.F64(); err != nil {
			return
		}
		if y, err = r.F64(); err != nil {
			return
		}
		if z, err = r.F64(); err != nil {
			return
		}
		if g, e := r.Bool(); e == nil {
			ground = g
		}
	case proto.CPlayerPositionLook:
		if x, err = r.F64(); err != nil {
			return
		}
		if y, err = r.F64(); err != nil {
			return
		}
		if z, err = r.F64(); err != nil {
			return
		}
		if yaw, err = r.F32(); err != nil {
			return
		}
		if pitch, err = r.F32(); err != nil {
			return
		}
		if g, e := r.Bool(); e == nil {
			ground = g
		}
	case proto.CPlayerLook:
		if yaw, err = r.F32(); err != nil {
			return
		}
		if pitch, err = r.F32(); err != nil {
			return
		}
		if g, e := r.Bool(); e == nil {
			ground = g
		}
	case proto.CPlayerFlying:
		if g, e := r.Bool(); e == nil {
			ground = g
		}
	}

	moved := x != ox || y != oy || z != oz
	rotated := yaw != oyaw || pitch != opitch
	if math.IsNaN(x) || math.IsNaN(y) || math.IsNaN(z) ||
		math.IsInf(x, 0) || math.IsInf(y, 0) || math.IsInf(z, 0) ||
		math.Abs(x) > 30000000 || math.Abs(z) > 30000000 || math.Abs(y) > 30000000 ||
		math.IsNaN(float64(yaw)) || math.IsNaN(float64(pitch)) ||
		math.IsInf(float64(yaw), 0) || math.IsInf(float64(pitch), 0) {
		slog.Warn("invalid player movement", "player", p.Name)
		p.Disconnect()
		return
	}
	if moved {
		allowed, err := s.PlayerWorld(p).CanMovePlayer([3]float64{ox, oy, oz}, [3]float64{x, y, z}, playerHeight(p), ground)
		if !allowed || err != nil {
			s.Dbg("movement corrected for %s: collision or unavailable terrain (%v)", p.Name, err)
			s.syncPosition(p)
			return
		}
	}
	s.entityMu.Lock()
	chunkChanged := p.SetPos(x, y, z, yaw, pitch, ground)
	if moved || rotated || id == proto.CPlayerFlying {
		packetID, packet := entityMovementPacket(p.EID, [3]float64{ox, oy, oz}, [3]float64{x, y, z}, yaw, pitch, moved, rotated, ground)
		s.sendTracked(p, packetID, packet)
	}
	if rotated {
		s.sendTracked(p, proto.SEntityHeadLook, proto.NewWriter().VarInt(p.EID).I8(angleI8(yaw)).Payload())
	}
	s.updateVisibilityLocked(p)
	s.entityMu.Unlock()

	if chunkChanged {
		s.handleChunkChange(p)
	}
}

// playPacketNames names serverbound packets for -debug logging.
var playPacketNames = map[int32]string{
	proto.CTeleportConfirm:    "teleport_confirm",
	proto.CChatCommand:        "chat_command",
	proto.CChatCommandSigned:  "chat_command_signed",
	proto.CCommandSuggestions: "command_suggestions",
	proto.CChatMessage:        "chat_message",
	proto.CChunkBatchReceived: "chunk_batch_received",
	proto.CClientCommand:      "client_command",
	proto.CClientSettings:     "client_settings",
	proto.CCustomPayload:      "custom_payload",
	proto.CKeepAlive:          "keep_alive",
	proto.CPlayerPosition:     "position",
	proto.CPlayerPositionLook: "position_look",
	proto.CPlayerLook:         "look",
	proto.CPlayerFlying:       "flying",
	proto.CPlayerAbilities:    "abilities",
	proto.CEntityAction:       "entity_action",
	proto.CBlockDig:           "block_dig",
	proto.CBlockPlace:         "block_place",
	proto.CHeldItemSlot:       "held_item_slot",
	proto.CSetCreativeSlot:    "set_creative_slot",
	proto.CArmAnimation:       "arm_animation",
}

// handleChunkChange re-centers the client's chunk cache, prunes
// out-of-range chunks, streams the newly visible ones, and refreshes
// entity visibility.
func (s *Server) handleChunkChange(p *player.Player) {
	cx, cz := p.ChunkPos()
	vd := p.ViewDistance()
	// The client's chunk cache is a fixed window around the center it was
	// last told; moving without re-centering wraps slots and destroys world.
	p.SendPacket(proto.SSetViewPosition, proto.NewWriter().VarInt(cx).VarInt(cz).Payload())
	p.PruneChunks(cx, cz, (vd+2)*(vd+2))
	s.requestChunksAround(p, cx, cz)
	s.updateVisibility(p)
	s.updateGeneratedEntityVisibility(p)
}

func (s *Server) requestChunksAround(p *player.Player, cx, cz int32) {
	w := s.PlayerWorld(p)
	sink := p.WorldPacketSinkInDimension(w.DimensionName())
	if sink == nil {
		return
	}
	vd := p.ViewDistance()
	var positions [][2]int32
	for dx := -vd; dx <= vd; dx++ {
		for dz := -vd; dz <= vd; dz++ {
			if dx*dx+dz*dz > (vd+1)*(vd+1) {
				continue
			}
			ncx, ncz := cx+dx, cz+dz
			positions = append(positions, [2]int32{ncx, ncz})
		}
	}
	sort.SliceStable(positions, func(i, j int) bool {
		a, b := positions[i], positions[j]
		ax, az, bx, bz := int64(a[0]-cx), int64(a[1]-cz), int64(b[0]-cx), int64(b[1]-cz)
		return ax*ax+az*az < bx*bx+bz*bz
	})
	for _, k := range positions {
		if sink.ReserveChunk(k[0], k[1]) {
			w.RequestChunk(sink, k[0], k[1])
		}
	}
}

// updateVisibility spawns/despawns the mutual player entities based on
// current chunk distance and each side's view distance.
func (s *Server) updateVisibility(p *player.Player) {
	s.entityMu.Lock()
	defer s.entityMu.Unlock()
	s.updateVisibilityLocked(p)
}

func (s *Server) updateVisibilityLocked(p *player.Player) {
	cx, cz := p.ChunkPos()
	vd := p.ViewDistance()
	s.Hub.Others(p, func(o *player.Player) {
		ocx, ocz := o.ChunkPos()
		dx, dz := int64(ocx)-int64(cx), int64(ocz)-int64(cz)
		d2 := dx*dx + dz*dz

		sameDimension := p.Dimension() == o.Dimension()
		canSeeOther := sameDimension && p.PlayReady() && o.PlayReady() && d2 <= int64(vd+1)*int64(vd+1) && p.IsChunkSent(ocx, ocz)
		if canSeeOther && !p.IsVisible(o.EID) {
			s.spawnPlayerFor(p, o)
		} else if !canSeeOther && p.IsVisible(o.EID) {
			p.SendPacket(proto.SEntityDestroy, destroyPayload([]int32{o.EID}))
			p.MarkInvisible(o.EID)
		}

		ovd := o.ViewDistance()
		canSeePlayer := sameDimension && p.PlayReady() && o.PlayReady() && d2 <= int64(ovd+1)*int64(ovd+1) && o.IsChunkSent(cx, cz)
		if canSeePlayer && !o.IsVisible(p.EID) {
			s.spawnPlayerFor(o, p)
		} else if !canSeePlayer && o.IsVisible(p.EID) {
			o.SendPacket(proto.SEntityDestroy, destroyPayload([]int32{p.EID}))
			o.MarkInvisible(p.EID)
		}
	})
}

// Re-publish the profile before each spawn. The vanilla client discards a
// player spawn whose UUID is absent from its player-info cache; simultaneous
// joins and a client world reset can invalidate earlier profile publication.
func (s *Server) spawnPlayerFor(viewer, subject *player.Player) {
	viewer.SendPacket(proto.SPlayerInfo, playerInfoAdd(subject))
	viewer.SendPacket(proto.SSpawnEntity, spawnEntityPayload(subject))
	viewer.SendPacket(proto.SEntityMetadata, playerSkinMetadata(subject))
	viewer.SendPacket(proto.SEntityMetadata, playerPoseMetadata(subject))
	viewer.SendPacket(proto.SSetEquipment, equipmentPayload(subject))
	viewer.MarkVisible(subject.EID)
	s.Dbg("player entity spawned: viewer=%s player=%s eid=%d type=%d", viewer.Name, subject.Name, subject.EID, registry.EntityPlayerID)
}

// despawnFor removes a leaving player's entity and list entry everywhere.
func (s *Server) despawnFor(p *player.Player) {
	s.entityMu.Lock()
	defer s.entityMu.Unlock()
	p.SetPlayReady(false)
	s.clearGeneratedEntityView(p, false)
	rm := proto.NewWriter().VarInt(1).UUID(p.UUID).Payload()
	destroy := destroyPayload([]int32{p.EID})
	s.Hub.Others(p, func(o *player.Player) {
		if !o.PlayReady() {
			return
		}
		o.SendPacket(proto.SPlayerRemove, rm)
		if o.IsVisible(p.EID) {
			o.SendPacket(proto.SEntityDestroy, destroy)
		}
		o.MarkInvisible(p.EID)
	})
}

// handleDig implements block breaking (creative instant-mine semantics).
func (s *Server) handleDig(p *player.Player, payload []byte) {
	r := proto.NewReader(payload)
	status, err := r.VarInt()
	if err != nil {
		return
	}
	x, y, z, err := r.Position()
	if err != nil {
		return
	}
	if face, err := r.I8(); err != nil || face < 0 || face > 5 {
		return
	}
	seq, err := r.VarInt()
	if err != nil || seq < 0 || r.Remaining() != 0 {
		return
	}
	if status < 0 || status > 2 {
		return
	}
	defer p.SendPacket(proto.SAckBlockChange, proto.NewWriter().VarInt(seq).Payload())
	position := world.BlockPosition{X: x, Y: y, Z: z}
	currentWorld := s.PlayerWorld(p)
	if status == 0 && !p.TeleportPending() && canReachBlock(p, position) && y >= currentWorld.MinBuildHeight() && y < currentWorld.MaxBuildHeight() {
		states, err := currentWorld.ReadLoadedBlocks([]world.BlockPosition{position})
		if err == nil {
			if err := currentWorld.TryPlaceBlock(position, states[0], world.StateAir, nil); err != nil {
				s.Logf("dig failed for %s: %v", p.Name, err)
			}
		}
	}
	// Cancel/finish never start a second creative break. Correct all predictions.
	s.resendBlock(p, position)
}

// handlePlace implements block placement using the selected hotbar state.

// handleCreativeSlot stores complete stacks for every creative inventory slot.
func (s *Server) handleCreativeSlot(p *player.Player, payload []byte) {
	s.handleCreativeInventorySlot(p, payload)
}
