package worldgen

import (
	"atlas/internal/blockstate"
	"encoding/json"
)

// Matches BlockPos.withinManhattan: shells, ascending x/y, positive z then its
// mirror. Feature edits can change later eligibility, so this order matters.
func visitFeatureManhattan(origin FeaturePosition, extentX, extentY, extentZ, limit int32, visit func(FeaturePosition)) {
	for distance := int32(0); distance <= min(limit, extentX+extentY+extentZ); distance++ {
		for x := -min(extentX, distance); x <= min(extentX, distance); x++ {
			yLimit := min(extentY, distance-abs32(x))
			for y := -yLimit; y <= yLimit; y++ {
				z := distance - abs32(x) - abs32(y)
				if z > extentZ {
					continue
				}
				visit(treeOffset(origin, x, y, z))
				if z != 0 {
					visit(treeOffset(origin, x, y, -z))
				}
			}
		}
	}
}

func netherStructureBlock(s SurfaceBlockState) bool {
	switch s.Name {
	case "minecraft:bedrock", "minecraft:nether_bricks", "minecraft:nether_brick_fence", "minecraft:nether_brick_stairs", "minecraft:nether_wart", "minecraft:chest", "minecraft:spawner":
		return true
	}
	return false
}

func compileDeltaPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Contents SurfaceBlockState `json:"contents"`
		Rim      SurfaceBlockState `json:"rim"`
		Size     json.RawMessage   `json:"size"`
		RimSize  json.RawMessage   `json:"rim_size"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	for _, s := range []SurfaceBlockState{c.Contents, c.Rim} {
		if _, err := blockstate.BlockStateID(s.Name, s.Properties); err != nil {
			return nil, err
		}
	}
	size, err := decodeFeatureIntProvider(c.Size, 0, 16)
	if err != nil {
		return nil, err
	}
	rimSize, err := decodeFeatureIntProvider(c.RimSize, 0, 16)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		valid := func(p FeaturePosition) bool {
			s := e.block(p)
			if s.Name == c.Contents.Name || netherStructureBlock(s) {
				return false
			}
			for d := range featureDirections {
				if surfaceAir(e.block(directionPosition(p, d))) != (d == 1) {
					return false
				}
			}
			return true
		}
		rimEnabled := r.NextDouble() < 0.9
		var rimX, rimZ int32
		if rimEnabled {
			rimX, rimZ = rimSize(r), rimSize(r)
		}
		rimEnabled = rimEnabled && rimX != 0 && rimZ != 0
		x, z := size(r), size(r)
		placed := false
		visitFeatureManhattan(origin, x, 0, z, max(x, z), func(p FeaturePosition) {
			if !valid(p) {
				return
			}
			if rimEnabled {
				e.set(p, c.Rim)
				placed = true
			}
			q := treeOffset(p, rimX, 0, rimZ)
			if valid(q) {
				e.set(q, c.Contents)
				placed = true
			}
		})
		return placed, e.err
	}, nil
}

func compileReplaceBlobsPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Target SurfaceBlockState `json:"target"`
		State  SurfaceBlockState `json:"state"`
		Radius json.RawMessage   `json:"radius"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	for _, s := range []SurfaceBlockState{c.Target, c.State} {
		if _, err := blockstate.BlockStateID(s.Name, s.Properties); err != nil {
			return nil, err
		}
	}
	radius, err := decodeFeatureIntProvider(c.Radius, 0, 12)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		origin.Y = max(region.MinBuildHeight()+1, min(region.MaxBuildHeight()-1, origin.Y))
		found := false
		for origin.Y > region.MinBuildHeight()+1 {
			if e.block(origin).Name == c.Target.Name {
				found = true
				break
			}
			origin.Y--
		}
		if !found {
			return false, e.err
		}
		x, y, z := radius(r), radius(r), radius(r)
		placed := false
		visitFeatureManhattan(origin, x, y, z, max(x, max(y, z)), func(p FeaturePosition) {
			if e.block(p).Name == c.Target.Name {
				e.set(p, c.State)
				placed = true
			}
		})
		return placed, e.err
	}, nil
}

func compileBasaltPillarPlan(ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !surfaceAir(e.block(origin)) || surfaceAir(e.block(treeOffset(origin, 0, 1, 0))) {
			return false, e.err
		}
		basalt := SurfaceBlockState{Name: "minecraft:basalt"}
		flags := [4]bool{true, true, true, true}
		p := origin
		for surfaceAir(e.block(p)) {
			if p.Y < region.MinBuildHeight() || p.Y >= region.MaxBuildHeight() {
				return true, e.err
			}
			e.set(p, basalt)
			for i := range flags {
				if flags[i] {
					flags[i] = r.NextIntBound(10) != 0
					if flags[i] {
						e.set(directionPosition(p, i+2), basalt)
					}
				}
			}
			p.Y--
		}
		p.Y++
		for d := 2; d < 6; d++ {
			if r.NextBoolean() {
				e.set(directionPosition(p, d), basalt)
			}
		}
		p.Y--
		for x := int32(-3); x <= 3; x++ {
			for z := int32(-3); z <= 3; z++ {
				if r.NextIntBound(10) >= 10-abs32(x)*abs32(z) {
					continue
				}
				q := treeOffset(p, x, 0, z)
				for remaining := 3; remaining > 0 && surfaceAir(e.block(treeOffset(q, 0, -1, 0))); remaining-- {
					q.Y--
				}
				if !surfaceAir(e.block(treeOffset(q, 0, -1, 0))) {
					e.set(q, basalt)
				}
			}
		}
		return true, e.err
	}, nil
}

func basaltColumnExcluded(s SurfaceBlockState) bool {
	return netherStructureBlock(s) || s.Name == "minecraft:lava" || s.Name == "minecraft:magma_block" || s.Name == "minecraft:soul_sand"
}

func compileBasaltColumnsPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Reach  json.RawMessage `json:"reach"`
		Height json.RawMessage `json:"height"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	reach, err := decodeFeatureIntProvider(c.Reach, 0, 3)
	if err != nil {
		return nil, err
	}
	height, err := decodeFeatureIntProvider(c.Height, 1, 10)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		replaceable := func(p FeaturePosition) bool {
			s := e.block(p)
			return surfaceAir(s) || s.Name == "minecraft:lava" && p.Y <= region.GenerationSeaLevel()
		}
		valid := func(p FeaturePosition) bool {
			if !replaceable(p) {
				return false
			}
			s := e.block(treeOffset(p, 0, -1, 0))
			return !surfaceAir(s) && !basaltColumnExcluded(s)
		}
		if !valid(origin) {
			return false, e.err
		}
		h := height(r)
		clustered := r.NextFloat() < 0.9
		radius, count := min(h, int32(8)), 15
		if clustered {
			radius, count = min(h, int32(5)), 50
		}
		placed := false
		for i := 0; i < count; i++ {
			p := treeOffset(origin, r.NextIntBound(radius*2+1)-radius, r.NextIntBound(1), r.NextIntBound(radius*2+1)-radius)
			maxHeight := h - abs32(p.X-origin.X) - abs32(p.Z-origin.Z)
			if maxHeight < 0 {
				continue
			}
			footprint := reach(r)
			for z := -footprint; z <= footprint; z++ {
				for x := -footprint; x <= footprint; x++ {
					q := treeOffset(p, x, 0, z)
					distance := abs32(x) + abs32(z)
					found := false
					if replaceable(q) {
						for n := distance; q.Y > region.MinBuildHeight()+1 && n > 0; n-- {
							if valid(q) {
								found = true
								break
							}
							q.Y--
						}
					} else {
						for n := distance; q.Y < region.MaxBuildHeight() && n > 0; n-- {
							s := e.block(q)
							if basaltColumnExcluded(s) {
								break
							}
							if surfaceAir(s) {
								found = true
								break
							}
							q.Y++
						}
					}
					if !found {
						continue
					}
					for n := maxHeight - distance/2; n >= 0; n-- {
						if replaceable(q) {
							e.set(q, SurfaceBlockState{Name: "minecraft:basalt"})
							placed = true
						} else if e.block(q).Name != "minecraft:basalt" {
							break
						}
						q.Y++
					}
				}
			}
		}
		return placed, e.err
	}, nil
}
