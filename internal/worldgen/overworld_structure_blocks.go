package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"sort"
	"strings"
)

type structureBlockKey struct {
	name string
	x, z int32
}
type structureBlockEntry struct {
	ready    chan struct{}
	result   *structureBlockResult
	err      error
	used     uint64
	complete bool
}
type structureBlockResult struct {
	start          StructureStart
	blocks         map[FeaturePosition]SurfaceBlockState
	postProcessing map[FeaturePosition]bool
	blockEntities  map[FeaturePosition]featureBlockEntity
	pending        []PendingStructureEntity
}

// A structure is aligned and postprocessed once in its own canonical region.
// Cached records include only structure writes, not preceding features. This
// keeps HPos, chest flags and cross-chunk room geometry stable when each final
// chunk is later requested independently or by a different worker.
type structureRecordingWorld struct {
	region *OverworldFeatureRegion
	result *structureBlockResult
}

func (w *structureRecordingWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	return w.region.BlockAt(p)
}
func (w *structureRecordingWorld) InWorld(p FeaturePosition) bool { return w.region.InWorld(p) }
func (w *structureRecordingWorld) MinBuildHeight() int32          { return w.region.minY }
func (w *structureRecordingWorld) WorldSeed() int64               { return w.region.WorldSeed() }
func (w *structureRecordingWorld) GenerationSeaLevel() int32      { return w.region.generationSeaLevel }
func (w *structureRecordingWorld) HeightmapAt(kind string, x, z int32) (int32, error) {
	return w.region.HeightmapAt(kind, x, z)
}
func (w *structureRecordingWorld) SetBlock(p FeaturePosition, s SurfaceBlockState) error {
	if err := w.region.SetBlock(p, s); err != nil {
		return err
	}
	w.result.blocks[p] = cloneSurfaceState(s)
	if entity, ok := w.result.blockEntities[p]; ok && entity.block != s.Name {
		delete(w.result.blockEntities, p)
	}
	return nil
}
func (w *structureRecordingWorld) SetBlockEntity(p FeaturePosition, block string, data nbt.Object) error {
	if err := w.region.SetBlockEntity(p, block, data); err != nil {
		return err
	}
	w.result.blockEntities[p] = featureBlockEntity{block: block, data: cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: data}).Value.(nbt.Object)}
	return nil
}
func (w *structureRecordingWorld) MarkPostProcessing(p FeaturePosition) error {
	if err := w.region.MarkPostProcessing(p); err != nil {
		return err
	}
	w.result.postProcessing[p] = true
	return nil
}
func (w *structureRecordingWorld) ScheduleFluidUpdate(p FeaturePosition) error {
	return w.MarkPostProcessing(p)
}

func (g *OverworldGenerator) structureBlocksFor(start StructureStart) (*structureBlockResult, error) {
	if start.Structure == "minecraft:stronghold" {
		return g.strongholdStructureBlocksFor(start, start.Chunk.X, start.Chunk.Z)
	}
	if isOverworldJigsawStructure(start.Structure) {
		return g.jigsawStructureBlocksFor(start, start.Chunk.X, start.Chunk.Z)
	}
	c := g.structures
	key := structureBlockKey{start.Structure, start.Chunk.X, start.Chunk.Z}
	c.mu.Lock()
	c.clock++
	if entry := c.blocks[key]; entry != nil {
		entry.used = c.clock
		c.mu.Unlock()
		<-entry.ready
		return entry.result, entry.err
	}
	entry := &structureBlockEntry{ready: make(chan struct{}), used: c.clock}
	c.blocks[key] = entry
	c.mu.Unlock()
	result, err := g.generateStructureBlocks(start)
	c.mu.Lock()
	entry.result, entry.err, entry.complete = result, err, true
	close(entry.ready)
	if err != nil {
		delete(c.blocks, key)
	}
	for len(c.blocks) > 32 {
		var oldest structureBlockKey
		used, found := ^uint64(0), false
		for position, candidate := range c.blocks {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(c.blocks, oldest)
	}
	c.mu.Unlock()
	return result, err
}

func (g *OverworldGenerator) generateStructureBlocks(start StructureStart) (*structureBlockResult, error) {
	if isOverworldMineshaftStructure(start.Structure) {
		return g.generateMineshaftStructureBlocks(start)
	}
	order, err := OriginalStructureDecorationOrder(start.Structure)
	if err != nil {
		return nil, err
	}
	cx, cz := start.Chunk.X, start.Chunk.Z
	box, _ := start.BoundingBox()
	ruinedPortal := strings.HasPrefix(start.Structure, "minecraft:ruined_portal")
	if ruinedPortal {
		// RuinedPortalPiece is invoked once, in its bounding-box center's
		// chunk. Rotation can put that chunk outside the start owner.
		cx = (box.MinX + (box.MaxX-box.MinX+1)/2) >> 4
		cz = (box.MinZ + (box.MaxZ-box.MinZ+1)/2) >> 4
	}
	bases, chunks, err := g.decorationBases(cx, cz)
	if err != nil {
		return nil, err
	}
	var preceding []StructureStart
	for _, neighbor := range g.decorationStructures(bases, cx, cz) {
		previousOrder, err := OriginalStructureDecorationOrder(neighbor.Structure)
		if err != nil {
			return nil, err
		}
		if previousOrder.Step < order.Step {
			preceding = append(preceding, neighbor)
		}
	}
	region, _, err := g.decorateRegion(bases, chunks, cx, cz, int(order.Step), preceding)
	if err != nil {
		return nil, err
	}
	result := &structureBlockResult{start: start.Clone(), blocks: make(map[FeaturePosition]SurfaceBlockState), postProcessing: make(map[FeaturePosition]bool), blockEntities: make(map[FeaturePosition]featureBlockEntity)}
	world := &structureRecordingWorld{region: region, result: result}
	minX, maxX, minZ, maxZ := box.MinX>>4, box.MaxX>>4, box.MinZ>>4, box.MaxZ>>4
	if ruinedPortal {
		minX, maxX, minZ, maxZ = cx, cx, cz, cz
	}
	// Scattered pieces are anchored at their start's minimum X/Z, so the
	// owner's clip comes first. Reuse each private piece across all clips.
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			decorationSeed := OverworldDecorationSeed(g.seed, x, z)
			random := NewWorldgenRandom(decorationSeed + int64(order.Index) + 10000*int64(order.Step))
			worldRandom := NewStructureRegionRandom(g.seed, x, z)
			for i := range result.start.Pieces {
				piece := &result.start.Pieces[i]
				var placed StructurePiecePostprocessResult
				switch start.Structure {
				case "minecraft:igloo":
					placed, err = PostProcessIglooBlocks(world, piece, random, x, z)
				case "minecraft:swamp_hut":
					placed, err = PostProcessSwampHutBlocks(world, piece, x, z)
				case "minecraft:jungle_pyramid":
					placed, err = PostProcessJungleTempleBlocks(world, piece, random, x, z)
				case "minecraft:desert_pyramid":
					placed, err = PostProcessDesertPyramidBlocks(world, piece, random, worldRandom, g.seed, x, z)
				case "minecraft:buried_treasure":
					placed, err = PostProcessBuriedTreasureBlocks(world, piece, random, x, z)
				case "minecraft:shipwreck", "minecraft:shipwreck_beached":
					placed, err = PostProcessShipwreckBlocks(world, piece, random, x, z)
				case "minecraft:ocean_ruin_cold", "minecraft:ocean_ruin_warm":
					placed, err = PostProcessOceanRuinBlocks(world, piece, random, g.seed, x, z)
				default:
					if ruinedPortal {
						placed, err = PostProcessRuinedPortalBlocks(world, piece, random, x, z)
					} else {
						return nil, fmt.Errorf("unsupported structure block placer %s", start.Structure)
					}
				}
				if err != nil {
					return nil, fmt.Errorf("postprocess %s in %d,%d: %w", start.Structure, x, z, err)
				}
				if start.Structure == "minecraft:swamp_hut" {
					for i := range placed.PendingEntities {
						request := &placed.PendingEntities[i]
						request.Data, err = FinalizeSwampHutEntity(request.ID, worldRandom, g.structureFullMoon())
						if err != nil {
							return nil, err
						}
					}
				}
				if start.Structure == "minecraft:ocean_ruin_cold" || start.Structure == "minecraft:ocean_ruin_warm" {
					for actorIndex := range placed.PendingEntities {
						request := &placed.PendingEntities[actorIndex]
						if request.ID != "minecraft:drowned" {
							continue
						}
						request.SavedFlag = fmt.Sprintf("OceanRuinDrowned:%d,%d:%d:%d", start.Chunk.X, start.Chunk.Z, i, actorIndex)
						finalized, err := FinalizeDrownedStructureEntity(*request, worldRandom, NewLegacyRandom(g.seed^int64(positionSeed(request.Position.X, request.Position.Y, request.Position.Z))), false)
						if err != nil {
							return nil, err
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
				}
				result.pending = append(result.pending, placed.PendingEntities...)
			}
		}
	}
	return result, nil
}

func (g *OverworldGenerator) decorationStructures(bases map[featureChunkPosition]*overworldBase, cx, cz int32) []StructureStart {
	if g.structures == nil {
		return nil
	}
	seen := make(map[structureBlockKey]bool)
	var result []StructureStart
	for _, base := range bases {
		for _, start := range base.starts {
			key := structureBlockKey{start.Structure, start.Chunk.X, start.Chunk.Z}
			if seen[key] {
				continue
			}
			seen[key] = true
			box, valid := structureWriteBox(start)
			if valid && box.MaxX >= (cx-1)*16 && box.MinX <= (cx+1)*16+15 && box.MaxZ >= (cz-1)*16 && box.MinZ <= (cz+1)*16+15 {
				result = append(result, start)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		oa, _ := OriginalStructureDecorationOrder(a.Structure)
		ob, _ := OriginalStructureDecorationOrder(b.Structure)
		if oa.Step != ob.Step {
			return oa.Step < ob.Step
		}
		if oa.Index != ob.Index {
			return oa.Index < ob.Index
		}
		if a.Chunk.X != b.Chunk.X {
			return a.Chunk.X < b.Chunk.X
		}
		return a.Chunk.Z < b.Chunk.Z
	})
	return result
}

func (g *OverworldGenerator) applyStructureStep(region *OverworldFeatureRegion, starts []StructureStart, cx, cz, step int32) error {
	for _, start := range starts {
		order, err := OriginalStructureDecorationOrder(start.Structure)
		if err != nil {
			return err
		}
		if order.Step != step {
			continue
		}
		box, _ := structureWriteBox(start)
		if box.MaxX < cx*16 || box.MinX > cx*16+15 || box.MaxZ < cz*16 || box.MinZ > cz*16+15 {
			continue
		}
		var cached *structureBlockResult
		jigsaw := isOverworldJigsawStructure(start.Structure)
		if jigsaw {
			cached, err = g.jigsawStructureBlocksFor(start, cx, cz)
		} else if start.Structure == "minecraft:stronghold" {
			cached, err = g.strongholdStructureBlocksFor(start, cx, cz)
		} else {
			cached, err = g.structureBlocksFor(start)
		}
		if err != nil {
			return err
		}
		for p, state := range cached.blocks {
			if jigsaw || p.X>>4 == cx && p.Z>>4 == cz {
				if err := region.SetBlock(p, state); err != nil {
					return err
				}
			}
		}
		for p := range cached.postProcessing {
			if jigsaw || p.X>>4 == cx && p.Z>>4 == cz {
				if err := region.MarkPostProcessing(p); err != nil {
					return err
				}
			}
		}
		for p, entity := range cached.blockEntities {
			if jigsaw || p.X>>4 == cx && p.Z>>4 == cz {
				if err := region.SetBlockEntity(p, entity.block, entity.data); err != nil {
					return err
				}
			}
		}
		for _, entity := range cached.pending {
			if jigsaw || entity.Position.X>>4 == cx && entity.Position.Z>>4 == cz {
				if err := region.AddPendingEntity(entity); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// References use the original template box. Runtime replay also includes
// template-edge shape/fluid updates and portal spread, drips and vines.
func structureWriteBox(start StructureStart) (StructureBoundingBox, bool) {
	box, ok := start.BoundingBox()
	if !ok {
		return box, false
	}
	if strings.HasPrefix(start.Structure, "minecraft:ruined_portal") {
		x, z := box.MinX+(box.MaxX-box.MinX+1)/2, box.MinZ+(box.MaxZ-box.MinZ+1)/2
		box.MinX, box.MaxX = min(box.MinX-1, x-14), max(box.MaxX+1, x+14)
		box.MinZ, box.MaxZ = min(box.MinZ-1, z-14), max(box.MaxZ+1, z+14)
		box.MinY, box.MaxY = box.MinY-12, box.MaxY+1
	} else if start.Structure == "minecraft:igloo" || strings.HasPrefix(start.Structure, "minecraft:shipwreck") || strings.HasPrefix(start.Structure, "minecraft:ocean_ruin_") {
		box.MinX--
		box.MaxX++
		box.MinZ--
		box.MaxZ++
	}
	if isOverworldJigsawStructure(start.Structure) {
		for _, piece := range start.Pieces {
			positions, _ := jigsawFeatureOrigins(piece)
			for _, position := range positions {
				// The ancient-city sculk cursor has a15-block horizontal
				// travel limit; feature placement ignores the template clip.
				box.MinX, box.MaxX = min(box.MinX, position.X-15), max(box.MaxX, position.X+15)
				box.MinZ, box.MaxZ = min(box.MinZ, position.Z-15), max(box.MaxZ, position.Z+15)
			}
		}
	}
	return box, true
}

func (g *OverworldGenerator) generatedStructureNBT(cx, cz int32, starts []StructureStart) (nbt.Object, error) {
	var owned []StructureStart
	for _, start := range starts {
		if start.Chunk.X == cx && start.Chunk.Z == cz {
			cached, err := g.structureBlocksFor(start)
			if err != nil {
				return nil, err
			}
			if start.Structure == "minecraft:stronghold" {
				owned = append(owned, g.strongholdSavedStart(cached.start))
			} else {
				owned = append(owned, cached.start.Clone())
			}
		}
	}
	return StructureNBTForChunk(cx, cz, owned, BuildStructureReferences(cx, cz, starts))
}
