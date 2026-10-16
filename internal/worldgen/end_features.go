package worldgen

import (
	"atlas/internal/nbt"
	"encoding/json"
	"fmt"
	"math"
)

func compileEndIslandPlan(ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		radius := float32(r.NextIntBound(3)) + 4
		for y := int32(0); radius > 0.5; y-- {
			for x := int32(math.Floor(float64(-radius))); x <= int32(math.Ceil(float64(radius))); x++ {
				for z := int32(math.Floor(float64(-radius))); z <= int32(math.Ceil(float64(radius))); z++ {
					if float32(x*x+z*z) <= (radius+1)*(radius+1) {
						e.set(treeOffset(origin, x, y, z), SurfaceBlockState{Name: "minecraft:end_stone"})
					}
				}
			}
			radius -= float32(r.NextIntBound(2)) + 0.5
		}
		return true, e.err
	}, nil
}

func compileEndPlatformPlan(ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		return true, PlaceEndPlatform(region, origin)
	}, nil
}

func compileEndGatewayPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Exit  []int32 `json:"exit"`
		Exact bool    `json:"exact"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Exit != nil && len(c.Exit) != 3 {
		return nil, fmt.Errorf("invalid End gateway exit position")
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		for z := int32(-1); z <= 1; z++ {
			for y := int32(-2); y <= 2; y++ {
				for x := int32(-1); x <= 1; x++ {
					name := "minecraft:air"
					if x == 0 && y == 0 && z == 0 {
						name = "minecraft:end_gateway"
					} else if y != 0 {
						if abs32(y) == 2 {
							if x == 0 && z == 0 {
								name = "minecraft:bedrock"
							}
						} else if x == 0 || z == 0 {
							name = "minecraft:bedrock"
						}
					}
					e.set(treeOffset(origin, x, y, z), SurfaceBlockState{Name: name})
				}
			}
		}
		if e.err == nil {
			data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:end_gateway"}, "Age": {Type: nbt.Long, Value: int64(0)}}
			if c.Exit != nil {
				data["exit_portal"] = nbt.Tag{Type: nbt.IntArray, Value: append([]int32(nil), c.Exit...)}
				if c.Exact {
					data["ExactTeleport"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
				}
			}
			e.err = region.SetBlockEntity(origin, "minecraft:end_gateway", data)
		}
		return true, e.err
	}, nil
}

func chorusPlantState(e *featureEditor, p FeaturePosition) SurfaceBlockState {
	properties := map[string]string{}
	for d, name := range directionNames {
		s := e.block(directionPosition(p, d))
		connected := s.Name == "minecraft:chorus_plant" || s.Name == "minecraft:chorus_flower" || d == 0 && s.Name == "minecraft:end_stone"
		properties[name] = fmt.Sprint(connected)
	}
	return SurfaceBlockState{Name: "minecraft:chorus_plant", Properties: properties}
}

func compileChorusPlantPlan(ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !surfaceAir(e.block(origin)) || e.block(treeOffset(origin, 0, -1, 0)).Name != "minecraft:end_stone" {
			return false, e.err
		}
		e.set(origin, chorusPlantState(&e, origin))
		placeChorusBranch(&e, r, origin, origin, 8, 0)
		return true, e.err
	}, nil
}

func chorusNeighborsClear(e *featureEditor, p FeaturePosition, except int) bool {
	for _, d := range []int{2, 5, 3, 4} {
		if d != except && !surfaceAir(e.block(directionPosition(p, d))) {
			return false
		}
	}
	return true
}

func placeChorusBranch(e *featureEditor, r *WorldgenRandom, start, origin FeaturePosition, radius, depth int32) {
	height := r.NextIntBound(4) + 1
	if depth == 0 {
		height++
	}
	for y := int32(0); y < height; y++ {
		p := treeOffset(start, 0, y+1, 0)
		if !chorusNeighborsClear(e, p, -1) {
			return
		}
		e.set(p, chorusPlantState(e, p))
		below := treeOffset(p, 0, -1, 0)
		e.set(below, chorusPlantState(e, below))
	}
	branched := false
	if depth < 4 {
		attempts := r.NextIntBound(4)
		if depth == 0 {
			attempts++
		}
		for i := int32(0); i < attempts; i++ {
			d := []int{2, 5, 3, 4}[r.NextIntBound(4)]
			p := directionPosition(treeOffset(start, 0, height, 0), d)
			if abs32(p.X-origin.X) >= radius || abs32(p.Z-origin.Z) >= radius || !surfaceAir(e.block(p)) || !surfaceAir(e.block(treeOffset(p, 0, -1, 0))) || !chorusNeighborsClear(e, p, d^1) {
				continue
			}
			branched = true
			e.set(p, chorusPlantState(e, p))
			back := directionPosition(p, d^1)
			e.set(back, chorusPlantState(e, back))
			placeChorusBranch(e, r, p, origin, radius, depth+1)
		}
	}
	if !branched {
		e.set(treeOffset(start, 0, height, 0), SurfaceBlockState{Name: "minecraft:chorus_flower", Properties: map[string]string{"age": "5"}})
	}
}

type endSpike struct {
	CenterX int32 `json:"centerX"`
	CenterZ int32 `json:"centerZ"`
	Radius  int32 `json:"radius"`
	Height  int32 `json:"height"`
	Guarded bool  `json:"guarded"`
}

func endSpikesForSeed(seed int64) []endSpike {
	key := NewLegacyRandom(seed).NextLong() & 65535
	r := NewLegacyRandom(int64(key))
	order := [10]int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := len(order); i > 1; i-- {
		j := r.NextIntBound(int32(i))
		order[i-1], order[j] = order[j], order[i-1]
	}
	spikes := make([]endSpike, 10)
	for i, n := range order {
		angle := float64(0.3141592653589793 * float64(i))
		angle = float64(-float64(math.Pi) + angle)
		angle = float64(2 * angle)
		s, c := featureSinCos(angle)
		spikes[i] = endSpike{int32(math.Floor(42 * c)), int32(math.Floor(42 * s)), 2 + n/3, 76 + n*3, n == 1 || n == 2}
	}
	return spikes
}

func compileEndSpikePlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Invulnerable bool       `json:"crystal_invulnerable"`
		Spikes       []endSpike `json:"spikes"`
		Beam         []int32    `json:"crystal_beam_target"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Beam != nil && len(c.Beam) != 3 {
		return nil, fmt.Errorf("invalid End crystal beam target")
	}
	for _, s := range c.Spikes {
		if s.Radius < 0 || s.Radius > 128 {
			return nil, fmt.Errorf("invalid End spike radius %d", s.Radius)
		}
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		spikes := c.Spikes
		if len(spikes) == 0 {
			spikes = endSpikesForSeed(region.WorldSeed())
		}
		for _, s := range spikes {
			if s.CenterX>>4 != origin.X>>4 || s.CenterZ>>4 != origin.Z>>4 {
				continue
			}
			for z := s.CenterZ - s.Radius; z <= s.CenterZ+s.Radius; z++ {
				for y := region.MinBuildHeight(); y <= s.Height+10; y++ {
					for x := s.CenterX - s.Radius; x <= s.CenterX+s.Radius; x++ {
						dx, dz := x-s.CenterX, z-s.CenterZ
						if dx*dx+dz*dz <= s.Radius*s.Radius+1 && y < s.Height {
							e.set(FeaturePosition{x, y, z}, SurfaceBlockState{Name: "minecraft:obsidian"})
						} else if y > 65 {
							e.set(FeaturePosition{x, y, z}, SurfaceBlockState{Name: "minecraft:air"})
						}
					}
				}
			}
			if s.Guarded {
				for x := int32(-2); x <= 2; x++ {
					for z := int32(-2); z <= 2; z++ {
						for y := int32(0); y <= 3; y++ {
							edgeX, edgeZ, top := abs32(x) == 2, abs32(z) == 2, y == 3
							if !edgeX && !edgeZ && !top {
								continue
							}
							connectX, connectZ := edgeX || top, edgeZ || top
							properties := map[string]string{"north": fmt.Sprint(connectX && z != -2), "south": fmt.Sprint(connectX && z != 2), "west": fmt.Sprint(connectZ && x != -2), "east": fmt.Sprint(connectZ && x != 2), "waterlogged": "false"}
							e.set(FeaturePosition{s.CenterX + x, s.Height + y, s.CenterZ + z}, SurfaceBlockState{Name: "minecraft:iron_bars", Properties: properties})
						}
					}
				}
			}
			yaw := r.NextFloat() * 360
			p := FeaturePosition{s.CenterX, s.Height + 1, s.CenterZ}
			data := nbt.Object{
				"id":           {Type: nbt.String, Value: "minecraft:end_crystal"},
				"Pos":          {Type: nbt.List, ListType: nbt.Double, Value: []nbt.Tag{{Type: nbt.Double, Value: float64(s.CenterX) + 0.5}, {Type: nbt.Double, Value: float64(s.Height + 1)}, {Type: nbt.Double, Value: float64(s.CenterZ) + 0.5}}},
				"Rotation":     {Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{{Type: nbt.Float, Value: yaw}, {Type: nbt.Float, Value: float32(0)}}},
				"Invulnerable": {Type: nbt.Byte, Value: int8(0)}, "ShowBottom": {Type: nbt.Byte, Value: int8(1)},
			}
			if c.Invulnerable {
				data["Invulnerable"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
			}
			if c.Beam != nil {
				data["beam_target"] = nbt.Tag{Type: nbt.IntArray, Value: append([]int32(nil), c.Beam...)}
			}
			if e.err == nil {
				e.err = region.AddPendingEntity(PendingStructureEntity{ID: "minecraft:end_crystal", Position: p, Data: data})
			}
			e.set(treeOffset(p, 0, -1, 0), SurfaceBlockState{Name: "minecraft:bedrock"})
			e.set(p, SurfaceBlockState{Name: "minecraft:fire"})
		}
		return true, e.err
	}, nil
}
