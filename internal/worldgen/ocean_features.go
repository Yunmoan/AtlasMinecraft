package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed ocean_tags.json
var oceanTagData []byte

var oceanTags = func() map[string][]string {
	var tags map[string][]string
	if err := json.Unmarshal(oceanTagData, &tags); err != nil {
		panic(err)
	}
	return tags
}()

func oceanSupport(e *featureEditor, p FeaturePosition, pickle bool) bool {
	s := e.block(treeOffset(p, 0, -1, 0))
	id, err := blockstate.BlockStateID(s.Name, s.Properties)
	if err != nil {
		e.err = err
		return false
	}
	rule, err := blockrules.Lookup(id)
	if err != nil {
		e.err = err
		return false
	}
	if pickle {
		if rule.FaceSturdy[1] {
			return true
		}
		for _, b := range rule.Boxes {
			if b[4] >= 1 && b[0] < b[3] && b[2] < b[5] {
				return true
			}
		}
		return false
	}
	return s.Name != "minecraft:magma_block" && rule.FaceSturdy[1]
}

func kelpSupport(e *featureEditor, p FeaturePosition) bool {
	s := e.block(treeOffset(p, 0, -1, 0))
	if s.Name == "minecraft:kelp" || s.Name == "minecraft:kelp_plant" {
		return true
	}
	return oceanSupport(e, p, false)
}

func compileOceanPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	switch def.Type {
	case "minecraft:seagrass":
		var config struct {
			Probability float32 `json:"probability"`
		}
		if err := json.Unmarshal(def.Config, &config); err != nil {
			return nil, err
		}
		if config.Probability < 0 || config.Probability > 1 {
			return nil, fmt.Errorf("invalid seagrass probability")
		}
		return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
			e := featureEditor{world: region}
			dx, dz := r.NextIntBound(8)-r.NextIntBound(8), r.NextIntBound(8)-r.NextIntBound(8)
			p := treeOffset(origin, dx, 0, dz)
			var err error
			p.Y, err = region.OceanFloorHeightAt(p.X, p.Z)
			if err != nil {
				return false, err
			}
			if e.block(p).Name != "minecraft:water" {
				return false, e.err
			}
			tall := r.NextDouble() < float64(config.Probability)
			if !oceanSupport(&e, p, false) {
				return false, e.err
			}
			// The Java feature reports success even when the upper half is blocked.
			if tall {
				if hasSourceWater(e.block(p)) && e.block(treeOffset(p, 0, 1, 0)).Name == "minecraft:water" {
					e.set(p, SurfaceBlockState{Name: "minecraft:tall_seagrass", Properties: map[string]string{"half": "lower"}})
					e.set(treeOffset(p, 0, 1, 0), SurfaceBlockState{Name: "minecraft:tall_seagrass", Properties: map[string]string{"half": "upper"}})
				} else if !hasSourceWater(e.block(p)) {
					return false, e.err
				}
			} else {
				e.set(p, SurfaceBlockState{Name: "minecraft:seagrass"})
			}
			return true, e.err
		}, nil
	case "minecraft:sea_pickle":
		var config struct {
			Count json.RawMessage `json:"count"`
		}
		if err := json.Unmarshal(def.Config, &config); err != nil {
			return nil, err
		}
		count, err := decodeFeatureIntProvider(config.Count, 0, 256)
		if err != nil {
			return nil, err
		}
		return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
			e := featureEditor{world: region}
			placed := false
			for n := count(r); n > 0; n-- {
				dx, dz := r.NextIntBound(8)-r.NextIntBound(8), r.NextIntBound(8)-r.NextIntBound(8)
				p := treeOffset(origin, dx, 0, dz)
				var err error
				p.Y, err = region.OceanFloorHeightAt(p.X, p.Z)
				if err != nil {
					return false, err
				}
				pickles := fmt.Sprint(r.NextIntBound(4) + 1)
				if e.block(p).Name == "minecraft:water" && oceanSupport(&e, p, true) {
					e.set(p, SurfaceBlockState{Name: "minecraft:sea_pickle", Properties: map[string]string{"pickles": pickles, "waterlogged": "true"}})
					placed = true
				}
			}
			return placed, e.err
		}, nil
	case "minecraft:kelp":
		return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
			e := featureEditor{world: region}
			p := origin
			var err error
			p.Y, err = region.OceanFloorHeightAt(p.X, p.Z)
			if err != nil {
				return false, err
			}
			if e.block(p).Name != "minecraft:water" {
				return false, e.err
			}
			height := 1 + r.NextIntBound(10)
			placed := false
			for i := int32(0); i <= height; i++ {
				if e.block(p).Name == "minecraft:water" && e.block(treeOffset(p, 0, 1, 0)).Name == "minecraft:water" && kelpSupport(&e, p) {
					if i == height {
						e.set(p, SurfaceBlockState{Name: "minecraft:kelp", Properties: map[string]string{"age": fmt.Sprint(r.NextIntBound(4) + 20)}})
						placed = true
					} else {
						e.set(p, SurfaceBlockState{Name: "minecraft:kelp_plant"})
					}
				} else if i > 0 {
					p.Y--
					if kelpSupport(&e, p) && e.block(treeOffset(p, 0, -1, 0)).Name != "minecraft:kelp" {
						e.set(p, SurfaceBlockState{Name: "minecraft:kelp", Properties: map[string]string{"age": fmt.Sprint(r.NextIntBound(4) + 20)}})
						placed = true
					}
					break
				}
				p.Y++
			}
			return placed, e.err
		}, nil
	case "minecraft:coral_tree", "minecraft:coral_claw", "minecraft:coral_mushroom":
		return func(region *OverworldFeatureRegion, r *WorldgenRandom, p FeaturePosition) (bool, error) {
			return placeCoralFeature(region, r, p, def.Type)
		}, nil
	}
	return nil, fmt.Errorf("unsupported ocean feature %s", def.Type)
}

func coralTagContains(tag, name string) bool {
	for _, value := range oceanTags[tag] {
		if value == name {
			return true
		}
	}
	return false
}

func randomCoral(r *WorldgenRandom, tag string) SurfaceBlockState {
	values := oceanTags[tag]
	return SurfaceBlockState{Name: values[r.NextIntBound(int32(len(values)))]}
}

func placeCoralBlock(e *featureEditor, r *WorldgenRandom, p FeaturePosition, state SurfaceBlockState) bool {
	s := e.block(p)
	above := treeOffset(p, 0, 1, 0)
	if (s.Name != "minecraft:water" && !coralTagContains("corals", s.Name)) || e.block(above).Name != "minecraft:water" {
		return false
	}
	e.set(p, state)
	if r.NextFloat() < 0.25 {
		e.set(above, randomCoral(r, "corals"))
	} else if r.NextFloat() < 0.05 {
		e.set(above, SurfaceBlockState{Name: "minecraft:sea_pickle", Properties: map[string]string{"pickles": fmt.Sprint(r.NextIntBound(4) + 1), "waterlogged": "true"}})
	}
	for _, d := range []int{2, 5, 3, 4} {
		if r.NextFloat() < 0.2 {
			q := directionPosition(p, d)
			if e.block(q).Name == "minecraft:water" {
				e.set(q, stateWithProperty(randomCoral(r, "wall_corals"), "facing", directionNames[d]))
			}
		}
	}
	return true
}

func placeCoralFeature(world FeatureBlockWorld, r *WorldgenRandom, origin FeaturePosition, kind string) (bool, error) {
	e := featureEditor{world: world}
	state := randomCoral(r, "coral_blocks")
	switch kind {
	case "minecraft:coral_tree":
		p := origin
		height := 1 + r.NextIntBound(3)
		for i := int32(0); i < height; i++ {
			if !placeCoralBlock(&e, r, p, state) {
				return true, e.err
			}
			p.Y++
		}
		start := p
		branches := 2 + r.NextIntBound(3)
		for _, d := range shuffledDirections(r, []int{2, 5, 3, 4})[:branches] {
			p = directionPosition(start, d)
			length := 2 + r.NextIntBound(5)
			shift := 0
			for i := int32(0); i < length && placeCoralBlock(&e, r, p, state); i++ {
				p.Y++
				if i > 0 {
					shift++
					if shift < 2 || r.NextFloat() >= 0.25 {
						continue
					}
				}
				p = directionPosition(p, d)
				shift = 0
			}
		}
	case "minecraft:coral_claw":
		if !placeCoralBlock(&e, r, origin, state) {
			return false, e.err
		}
		horizontal := []int{2, 5, 3, 4}
		idx := int(r.NextIntBound(4))
		forward := horizontal[idx]
		branches := 2 + r.NextIntBound(2)
		for _, d := range shuffledDirections(r, []int{forward, horizontal[(idx+1)%4], horizontal[(idx+3)%4]})[:branches] {
			p := directionPosition(origin, d)
			first := 1 + r.NextIntBound(2)
			travel := forward
			var length int32
			if d == forward {
				length = 2 + r.NextIntBound(3)
			} else {
				p.Y++
				travel = []int{d, 1}[r.NextIntBound(2)]
				length = 3 + r.NextIntBound(3)
			}
			for i := int32(0); i < first && placeCoralBlock(&e, r, p, state); i++ {
				p = directionPosition(p, travel)
			}
			p = directionPosition(p, travel^1)
			p.Y++
			for i := int32(0); i < length; i++ {
				p = directionPosition(p, forward)
				if !placeCoralBlock(&e, r, p, state) {
					break
				}
				if r.NextFloat() < 0.25 {
					p.Y++
				}
			}
		}
	case "minecraft:coral_mushroom":
		height, width, depth, down := 3+r.NextIntBound(3), 3+r.NextIntBound(3), 3+r.NextIntBound(3), 1+r.NextIntBound(3)
		for x := int32(0); x <= width; x++ {
			for y := int32(0); y <= height; y++ {
				for z := int32(0); z <= depth; z++ {
					bx, by, bz := x == 0 || x == width, y == 0 || y == height, z == 0 || z == depth
					if (!bx || !by) && (!bz || !by) && (!bx || !bz) && (bx || by || bz) && r.NextFloat() >= 0.1 {
						placeCoralBlock(&e, r, treeOffset(origin, x, y-down, z), state)
					}
				}
			}
		}
	}
	return true, e.err
}
