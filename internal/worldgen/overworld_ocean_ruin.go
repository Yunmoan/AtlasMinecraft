package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

// Ocean-ruin templates inspect neighboring plants and fluids while updating
// shapes. A canonical before-step cache cannot replay those callbacks: earlier
// neighboring chunks have already run later feature steps (e.g. seagrass).
// Place each clip in the actual decoration region with its original random.
func (g *OverworldGenerator) placeOceanRuinInDecoration(region *OverworldFeatureRegion, start StructureStart, cx, cz, index, step int32) error {
	private := start.Clone()
	random := NewWorldgenRandom(OverworldDecorationSeed(g.seed, cx, cz) + int64(index) + 10000*int64(step))
	worldRandom := NewStructureRegionRandom(g.seed, cx, cz)
	region.shapeRandom = worldRandom
	for i := range private.Pieces {
		piece := &private.Pieces[i]
		if piece.Box.MaxX < cx*16 || piece.Box.MinX > cx*16+15 || piece.Box.MaxZ < cz*16 || piece.Box.MinZ > cz*16+15 {
			continue
		}
		placed, err := PostProcessOceanRuinBlocks(region, piece, random, g.seed, cx, cz)
		if err != nil {
			return fmt.Errorf("ocean ruin %s in %d,%d: %w", start.Structure, cx, cz, err)
		}
		if err := g.finalizeOceanRuinActors(start, i, placed.PendingEntities, worldRandom); err != nil {
			return err
		}
		for _, request := range placed.PendingEntities {
			if err := region.AddPendingEntity(request); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *OverworldGenerator) finalizeOceanRuinActors(start StructureStart, pieceIndex int, requests []PendingStructureEntity, worldRandom *Xoroshiro) error {
	for actorIndex := range requests {
		request := &requests[actorIndex]
		if request.ID != "minecraft:drowned" {
			continue
		}
		request.SavedFlag = fmt.Sprintf("OceanRuinDrowned:%d,%d:%d:%d", start.Chunk.X, start.Chunk.Z, pieceIndex, actorIndex)
		finalized, err := FinalizeDrownedStructureEntity(*request, worldRandom, NewLegacyRandom(g.seed^int64(positionSeed(request.Position.X, request.Position.Y, request.Position.Z))), false)
		if err != nil {
			return err
		}
		request.Data = finalized.Data
		if finalized.ChickenJockey {
			request.ID = "minecraft:chicken"
			request.Persistent = false
			request.Data = finalized.ChickenData
			request.Data["Passengers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: finalized.Data}}}
			constructor := NewLegacyRandom(g.seed ^ int64(positionSeed(request.Position.X, request.Position.Y, request.Position.Z)))
			request.Data["EggLayTime"] = nbt.Tag{Type: nbt.Int, Value: constructor.NextIntBound(6000) + 6000}
		}
	}
	return nil
}
