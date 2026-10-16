package worldgen

// Processors and canonical structures often read the same preceding stage in
// adjacent clips. Cache only its center chunk, not its entire 5x5 read region.
// Keys include the stage because later structures must see earlier decoration.
type structureChunkKey struct {
	x, z, step int32
}

type structureChunkSnapshot struct {
	base    *overworldBase
	overlay *featureChunkOverlay
}

type structureChunkEntry struct {
	ready    chan struct{}
	snapshot *structureChunkSnapshot
	err      error
	used     uint64
	complete bool
}

const structureChunkCacheLimit = 128

func (g *OverworldGenerator) structureChunkBefore(cx, cz, step int32) (*structureChunkSnapshot, error) {
	key := structureChunkKey{cx, cz, step}
	g.precedingMu.Lock()
	g.precedingClock++
	if entry := g.preceding[key]; entry != nil {
		entry.used = g.precedingClock
		g.precedingMu.Unlock()
		<-entry.ready
		return entry.snapshot, entry.err
	}
	if g.preceding == nil {
		g.preceding = make(map[structureChunkKey]*structureChunkEntry)
	}
	entry := &structureChunkEntry{ready: make(chan struct{}), used: g.precedingClock}
	g.preceding[key] = entry
	g.precedingMu.Unlock()
	region, err := g.structureRegionBefore(cx, cz, step)
	var snapshot *structureChunkSnapshot
	if err == nil {
		position := featureChunkPosition{cx, cz}
		snapshot = &structureChunkSnapshot{base: region.biomeBases[position], overlay: region.overlays[position]}
	}
	g.precedingMu.Lock()
	entry.snapshot, entry.err, entry.complete = snapshot, err, true
	close(entry.ready)
	if err != nil {
		delete(g.preceding, key)
	}
	for len(g.preceding) > structureChunkCacheLimit {
		var oldest structureChunkKey
		used, found := ^uint64(0), false
		for position, candidate := range g.preceding {
			if candidate.complete && candidate.used < used {
				oldest, used, found = position, candidate.used, true
			}
		}
		if !found {
			break
		}
		delete(g.preceding, oldest)
	}
	g.precedingMu.Unlock()
	return snapshot, err
}

// Terrain and biome arrays remain immutable. Copy the small write overlay so
// concurrent structures cannot change a cached preceding stage or each other.
func (region *OverworldFeatureRegion) installStructureChunk(key featureChunkPosition, snapshot *structureChunkSnapshot) {
	region.chunks[key] = snapshot.base.terrain
	region.borrowed[key] = true
	region.biomeBases[key] = snapshot.base
	if snapshot.overlay == nil {
		return
	}
	if region.overlays == nil {
		region.overlays = make(map[featureChunkPosition]*featureChunkOverlay)
	}
	overlay := &featureChunkOverlay{blocks: make(map[int]SurfaceBlockState, len(snapshot.overlay.blocks))}
	for index, state := range snapshot.overlay.blocks {
		overlay.blocks[index] = cloneSurfaceState(state)
	}
	if snapshot.overlay.postProcessing != nil {
		overlay.postProcessing = make(map[int]bool, len(snapshot.overlay.postProcessing))
		for index, marked := range snapshot.overlay.postProcessing {
			overlay.postProcessing[index] = marked
		}
	}
	region.overlays[key] = overlay
}
