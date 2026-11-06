package world

import (
	"atlas/internal/worldgen"
	"math"
)

// A freshly loaded brushable entity has zero progress, even if its block has
// nonzero DUSTED. Counts/timers/direction live with the chunk and are discarded
// on replacement or unload, as in BrushableBlockEntity's load/save methods.
func (r *fluidRuntime) ResetBrushableBlockProgress(p worldgen.FeaturePosition, state worldgen.SurfaceBlockState) error {
	key := BlockPosition{X: p.X, Y: p.Y, Z: p.Z}
	progress := worldgen.BrushableProgress{HitDirection: -1}
	found := false
	if err := r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(_ *Region, ch *Chunk) error {
		root, _ := ch.anvilRoot.Object()
		entities, _ := root["block_entities"].Elements()
		for _, tag := range entities {
			entity, _ := tag.Object()
			id, _ := entity.Text("id")
			x, xok := entity.Number("x")
			y, yok := entity.Number("y")
			z, zok := entity.Number("z")
			if id == "minecraft:brushable_block" && xok && yok && zok && x == int64(p.X) && y == int64(p.Y) && z == int64(p.Z) {
				found = true
				break
			}
		}
		if value, ok := ch.brushableProgress[key]; ok {
			progress = value
		}
		return nil
	}); err != nil {
		return err
	}
	if !found {
		return nil
	}
	now, _ := r.world.Time()
	after, err := worldgen.ResetBrushableProgress(r, p, state, progress, now)
	if err != nil {
		return err
	}
	return r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(_ *Region, ch *Chunk) error {
		if after.Count == 0 {
			delete(ch.brushableProgress, key)
		} else {
			if ch.brushableProgress == nil {
				ch.brushableProgress = make(map[BlockPosition]worldgen.BrushableProgress)
			}
			ch.brushableProgress[key] = after
		}
		return nil
	})
}

func (r *fluidRuntime) BreakFallingBrushableBlock(center [3]float64, state worldgen.SurfaceBlockState) error {
	p := worldgen.FeaturePosition{X: int32(math.Floor(center[0])), Y: int32(math.Floor(center[1])), Z: int32(math.Floor(center[2]))}
	id, err := BlockStateID(state.Name, state.Properties)
	if err != nil {
		return err
	}
	if err := r.FluidLevelEvent(p, 2001, int32(id)); err != nil {
		return err
	}
	return r.FallingGameEvent(p, "minecraft:block_destroy") // vibration consumers are not wired yet
}
