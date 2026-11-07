package world

import (
	"atlas/internal/light"
	"atlas/internal/proto"
	"atlas/internal/tickqueue"
	"atlas/internal/worldgen"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
)

// TickFluids executes one scheduled-fluid phase at the current world age.
// Eligibility is supplied by the caller's chunk tickets; it must be a pure,
// nonblocking predicate. No missing chunk is generated. Call AdvanceTime for
// each elapsed game tick; skipping straight to the final time is not equivalent.
// The server clock is not wired here until the remaining block-update handlers
// and ticket policy have complete reference coverage.
func (w *World) TickFluids(eligible func(int32, int32) bool) error {
	return w.tickScheduled(eligible, false)
}

// TickScheduled executes the original block phase followed by the fluid phase.
// Unsupported block handlers return an error and retain unexecuted requests.
// Eligibility and time are explicit, as for TickFluids.
func (w *World) TickScheduled(eligible func(int32, int32) bool) error {
	return w.tickScheduled(eligible, true)
}

func (w *World) tickScheduled(eligible func(int32, int32) bool, blocks bool) error {
	if eligible == nil {
		return fmt.Errorf("fluid ticking requires explicit chunk eligibility")
	}
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	select {
	case <-w.done:
		return ErrWorldClosed
	default:
	}
	if err := w.ensureTickRandom(); err != nil {
		return err
	}
	r := &fluidRuntime{world: w, settings: w.fluidSettings()}
	if err := r.recoverFallingSpawns(); err != nil {
		return err
	}
	if err := r.recoverKelpBreaks(); err != nil {
		return err
	}
	if err := w.scheduledTicks.activate(eligible, r.postProcess); err != nil {
		return err
	}
	if blocks {
		if err := w.scheduledTicks.tickBlocks(65536, eligible, func(t tickqueue.Scheduled[string]) error {
			return worldgen.TickScheduledBlock(r, worldgen.FeaturePosition{X: t.Position.X, Y: t.Position.Y, Z: t.Position.Z}, t.Type)
		}); err != nil {
			return err
		}
	}
	return w.scheduledTicks.tick(65536, eligible, func(t tickqueue.Scheduled[worldgen.FluidType]) error {
		p := worldgen.FeaturePosition{X: t.Position.X, Y: t.Position.Y, Z: t.Position.Z}
		state, err := r.BlockAt(p)
		if err != nil {
			return err
		}
		fluid, err := worldgen.FluidStateForBlock(state)
		if err != nil {
			return err
		}
		if fluid.Type != t.Type {
			return nil
		} // exact registered type, not family
		return worldgen.TickFluid(r, p, r.settings, w.tickRandom)
	})
}

// Caller holds blockUpdateMu. Runtime randomness is independent of terrain;
// reference replays supply its state inside package tests.
func (w *World) ensureTickRandom() error {
	if w.tickRandom != nil {
		return nil
	}
	var seed [8]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return err
	}
	w.tickRandom = worldgen.NewLegacyRandom(int64(binary.LittleEndian.Uint64(seed[:])))
	return nil
}

func (w *World) fluidSettings() worldgen.FluidSettings {
	s := worldgen.FluidSettings{UltraWarm: w.DimensionName() == "minecraft:the_nether", WaterSourceConversion: true}
	if w.store != nil {
		v := w.store.vanilla
		v.mu.Lock()
		defer v.mu.Unlock()
		root, _ := v.root.Object()
		data, _ := root.Object("Data")
		rules, _ := data.Object("GameRules")
		if value, ok := rules.Text("waterSourceConversion"); ok {
			s.WaterSourceConversion = strings.EqualFold(value, "true")
		}
		if value, ok := rules.Text("lavaSourceConversion"); ok {
			s.LavaSourceConversion = strings.EqualFold(value, "true")
		}
	}
	return s
}

type fluidRuntime struct {
	world        *World
	settings     worldgen.FluidSettings
	stack, added []*fluidNeighborUpdate
	chained      int
	fallingDrops []WorldEntity // staged until the falling entity's atomic save
}
type fluidNeighborUpdate struct {
	position     worldgen.FeaturePosition
	neighbor     worldgen.SurfaceBlockState
	face, next   int
	flags, limit int32
	multi        bool
}

var fluidNotifyOrder = [6]int{4, 5, 0, 1, 2, 3}
var fluidShapeOrder = [6]int{4, 5, 2, 3, 0, 1}

func fluidAdjacent(p worldgen.FeaturePosition, face int) worldgen.FeaturePosition {
	d := directionOffsets[face]
	return worldgen.FeaturePosition{X: p.X + d.X, Y: p.Y + d.Y, Z: p.Z + d.Z}
}
func liquidBlock(s worldgen.SurfaceBlockState) bool {
	return s.Name == "minecraft:water" || s.Name == "minecraft:lava"
}

func (r *fluidRuntime) InWorld(p worldgen.FeaturePosition) bool {
	return p.Y >= r.world.MinBuildHeight() && p.Y < r.world.MaxBuildHeight() && p.X >= -30000000 && p.X < 30000000 && p.Z >= -30000000 && p.Z < 30000000
}
func (r *fluidRuntime) MinBuildHeight() int32 { return r.world.MinBuildHeight() }
func (r *fluidRuntime) ShapeRandomInt(bound int32) (int32, error) {
	if err := r.world.ensureTickRandom(); err != nil {
		return 0, err
	}
	return r.world.tickRandom.NextIntBound(bound), nil
}
func (r *fluidRuntime) BlockAt(p worldgen.FeaturePosition) (worldgen.SurfaceBlockState, error) {
	if !r.InWorld(p) {
		return worldgen.SurfaceBlockState{Name: "minecraft:void_air"}, nil
	}
	state, _, err := r.world.loadedState(BlockPosition{X: p.X, Y: p.Y, Z: p.Z})
	return worldgen.SurfaceBlockState{Name: state.Name, Properties: state.Properties}, err
}
func (r *fluidRuntime) SetBlock(p worldgen.FeaturePosition, s worldgen.SurfaceBlockState) error {
	return r.SetFluidBlock(p, s, 2)
}

// Read the light engine's immutable snapshot; terrain height is not a proxy
// for brightness (e.g. underground torches and transparent roofs).
func (r *fluidRuntime) FeatureRawBrightness(p worldgen.FeaturePosition, skyDarkening int32) (int32, error) {
	if r.world.lighting == nil {
		return 0, fmt.Errorf("runtime survival requires lighting")
	}
	data := r.world.lighting.Snapshot(light.Position{X: p.X >> 4, Z: p.Z >> 4})
	if data == nil {
		return 0, fmt.Errorf("runtime survival light column is unavailable at %v", p)
	}
	layer := int((p.Y-r.world.MinBuildHeight())>>4) + 1
	if layer < 0 || layer >= len(data.Sky) || layer >= len(data.Block) {
		return 0, fmt.Errorf("runtime survival light position outside stored sections: %v", p)
	}
	index := int((p.Y&15)<<8 | (p.Z&15)<<4 | (p.X & 15))
	nibble := func(array []byte) int32 {
		if len(array) == 0 {
			return 0
		}
		return int32(array[index>>1] >> uint((index&1)*4) & 15)
	}
	return max(0, nibble(data.Sky[layer])-skyDarkening, nibble(data.Block[layer])), nil
}
func (r *fluidRuntime) SetFluidBlock(p worldgen.FeaturePosition, s worldgen.SurfaceBlockState, flags int32) error {
	return r.set(p, s, flags, 512)
}

func (r *fluidRuntime) set(p worldgen.FeaturePosition, s worldgen.SurfaceBlockState, flags, limit int32) error {
	if !r.InWorld(p) {
		return nil
	}
	before, err := r.BlockAt(p)
	if err != nil {
		return err
	}
	old, err := BlockStateID(before.Name, before.Properties)
	if err != nil {
		return err
	}
	id, err := BlockStateID(s.Name, s.Properties)
	if err != nil {
		return err
	}
	if old == id {
		return nil
	}
	if _, err := r.world.editBlock(setBlockMsg{x: p.X, y: p.Y, z: p.Z, state: id, expected: &old, raw: true, flags: flags}); err != nil {
		return err
	}
	// LevelChunk calls onPlace even for another state of the same liquid.
	if err := worldgen.NotifyFallingBlockPlacement(r, p, s); err != nil {
		return err
	}
	if err := worldgen.NotifyBrushableBlockPlacement(r, p, s); err != nil {
		return err
	}
	if err := worldgen.NotifyBubbleBlockPlacement(r, p, s); err != nil {
		return err
	}
	if err := worldgen.NotifyLiquidBlock(r, p, s, r.settings); err != nil {
		return err
	}
	current, err := r.BlockAt(p)
	if err != nil {
		return err
	}
	currentID, err := BlockStateID(current.Name, current.Properties)
	if err != nil {
		return err
	}
	if currentID != id {
		return nil
	} // onPlace (e.g. lava mixing) replaced it again
	if flags&1 != 0 {
		if err := r.NotifyFluidNeighbors(p, before.Name); err != nil {
			return err
		}
	}
	if flags&16 == 0 && limit > 0 {
		for _, face := range fluidShapeOrder {
			if err := r.enqueue(&fluidNeighborUpdate{position: fluidAdjacent(p, face), neighbor: s, face: face ^ 1, flags: flags &^ 33, limit: limit - 1}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (r *fluidRuntime) ScheduleFluid(p worldgen.FeaturePosition, kind worldgen.FluidType, delay int32) error {
	if !r.InWorld(p) {
		return nil
	}
	return r.world.scheduledTicks.schedule(p, kind, delay)
}

func (r *fluidRuntime) ScheduleBlockTick(p worldgen.FeaturePosition, name string, delay, priority int32) error {
	if !r.InWorld(p) {
		return nil
	}
	canonical, ok := parseScheduledBlock(name)
	if !ok {
		return fmt.Errorf("unknown scheduled block %q", name)
	}
	return r.world.scheduledTicks.scheduleBlock(p, canonical, delay, tickqueue.ClampPriority(priority))
}
func (r *fluidRuntime) ScheduleFluidUpdate(p worldgen.FeaturePosition) error {
	s, err := r.BlockAt(p)
	if err != nil {
		return err
	}
	fluid, err := worldgen.FluidStateForBlock(s)
	if err != nil {
		return err
	}
	return r.ScheduleFluid(p, fluid.Type, worldgen.FluidTickDelay(fluid.Type, r.settings.UltraWarm))
}
func (r *fluidRuntime) NotifyFluidNeighbors(p worldgen.FeaturePosition, _ string) error {
	return r.enqueue(&fluidNeighborUpdate{position: p, multi: true})
}

// CollectingNeighborUpdater drains nested additions before resuming the next
// face of a multi-neighbor update, while preserving insertion order among the
// additions. A plain FIFO or immediate recursive six-face loop is different.
func (r *fluidRuntime) enqueue(update *fluidNeighborUpdate) error {
	processing := r.chained > 0
	r.chained++
	if r.chained > 1000000 {
		return fmt.Errorf("chained fluid neighbor update limit exceeded")
	}
	if processing {
		r.added = append(r.added, update)
		return nil
	}
	r.stack = append(r.stack, update)
	defer func() { r.stack = nil; r.added = nil; r.chained = 0 }()
	for len(r.stack) > 0 || len(r.added) > 0 {
		for i := len(r.added) - 1; i >= 0; i-- {
			r.stack = append(r.stack, r.added[i])
		}
		r.added = nil
		current := r.stack[len(r.stack)-1]
		for len(r.added) == 0 {
			more, err := r.runNeighbor(current)
			if err != nil {
				return err
			}
			if !more {
				r.stack = r.stack[:len(r.stack)-1]
				break
			}
		}
	}
	return nil
}
func (r *fluidRuntime) runNeighbor(u *fluidNeighborUpdate) (bool, error) {
	if u.multi {
		p := fluidAdjacent(u.position, fluidNotifyOrder[u.next])
		u.next++
		s, err := r.BlockAt(p)
		if err != nil {
			return false, err
		}
		return u.next < 6, worldgen.NotifyLiquidBlock(r, p, s, r.settings)
	}
	if !r.InWorld(u.position) {
		return false, nil
	}
	before, err := r.BlockAt(u.position)
	if err != nil {
		return false, err
	}
	if liquidBlock(before) {
		return false, worldgen.UpdateLiquidShape(r, u.position, before, u.neighbor, r.settings)
	}
	after, err := worldgen.UpdateFluidNeighborShape(r, u.position, before, u.face, u.neighbor)
	if err != nil {
		return false, err
	}
	oldID, err := BlockStateID(before.Name, before.Properties)
	if err != nil {
		return false, err
	}
	newID, err := BlockStateID(after.Name, after.Properties)
	if err != nil {
		return false, err
	}
	if oldID == newID {
		return false, nil
	}
	if isAir(newID) {
		fluid, err := worldgen.FluidStateForBlock(before)
		if err != nil {
			return false, err
		}
		after = fluid.LegacyBlock()
		// Block.updateOrDestroy calls Level.destroyBlock for an air result.
		// That path writes the fluid legacy state with flags 3 regardless of
		// the incoming shape flags, notifying neighbors before shape updates.
		// In particular, removing tall seagrass must not keep flags 2 here.
		return false, r.set(u.position, after, 3, u.limit)
	}
	return false, r.set(u.position, after, u.flags, u.limit)
}

func (r *fluidRuntime) postProcess(p worldgen.FeaturePosition) error {
	before, err := r.BlockAt(p)
	if err != nil {
		return err
	}
	fluid, err := worldgen.FluidStateForBlock(before)
	if err != nil {
		return err
	}
	if fluid.Type != worldgen.FluidEmpty {
		if err := worldgen.TickFluid(r, p, r.settings, r.world.tickRandom); err != nil {
			return err
		}
	}
	if liquidBlock(before) {
		return nil
	}
	after := before
	for _, face := range fluidShapeOrder {
		other, err := r.BlockAt(fluidAdjacent(p, face))
		if err != nil {
			return err
		}
		after, err = worldgen.UpdateFluidNeighborShape(r, p, after, face, other)
		if err != nil {
			return err
		}
	}
	return r.SetFluidBlock(p, after, 20)
}
func (r *fluidRuntime) FluidLevelEvent(p worldgen.FeaturePosition, event, data int32) error {
	payload := proto.NewWriter().I32(event).I64(proto.PackPosition(p.X, p.Y, p.Z)).I32(data).Bool(false).Payload()
	if r.world.hub != nil {
		for _, sink := range r.world.hub.ViewersOf(p.X, p.Z) {
			if chunkDelivered(sink, p.X>>4, p.Z>>4) {
				sink.SendPacket(proto.SLevelEvent, payload)
			}
		}
	}
	return nil
}
func (r *fluidRuntime) BeforeFluidDestroysBlock(p worldgen.FeaturePosition, _ worldgen.SurfaceBlockState, kind worldgen.FluidType) error {
	if kind == worldgen.FluidLava || kind == worldgen.FluidFlowingLava {
		return r.FluidLevelEvent(p, 1501, 0)
	}
	// Dropped item entities are not implemented by the existing world block
	// update path. This callback currently preserves terrain effects only.
	return nil
}
func (r *fluidRuntime) PlaceFluidInContainer(p worldgen.FeaturePosition, state worldgen.SurfaceBlockState, fluid worldgen.FluidState) error {
	_, err := worldgen.PlaceFluidInBlock(r, p, state, fluid)
	return err
}
