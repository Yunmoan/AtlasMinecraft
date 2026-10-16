package server

import (
	"atlas/internal/player"
	"atlas/internal/world"
	"context"
	"log/slog"
	"math"
)

// TickPortals schedules bounded player work; destination generation and actor
// reads never run in the server's 20 Hz world-clock loop.
func (s *Server) TickPortals(ticks int64) {
	if ticks <= 0 {
		return
	}
	for _, p := range s.Hub.Snapshot() {
		if p.BeginPortalCheck() {
			go s.checkPlayerPortal(p, int32(min(ticks, 300)))
		}
	}
}

func (s *Server) playerPortalContact(p *player.Player, w *world.World) (world.BlockPosition, string, error) {
	x, y, z, _, _, _ := p.Pos()
	const epsilon = 1e-7
	var positions []world.BlockPosition
	for by := int32(math.Floor(y + epsilon)); by <= int32(math.Floor(y+playerHeight(p)-epsilon)); by++ {
		if by < w.MinBuildHeight() || by >= w.MaxBuildHeight() {
			continue
		}
		for bz := int32(math.Floor(z - .3 + epsilon)); bz <= int32(math.Floor(z+.3-epsilon)); bz++ {
			for bx := int32(math.Floor(x - .3 + epsilon)); bx <= int32(math.Floor(x+.3-epsilon)); bx++ {
				positions = append(positions, world.BlockPosition{X: bx, Y: by, Z: bz})
			}
		}
	}
	if len(positions) == 0 {
		return world.BlockPosition{}, "", nil
	}
	states, err := w.ReadLoadedBlocks(positions)
	if err != nil {
		return world.BlockPosition{}, "", err
	}
	for i, id := range states {
		state, err := world.BlockStateByID(id)
		if err != nil {
			return world.BlockPosition{}, "", err
		}
		switch state.Name {
		case "minecraft:nether_portal":
			if w.DimensionName() != "minecraft:the_end" {
				return positions[i], state.Name, nil
			}
		case "minecraft:end_portal":
			if w.DimensionName() == "minecraft:overworld" && y < float64(positions[i].Y)+.75 && y+playerHeight(p) > float64(positions[i].Y)+.375 {
				return positions[i], state.Name, nil
			}
		}
	}
	return world.BlockPosition{}, "", nil
}

func (s *Server) checkPlayerPortal(p *player.Player, ticks int32) {
	transferred := false
	defer func() { p.EndPortalCheck(transferred) }()
	p.MovementMu.Lock()
	w := s.PlayerWorld(p)
	source := p.WorldPacketSinkInDimension(w.DimensionName())
	if source == nil {
		p.MovementMu.Unlock()
		return
	}
	portal, kind, err := s.playerPortalContact(p, w)
	delay := int32(0)
	if kind == "minecraft:nether_portal" {
		delay = 1
	}
	ready := p.AdvancePortalTick(err == nil && kind != "", ticks, delay)
	dimension, x, y, z, yaw, pitch, flying := p.WorldPosition()
	height := float32(playerHeight(p))
	p.MovementMu.Unlock()
	if !ready {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		select {
		case <-p.Done:
			cancel()
		case <-ctx.Done():
		}
	}()
	position := world.PlayerPosition{Dimension: dimension, X: x, Y: y, Z: z, Yaw: yaw, Pitch: pitch, Flying: flying}
	var arrival world.PlayerPosition
	switch kind {
	case "minecraft:nether_portal":
		name := "minecraft:the_nether"
		if dimension == name {
			name = "minecraft:overworld"
		}
		destination, ok := s.WorldInDimension(name)
		if !ok {
			return
		}
		arrival, _, err = destination.NetherPortalArrival(ctx, w, portal, position, [3]float64{}, .6, height)
	case "minecraft:end_portal":
		destination, ok := s.WorldInDimension("minecraft:the_end")
		if !ok {
			return
		}
		arrival, err = destination.PrepareEndArrival()
	default:
		return
	}
	if err == nil {
		rotation := [2]float32{arrival.Yaw, arrival.Pitch}
		err = s.teleportFromWorld(p, arrival.Dimension, [3]float64{arrival.X, arrival.Y, arrival.Z}, &rotation, source)
	}
	if err != nil {
		p.SetPortalCooldown(20)
		slog.Warn("portal destination unavailable", "player", p.Name, "error", err)
		return
	}
	transferred = true
}
