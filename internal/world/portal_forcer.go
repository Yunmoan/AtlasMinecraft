package world

import (
	"atlas/internal/blockrules"
	"atlas/internal/worldgen"
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

type portalChunkQuery struct {
	cx, cz int32
	dense  bool
	result chan portalChunkSnapshot
}
type portalChunkSnapshot struct {
	states  []uint32
	motion  [256]int32
	portals []worldgen.FeaturePosition
	err     error
}

func (r *Region) portalChunkSnapshot(q portalChunkQuery) portalChunkSnapshot {
	ch := r.chunks[chunkKey{q.cx, q.cz}]
	if ch == nil {
		return portalChunkSnapshot{err: ErrChunkNotLoaded}
	}
	ch.lastAccess = time.Now()
	var s portalChunkSnapshot
	portalX, _ := BlockStateID("minecraft:nether_portal", map[string]string{"axis": "x"})
	portalZ, _ := BlockStateID("minecraft:nether_portal", map[string]string{"axis": "z"})
	if q.dense {
		s.states = make([]uint32, int(ch.maxY-ch.minY)*256)
		for i := range s.motion {
			s.motion[i] = ch.minY
		}
		for y := ch.minY; y < ch.maxY; y++ {
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					id := ch.GetBlock(x, z, y)
					s.states[int(y-ch.minY)*256+z*16+x] = id
					if stateFlag(motionFlags, id) {
						s.motion[z*16+x] = y + 1
					}
				}
			}
		}
	}
	for section, blocks := range ch.sect {
		if blocks == nil {
			continue
		}
		hasPortal := false
		if blocks.bits > 8 {
			hasPortal = true
		} else {
			for _, id := range blocks.palette {
				if id == portalX || id == portalZ {
					hasPortal = true
					break
				}
			}
		}
		if !hasPortal {
			continue
		}
		for i := 0; i < 4096; i++ {
			id := blocks.get(i)
			if id == portalX || id == portalZ {
				s.portals = append(s.portals, worldgen.FeaturePosition{X: q.cx*16 + int32(i&15), Y: ch.minY + int32(section*16+i/256), Z: q.cz*16 + int32((i/16)&15)})
			}
		}
	}
	return s
}

func (w *World) readPortalChunk(ctx context.Context, cx, cz int32, dense bool) (portalChunkSnapshot, error) {
	for {
		if err := w.ensureChunk(ctx, cx, cz); err != nil {
			return portalChunkSnapshot{}, err
		}
		result := make(chan portalChunkSnapshot, 1)
		if err := w.post(ctx, cx, cz, portalChunkQuery{cx: cx, cz: cz, dense: dense, result: result}); err != nil {
			return portalChunkSnapshot{}, err
		}
		select {
		case s := <-result:
			if s.err == ErrChunkNotLoaded {
				continue
			}
			return s, s.err
		case <-ctx.Done():
			return portalChunkSnapshot{}, ctx.Err()
		case <-w.done:
			return portalChunkSnapshot{}, ErrWorldClosed
		}
	}
}

type portalSnapshotWorld struct {
	w      *World
	ctx    context.Context
	chunks map[chunkKey]portalChunkSnapshot
}

func (a *portalSnapshotWorld) chunk(x, z int32) (portalChunkSnapshot, error) {
	k := chunkKey{x >> 4, z >> 4}
	if s, ok := a.chunks[k]; ok {
		return s, nil
	}
	s, err := a.w.readPortalChunk(a.ctx, k.x, k.z, true)
	if err == nil {
		a.chunks[k] = s
	}
	return s, err
}
func (a *portalSnapshotWorld) InWorld(p worldgen.FeaturePosition) bool {
	return (portalBlockWorld{a.w}).InWorld(p)
}
func (a *portalSnapshotWorld) BlockAt(p worldgen.FeaturePosition) (worldgen.SurfaceBlockState, error) {
	if !a.InWorld(p) {
		return worldgen.SurfaceBlockState{Name: "minecraft:air"}, nil
	}
	s, err := a.chunk(p.X, p.Z)
	if err != nil {
		return worldgen.SurfaceBlockState{}, err
	}
	id := s.states[int(p.Y-a.w.MinBuildHeight())*256+int(p.Z&15)*16+int(p.X&15)]
	state, err := BlockStateByID(id)
	return worldgen.SurfaceBlockState{Name: state.Name, Properties: state.Properties}, err
}
func (a *portalSnapshotWorld) HeightmapAt(_ string, x, z int32) (int32, error) {
	s, err := a.chunk(x, z)
	return s.motion[int(z&15)*16+int(x&15)], err
}
func (a *portalSnapshotWorld) SetBlock(p worldgen.FeaturePosition, s worldgen.SurfaceBlockState) error {
	if err := (portalBlockWorld{a.w}).SetBlock(p, s); err != nil {
		return err
	}
	if ch, ok := a.chunks[chunkKey{p.X >> 4, p.Z >> 4}]; ok {
		id, err := BlockStateID(s.Name, s.Properties)
		if err != nil {
			return err
		}
		ch.states[int(p.Y-a.w.MinBuildHeight())*256+int(p.Z&15)*16+int(p.X&15)] = id
	}
	return nil
}

// NetherPortalCandidates reconstructs the portal POI view from authoritative
// chunk blocks. Queries run in actors and palette skips avoid scanning empty
// sections. Chunk generation uses the existing shared worker queue.
func (w *World) NetherPortalCandidates(ctx context.Context, target BlockPosition) ([]worldgen.FeaturePosition, error) {
	radius := int32(128)
	if w.DimensionName() == "minecraft:the_nether" {
		radius = 16
	}
	var keys []chunkKey
	for cz := (target.Z - radius) >> 4; cz <= (target.Z+radius)>>4; cz++ {
		for cx := (target.X - radius) >> 4; cx <= (target.X+radius)>>4; cx++ {
			keys = append(keys, chunkKey{cx, cz})
		}
	}
	results := make([]portalChunkSnapshot, len(keys))
	jobs := make(chan int, len(keys))
	for i := range keys {
		jobs <- i
	}
	close(jobs)
	var wg sync.WaitGroup
	for worker := 0; worker < min(8, len(keys)); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				s, err := w.readPortalChunk(ctx, keys[i].x, keys[i].z, false)
				s.err = err
				results[i] = s
			}
		}()
	}
	wg.Wait()
	var positions []worldgen.FeaturePosition
	for _, s := range results {
		if s.err != nil {
			return nil, s.err
		}
		positions = append(positions, s.portals...)
	}
	return positions, nil
}

// FindOrCreateNetherPortal preserves PortalForcer search/create geometry.
// Player arrival and cooldown are separate from the returned block rectangle.
func (w *World) FindOrCreateNetherPortal(ctx context.Context, target BlockPosition, axis string) (worldgen.PortalRectangle, error) {
	if w.DimensionName() == "minecraft:the_end" {
		return worldgen.PortalRectangle{}, fmt.Errorf("End dimension has no Nether portal")
	}
	border, err := w.PortalBorder()
	if err != nil {
		return worldgen.PortalRectangle{}, err
	}
	candidates, err := w.NetherPortalCandidates(ctx, target)
	if err != nil {
		return worldgen.PortalRectangle{}, err
	}
	a := &portalSnapshotWorld{w: w, ctx: ctx, chunks: make(map[chunkKey]portalChunkSnapshot)}
	p := worldgen.FeaturePosition{X: target.X, Y: target.Y, Z: target.Z}
	closest, found, err := worldgen.FindClosestNetherPortal(a, candidates, p, w.DimensionName() == "minecraft:the_nether", border)
	if err != nil {
		return worldgen.PortalRectangle{}, err
	}
	if found {
		return worldgen.NetherPortalRectangle(a, closest)
	}
	logical := w.MaxBuildHeight() - w.MinBuildHeight()
	if w.DimensionName() == "minecraft:the_nether" {
		logical = 128
	}
	rect, found, err := worldgen.CreateNetherPortal(a, p, axis, w.MinBuildHeight(), w.MaxBuildHeight(), logical, border)
	if err == nil && !found {
		err = fmt.Errorf("unable to create Nether portal within logical build height")
	}
	return rect, err
}

func (a *portalSnapshotWorld) collisionBoxes(bounds blockrules.Box, border worldgen.PortalBorder) ([]blockrules.Box, error) {
	var boxes []blockrules.Box
	for y := int32(math.Floor(bounds[1])) - 1; y <= int32(math.Floor(bounds[4]))+1; y++ {
		if y < a.w.MinBuildHeight() || y >= a.w.MaxBuildHeight() {
			continue
		}
		for x := int32(math.Floor(bounds[0])) - 1; x <= int32(math.Floor(bounds[3]))+1; x++ {
			for z := int32(math.Floor(bounds[2])) - 1; z <= int32(math.Floor(bounds[5]))+1; z++ {
				p := worldgen.FeaturePosition{X: x, Y: y, Z: z}
				s, err := a.BlockAt(p)
				if err != nil {
					return nil, err
				}
				id, err := BlockStateID(s.Name, s.Properties)
				if err != nil {
					return nil, err
				}
				rules, err := blockrules.Lookup(id)
				if err != nil {
					return nil, err
				}
				offset, zero := worldgen.BlockOffset(s.Name, p), worldgen.BlockOffset(s.Name, worldgen.FeaturePosition{})
				for _, box := range rules.Boxes {
					for axis, coordinate := range []float64{float64(x), float64(y), float64(z)} {
						box[axis] += coordinate + offset[axis] - zero[axis]
						box[axis+3] += coordinate + offset[axis] - zero[axis]
					}
					if !overlaps(box, bounds) {
						continue
					}
					const borderEpsilon = float64(float32(1e-5))
					if box[0] < border.MinX || box[2] < border.MinZ || box[3]-borderEpsilon >= border.MaxX || box[5]-borderEpsilon >= border.MaxZ {
						continue
					}
					boxes = append(boxes, box)
				}
			}
		}
	}
	return boxes, nil
}

// NetherPortalArrival computes a living player's destination through the
// actual source rectangle, 8:1 search target and collision-adjusted exit.
func (w *World) NetherPortalArrival(ctx context.Context, source *World, sourcePortal BlockPosition, position PlayerPosition, velocity [3]float64, width, height float32) (PlayerPosition, [3]float64, error) {
	if source == nil || width <= 0 || height <= 0 || width > 16 || height > 16 || math.IsNaN(float64(width)) || math.IsNaN(float64(height)) {
		return PlayerPosition{}, velocity, fmt.Errorf("invalid portal entity dimensions")
	}
	target, err := w.NetherPortalSearchTarget(source.DimensionName(), position.X, position.Y, position.Z)
	if err != nil {
		return PlayerPosition{}, velocity, err
	}
	a := &portalSnapshotWorld{w: source, ctx: ctx, chunks: make(map[chunkKey]portalChunkSnapshot)}
	p := worldgen.FeaturePosition{X: sourcePortal.X, Y: sourcePortal.Y, Z: sourcePortal.Z}
	s, err := a.BlockAt(p)
	if err != nil {
		return PlayerPosition{}, velocity, err
	}
	sourceAxis := "x"
	relative := [3]float64{.5, 0, 0}
	if s.Name == "minecraft:nether_portal" {
		sourceAxis = s.Properties["axis"]
		rect, err := worldgen.NetherPortalRectangle(a, p)
		if err != nil {
			return PlayerPosition{}, velocity, err
		}
		relative = worldgen.PortalRelativePosition(rect, [3]float64{position.X, position.Y, position.Z}, width, height, true)
	}
	rect, err := w.FindOrCreateNetherPortal(ctx, target, sourceAxis)
	if err != nil {
		return PlayerPosition{}, velocity, err
	}
	exit, yaw, speed := worldgen.PortalExitPosition(rect, sourceAxis, relative, width, height, position.Yaw, velocity)
	if width <= 4 && height <= 4 {
		destination := &portalSnapshotWorld{w: w, ctx: ctx, chunks: make(map[chunkKey]portalChunkSnapshot)}
		border, err := w.PortalBorder()
		if err != nil {
			return PlayerPosition{}, velocity, err
		}
		boxes, err := destination.collisionBoxes(worldgen.PortalCollisionSearchBounds(exit, width, height), border)
		if err != nil {
			return PlayerPosition{}, velocity, err
		}
		exit = worldgen.FindCollisionFreePortalPosition(exit, width, height, boxes)
	}
	position.X, position.Y, position.Z = exit[0], exit[1], exit[2]
	position.Yaw, position.Dimension = yaw, w.DimensionName()
	return position, speed, nil
}
