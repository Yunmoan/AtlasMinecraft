package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"encoding/json"
)

func compileBambooPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config struct {
		Probability float32 `json:"probability"`
	}
	if err := json.Unmarshal(def.Config, &config); err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		p := origin
		if !surfaceAir(e.block(p)) {
			return false, e.err
		}
		if vegetationTags["bamboo_plantable_on"][e.block(treeOffset(p, 0, -1, 0)).Name] {
			height := r.NextIntBound(12) + 5
			if r.NextFloat() < config.Probability {
				radius := r.NextIntBound(4) + 1
				for x := origin.X - radius; x <= origin.X+radius; x++ {
					for z := origin.Z - radius; z <= origin.Z+radius; z++ {
						if (x-origin.X)*(x-origin.X)+(z-origin.Z)*(z-origin.Z) > radius*radius {
							continue
						}
						y, err := region.HeightmapAt("WORLD_SURFACE", x, z)
						if err != nil {
							return false, err
						}
						q := FeaturePosition{x, y - 1, z}
						if forestRockDirt[e.block(q).Name] {
							e.set(q, SurfaceBlockState{Name: "minecraft:podzol"})
						}
					}
				}
			}
			state := SurfaceBlockState{Name: "minecraft:bamboo", Properties: map[string]string{"age": "1", "leaves": "none", "stage": "0"}}
			for i := int32(0); i < height && surfaceAir(e.block(p)); i++ {
				e.set(p, state)
				p.Y++
			}
			if p.Y-origin.Y >= 3 {
				e.set(p, stateWithProperty(stateWithProperty(state, "leaves", "large"), "stage", "1"))
				p.Y--
				e.set(p, stateWithProperty(state, "leaves", "large"))
				p.Y--
				e.set(p, stateWithProperty(state, "leaves", "small"))
			}
		}
		return true, e.err
	}, nil
}

func compileBlockPilePlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var config struct {
		Provider json.RawMessage `json:"state_provider"`
	}
	if err := json.Unmarshal(def.Config, &config); err != nil {
		return nil, err
	}
	provider, err := decodeFeatureStateProvider(config.Provider, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		if origin.Y < -59 {
			return false, nil
		}
		e := featureEditor{world: region}
		rx, rz := 2+r.NextIntBound(2), 2+r.NextIntBound(2)
		for z := origin.Z - rz; z <= origin.Z+rz; z++ {
			for y := origin.Y; y <= origin.Y+1; y++ {
				for x := origin.X - rx; x <= origin.X+rx; x++ {
					dx, dz := origin.X-x, origin.Z-z
					if float32(dx*dx+dz*dz) > float32(r.NextFloat()*float32(10))-float32(r.NextFloat()*float32(6)) && float64(r.NextFloat()) >= 0.031 {
						continue
					}
					p := FeaturePosition{x, y, z}
					if !surfaceAir(e.block(p)) {
						continue
					}
					below := e.block(treeOffset(p, 0, -1, 0))
					supported := false
					if below.Name == "minecraft:dirt_path" {
						supported = r.NextBoolean()
					} else {
						id, err := blockstate.BlockStateID(below.Name, below.Properties)
						if err != nil {
							return false, err
						}
						rule, err := blockrules.Lookup(id)
						if err != nil {
							return false, err
						}
						supported = rule.FaceSturdy[1]
					}
					if supported {
						e.set(p, provider(r, p))
					}
				}
			}
		}
		return true, e.err
	}, nil
}
