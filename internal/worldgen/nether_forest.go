package worldgen

import (
	"atlas/internal/blockstate"
	"encoding/json"
	"fmt"
)

func netherNylium(s SurfaceBlockState) bool {
	return s.Name == "minecraft:crimson_nylium" || s.Name == "minecraft:warped_nylium"
}

func validateNetherForestState(s SurfaceBlockState) error {
	switch s.Name {
	case "minecraft:crimson_roots", "minecraft:warped_roots", "minecraft:crimson_fungus", "minecraft:warped_fungus", "minecraft:nether_sprouts":
		return nil
	default:
		return fmt.Errorf("unsupported Nether forest vegetation %s", s.Name)
	}
}

func compileNetherForestPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Provider json.RawMessage `json:"state_provider"`
		Width    int32           `json:"spread_width"`
		Height   int32           `json:"spread_height"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Width <= 0 || c.Width > 64 || c.Height <= 0 || c.Height > 64 {
		return nil, fmt.Errorf("invalid Nether forest vegetation spread")
	}
	provider, err := decodeFeatureStateProvider(c.Provider, validateNetherForestState)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !netherNylium(e.block(treeOffset(origin, 0, -1, 0))) || origin.Y < region.MinBuildHeight()+1 || origin.Y+1 >= region.MaxBuildHeight() {
			return false, e.err
		}
		count := 0
		for i := int32(0); i < c.Width*c.Width; i++ {
			p := treeOffset(origin, r.NextIntBound(c.Width)-r.NextIntBound(c.Width), r.NextIntBound(c.Height)-r.NextIntBound(c.Height), r.NextIntBound(c.Width)-r.NextIntBound(c.Width))
			s := provider(r, p)
			if !surfaceAir(e.block(p)) || p.Y <= region.MinBuildHeight() {
				continue
			}
			below := e.block(treeOffset(p, 0, -1, 0))
			survives := netherNylium(below) || below.Name == "minecraft:soul_soil" || PlantGroundSurvives(s.Name, below.Name)
			if s.Name == "minecraft:crimson_fungus" || s.Name == "minecraft:warped_fungus" {
				survives = survives || below.Name == "minecraft:mycelium"
			}
			if survives {
				e.set(p, s)
				count++
			}
		}
		return count > 0, e.err
	}, nil
}

type hugeFungusConfiguration struct {
	Base        SurfaceBlockState `json:"valid_base_block"`
	Stem        SurfaceBlockState `json:"stem_state"`
	Hat         SurfaceBlockState `json:"hat_state"`
	Decoration  SurfaceBlockState `json:"decor_state"`
	Replaceable json.RawMessage   `json:"replaceable_blocks"`
	Planted     bool              `json:"planted"`
	replaceable featurePredicate
}

func compileHugeFungusPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c hugeFungusConfiguration
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	for _, state := range []SurfaceBlockState{c.Base, c.Stem, c.Hat, c.Decoration} {
		if _, err := blockstate.BlockStateID(state.Name, state.Properties); err != nil {
			return nil, err
		}
	}
	var err error
	c.replaceable, err = compileFeaturePredicate(c.Replaceable)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		return placeHugeFungus(region, r, origin, c, region.GenerationDepth())
	}, nil
}

func placeHugeFungus(world FeatureBlockWorld, r *WorldgenRandom, origin FeaturePosition, c hugeFungusConfiguration, generationDepth int32) (bool, error) {
	e := featureEditor{world: world}
	if e.block(treeOffset(origin, 0, -1, 0)).Name != c.Base.Name {
		return false, e.err
	}
	height := r.NextIntBound(10) + 4
	if r.NextIntBound(12) == 0 {
		height *= 2
	}
	if !c.Planted && origin.Y+height+1 >= generationDepth {
		return false, e.err
	}
	thick := !c.Planted && r.NextFloat() < 0.06
	e.set(origin, SurfaceBlockState{Name: "minecraft:air"})
	replaceable := func(p FeaturePosition, usePredicate bool) bool {
		if multifaceRules(&e, e.block(p)).Replaceable {
			return true
		}
		if usePredicate && e.err == nil {
			ok, err := c.replaceable(world, p)
			e.err = err
			return ok
		}
		return false
	}
	radius := int32(0)
	if thick {
		radius = 1
	}
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			corner := thick && abs32(x) == radius && abs32(z) == radius
			for y := int32(0); y < height; y++ {
				p := treeOffset(origin, x, y, z)
				if !replaceable(p, true) {
					continue
				}
				if c.Planted && !surfaceAir(e.block(treeOffset(p, 0, -1, 0))) {
					e.set(p, SurfaceBlockState{Name: "minecraft:air"})
				}
				if c.Planted || !corner || r.NextFloat() < 0.1 {
					e.set(p, c.Stem)
				}
			}
		}
	}
	crimson := c.Hat.Name == "minecraft:nether_wart_block"
	hatHeight := min(r.NextIntBound(1+height/3)+5, height)
	bottom := height - hatHeight
	for y := bottom; y <= height; y++ {
		radius := int32(1)
		if y < height-r.NextIntBound(3) {
			radius = 2
		}
		if hatHeight > 8 && y < bottom+4 {
			radius = 3
		}
		if thick {
			radius++
		}
		for x := -radius; x <= radius; x++ {
			for z := -radius; z <= radius; z++ {
				edgeX, edgeZ := abs32(x) == radius, abs32(z) == radius
				inside := !edgeX && !edgeZ && y != height
				corner := edgeX && edgeZ
				p := treeOffset(origin, x, y, z)
				if !replaceable(p, false) {
					continue
				}
				if c.Planted && !surfaceAir(e.block(treeOffset(p, 0, -1, 0))) {
					e.set(p, SurfaceBlockState{Name: "minecraft:air"})
				}
				if y < bottom+3 {
					if inside {
						continue
					}
					if e.block(treeOffset(p, 0, -1, 0)).Name == c.Hat.Name {
						e.set(p, c.Hat)
					} else if float64(r.NextFloat()) < 0.15 {
						e.set(p, c.Hat)
						if crimson && r.NextIntBound(11) == 0 {
							fungusWeepingVines(&e, r, p)
						}
					}
					continue
				}
				decorChance, hatChance, vineChance := float32(0.0005), float32(0.98), float32(0)
				if inside {
					decorChance, hatChance = 0.1, 0.2
					if crimson {
						vineChance = 0.1
					}
				} else if corner {
					decorChance, hatChance = 0.01, 0.7
					if crimson {
						vineChance = 0.083
					}
				} else if crimson {
					vineChance = 0.07
				}
				if r.NextFloat() < decorChance {
					e.set(p, c.Decoration)
				} else if r.NextFloat() < hatChance {
					e.set(p, c.Hat)
					if r.NextFloat() < vineChance {
						fungusWeepingVines(&e, r, p)
					}
				}
			}
		}
	}
	return e.err == nil, e.err
}

func fungusWeepingVines(e *featureEditor, r *WorldgenRandom, origin FeaturePosition) {
	p := treeOffset(origin, 0, -1, 0)
	if !surfaceAir(e.block(p)) {
		return
	}
	height := r.NextIntBound(5) + 1
	if r.NextIntBound(7) == 0 {
		height *= 2
	}
	placeWeepingVineColumn(e, r, p, height, 23, 25)
}

func placeWeepingVineColumn(e *featureEditor, r *WorldgenRandom, p FeaturePosition, height, minAge, maxAge int32) {
	for i := int32(0); i <= height; i++ {
		if surfaceAir(e.block(p)) {
			if i == height || !surfaceAir(e.block(treeOffset(p, 0, -1, 0))) {
				e.set(p, SurfaceBlockState{Name: "minecraft:weeping_vines", Properties: map[string]string{"age": fmt.Sprint(r.NextIntBound(maxAge-minAge+1) + minAge)}})
				break
			}
			e.set(p, SurfaceBlockState{Name: "minecraft:weeping_vines_plant"})
		}
		p.Y--
	}
}
