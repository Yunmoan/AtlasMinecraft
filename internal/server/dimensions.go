package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func (s *Server) AddWorld(w *world.World) error {
	if w == nil {
		return errors.New("cannot register nil world")
	}
	s.worldMu.Lock()
	defer s.worldMu.Unlock()
	if s.worlds == nil {
		s.worlds = map[string]*world.World{"minecraft:overworld": s.World}
	}
	name := w.DimensionName()
	if _, exists := s.worlds[name]; exists {
		return fmt.Errorf("dimension %s already registered", name)
	}
	s.worlds[name] = w
	return nil
}
func (s *Server) WorldInDimension(name string) (*world.World, bool) {
	s.worldMu.RLock()
	defer s.worldMu.RUnlock()
	if name == "" || name == "minecraft:overworld" {
		return s.World, s.World != nil
	}
	w, ok := s.worlds[name]
	return w, ok
}
func (s *Server) PlayerWorld(p *player.Player) *world.World {
	if p == nil {
		return s.World
	}
	w, ok := s.WorldInDimension(p.Dimension())
	if !ok {
		return s.World
	}
	return w
}
func (s *Server) Worlds() []*world.World {
	s.worldMu.RLock()
	defer s.worldMu.RUnlock()
	if len(s.worlds) == 0 {
		return []*world.World{s.World}
	}
	var names []string
	for name := range s.worlds {
		names = append(names, name)
	}
	sort.Strings(names)
	var worlds []*world.World
	for _, name := range names {
		worlds = append(worlds, s.worlds[name])
	}
	return worlds
}
func (s *Server) SaveWorlds() error {
	var errs []error
	for _, w := range s.Worlds() {
		if err := w.Save(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", w.DimensionName(), err))
		}
	}
	return errors.Join(errs...)
}

func dimensionTypeID(name string) int32 {
	switch name {
	case "minecraft:the_nether":
		return 3
	case "minecraft:the_end":
		return 2
	default:
		return 0
	}
}
func (s *Server) respawnPayload(p *player.Player, w *world.World) []byte {
	return proto.NewWriter().VarInt(dimensionTypeID(w.DimensionName())).String(w.DimensionName()).I64(w.HashedSeed()).I8(1).I8(1).
		Bool(false).Bool(w.IsFlat()).Bool(false).VarInt(0).U8(3).Payload()
}

// TeleportToDimension loads the destination before resetting the client's
// world. Failed generation leaves the original world and position intact.
func (s *Server) TeleportToDimension(p *player.Player, name string, xyz [3]float64) error {
	return s.teleportToDimension(p, name, xyz, nil)
}

func (s *Server) teleportToDimension(p *player.Player, name string, xyz [3]float64, rotation *[2]float32) error {
	return s.teleportFromWorld(p, name, xyz, rotation, nil)
}

func (s *Server) teleportFromWorld(p *player.Player, name string, xyz [3]float64, rotation *[2]float32, source *player.WorldSink) error {
	w, ok := s.WorldInDimension(name)
	if !ok {
		return fmt.Errorf("dimension %s is unavailable", name)
	}
	if err := validateCommandCoordinates(xyz, false); err != nil {
		return err
	}
	p.MovementMu.Lock()
	defer p.MovementMu.Unlock()
	if source != nil {
		if !source.Current() || !p.PlayReady() {
			return errors.New("portal source world changed")
		}
		select {
		case <-p.Done:
			return errors.New("player disconnected during portal preparation")
		default:
		}
		if _, kind, err := s.playerPortalContact(p, s.PlayerWorld(p)); err != nil || kind == "" {
			return errors.New("player left the portal during preparation")
		}
	}
	p.CancelPendingChunks()
	cx, cz := int32(math.Floor(xyz[0]/16)), int32(math.Floor(xyz[2]/16))
	if err := w.PrepareDestination(context.Background(), cx, cz); err != nil {
		oldX, oldZ := p.ChunkPos()
		s.requestChunksAround(p, oldX, oldZ)
		return err
	}
	_, _, _, yaw, pitch, _ := p.Pos()
	if rotation != nil {
		yaw, pitch = rotation[0], rotation[1]
	}
	s.closeChest(p, 255, true)
	if p.Dimension() == w.DimensionName() {
		s.entityMu.Lock()
		changed := p.SetPos(xyz[0], xyz[1], xyz[2], yaw, pitch, false)
		s.syncPosition(p)
		s.sendTracked(p, proto.SEntityTeleport, proto.NewWriter().VarInt(p.EID).F64(xyz[0]).F64(xyz[1]).F64(xyz[2]).
			I8(angleI8(yaw)).I8(angleI8(pitch)).Bool(false).Payload())
		s.updateVisibilityLocked(p)
		s.entityMu.Unlock()
		if changed {
			s.handleChunkChange(p)
		} else {
			s.requestChunksAround(p, cx, cz)
		}
		return nil
	}
	s.entityMu.Lock()
	s.clearGeneratedEntityView(p, true)
	p.ResetWorld(w.DimensionName(), xyz, yaw, pitch, s.respawnPayload(p, w))
	s.updateVisibilityLocked(p)
	s.entityMu.Unlock()
	sx, sy, sz, angle := w.SpawnPoint()
	p.SendPacket(proto.SSpawnPosition, proto.NewWriter().I64(proto.PackPosition(sx, sy, sz)).F32(angle).Payload())
	s.sendWorldState(p)
	p.SendPacket(proto.SGameStateChange, proto.NewWriter().U8(proto.GameEventLevelChunksLoadStart).F32(0).Payload())
	p.SendPacket(proto.SAbilities, proto.NewWriter().I8(0x0D|boolFlyingFlag(p)).F32(0.05).F32(0.1).Payload())
	s.syncPosition(p)
	p.SendPacket(proto.SSetViewPosition, proto.NewWriter().VarInt(cx).VarInt(cz).Payload())
	p.ReserveChunk(cx, cz)
	if err := w.SpawnChunk(p.WorldPacketSinkInDimension(w.DimensionName()), cx, cz, func(int32, int32) {}); err != nil {
		return err
	}
	s.requestChunksAround(p, cx, cz)
	return nil
}
func boolFlyingFlag(p *player.Player) int8 {
	if p.Flying() {
		return 2
	}
	return 0
}

func (s *Server) commandExecuteIn(source *player.Player, args []string) string {
	usage := "Usage: /execute in <dimension> run <tp|setblock> ..."
	if len(args) < 4 || args[0] != "in" || args[2] != "run" {
		return usage
	}
	targetWorld, ok := s.WorldInDimension(args[1])
	if !ok {
		return "Dimension unavailable: " + args[1]
	}
	command := strings.TrimPrefix(args[3], "minecraft:")
	args = append([]string(nil), args[4:]...)
	origin := source
	if source != nil && source.Dimension() != targetWorld.DimensionName() {
		x, y, z, yaw, pitch, ground := source.Pos()
		scale := 1.0
		if source.Dimension() == "minecraft:the_nether" {
			scale = 8
		}
		if targetWorld.DimensionName() == "minecraft:the_nether" {
			scale /= 8
		}
		origin = player.New(source.Name, source.UUID, source.EID, nil)
		origin.SetPos(x*scale, y, z*scale, yaw, pitch, ground)
	}
	resolve := func(offset int, block bool) error {
		position, err := resolveCoordinates(origin, args[offset:offset+3], block)
		if err != nil {
			return err
		}
		for i, value := range position {
			formatted := strconv.FormatFloat(value, 'f', -1, 64)
			if !block && i != 1 && !strings.Contains(formatted, ".") {
				formatted += ".0"
			}
			args[offset+i] = formatted
		}
		return nil
	}
	switch command {
	case "tp", "teleport":
		if len(args) == 3 || len(args) == 4 {
			if err := resolve(len(args)-3, false); err != nil {
				return err.Error()
			}
		}
		return s.commandTeleportIn(source, args, targetWorld)
	case "setblock":
		if len(args) < 4 {
			return usage
		}
		if err := resolve(0, true); err != nil {
			return err.Error()
		}
		return s.commandSetBlockIn(source, args, targetWorld)
	default:
		return "Supported execute commands: tp, setblock"
	}
}
