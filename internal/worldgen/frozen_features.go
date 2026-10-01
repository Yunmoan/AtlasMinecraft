package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"fmt"
	"math"
	"sort"
)

type featureEditor struct {
	world FeatureBlockWorld
	err   error
}

func (e *featureEditor) block(p FeaturePosition) SurfaceBlockState {
	if e.err != nil {
		return SurfaceBlockState{}
	}
	state, err := e.world.BlockAt(p)
	e.err = err
	return state
}
func (e *featureEditor) set(p FeaturePosition, state SurfaceBlockState) {
	if e.err == nil {
		e.err = setFeatureBlock(e.world, p, state)
	}
}

func PlaceIceSpikeFeature(world FeatureBlockWorld, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
	if world == nil || random == nil {
		return false, fmt.Errorf("ice spike requires a world and random source")
	}
	e := featureEditor{world: world}
	for surfaceAir(e.block(origin)) && origin.Y > -62 {
		origin.Y--
	}
	if e.err != nil {
		return false, e.err
	}
	if e.block(origin).Name != "minecraft:snow_block" {
		return false, e.err
	}
	origin.Y += random.NextIntBound(4)
	height := random.NextIntBound(4) + 7
	radius := height/4 + random.NextIntBound(2)
	if radius > 1 && random.NextIntBound(60) == 0 {
		origin.Y += 10 + random.NextIntBound(30)
	}
	packed := SurfaceBlockState{Name: "minecraft:packed_ice"}
	replace := func(state SurfaceBlockState) bool {
		return surfaceAir(state) || forestRockDirt[state.Name] || state.Name == "minecraft:snow_block" || state.Name == "minecraft:ice"
	}
	for y := int32(0); y < height; y++ {
		r := (1 - float32(y)/float32(height)) * float32(radius)
		extent := int32(math.Ceil(float64(r)))
		for x := -extent; x <= extent; x++ {
			for z := -extent; z <= extent; z++ {
				dx, dz := float32(abs32(x))-0.25, float32(abs32(z))-0.25
				if (x != 0 || z != 0) && dx*dx+dz*dz > r*r || (x == -extent || x == extent || z == -extent || z == extent) && random.NextFloat() > 0.75 {
					continue
				}
				p := treeOffset(origin, x, y, z)
				if replace(e.block(p)) {
					e.set(p, packed)
				}
				if y != 0 && extent > 1 {
					p = treeOffset(origin, x, -y, z)
					if replace(e.block(p)) {
						e.set(p, packed)
					}
				}
			}
		}
	}
	rootRadius := max(int32(0), min(int32(1), radius-1))
	for x := -rootRadius; x <= rootRadius; x++ {
		for z := -rootRadius; z <= rootRadius; z++ {
			p := treeOffset(origin, x, -1, z)
			remaining := int32(50)
			if abs32(x) == 1 && abs32(z) == 1 {
				remaining = random.NextIntBound(5)
			}
			for p.Y > 50 {
				state := e.block(p)
				if !replace(state) && state.Name != "minecraft:packed_ice" {
					break
				}
				e.set(p, packed)
				p.Y--
				remaining--
				if remaining <= 0 {
					p.Y -= random.NextIntBound(5) + 1
					remaining = random.NextIntBound(5)
				}
			}
		}
	}
	return e.err == nil, e.err
}

// SnowLayerBlock's tagged exceptions precede its full upper-face check.
func snowSupported(state SurfaceBlockState) (bool, error) {
	if vegetationTags["snow_layer_cannot_survive_on"][state.Name] {
		return false, nil
	}
	if vegetationTags["snow_layer_can_survive_on"][state.Name] || state.Name == "minecraft:snow" && state.Properties["layers"] == "8" {
		return true, nil
	}
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return false, err
	}
	rule, err := blockrules.Lookup(id)
	if err != nil {
		return false, err
	}
	// Snow tests the collision face, rather than SupportType.FULL. Leaves
	// expose a full collision face while deliberately failing sturdy support.
	var faces []blockrules.Box
	xs, zs := []float64{0, 1}, []float64{0, 1}
	for _, box := range rule.Boxes {
		if box[4] != 1 {
			continue
		}
		faces = append(faces, box)
		xs = append(xs, max(0, box[0]), min(1, box[3]))
		zs = append(zs, max(0, box[2]), min(1, box[5]))
	}
	sort.Float64s(xs)
	sort.Float64s(zs)
	for i := 1; i < len(xs); i++ {
		for j := 1; j < len(zs); j++ {
			if xs[i] <= xs[i-1] || zs[j] <= zs[j-1] {
				continue
			}
			x, z := (xs[i]+xs[i-1])/2, (zs[j]+zs[j-1])/2
			covered := false
			for _, box := range faces {
				if x >= box[0] && x <= box[3] && z >= box[2] && z <= box[5] {
					covered = true
					break
				}
			}
			if !covered {
				return false, nil
			}
		}
	}
	return true, nil
}

func PlaceSnowAndFreezeFeature(region *OverworldFeatureRegion, origin FeaturePosition) (bool, error) {
	if region == nil || region.biomeAtQuart == nil {
		return false, fmt.Errorf("snow and freezing require a region with a biome source")
	}
	e := featureEditor{world: region}
	for x := int32(0); x < 16; x++ {
		for z := int32(0); z < 16; z++ {
			bx, bz := origin.X+x, origin.Z+z
			y, err := region.HeightmapAt("MOTION_BLOCKING", bx, bz)
			if err != nil {
				return false, err
			}
			qx, qy, qz := BiomeQuartAtBlock(region.biomeSeed, bx, y, bz)
			biome, err := region.biomeAtQuart(qx, qy, qz)
			if err != nil {
				return false, err
			}
			below, p := FeaturePosition{bx, y - 1, bz}, FeaturePosition{bx, y, bz}
			cold, err := BiomeColdAt(biome, bx, below.Y, bz)
			if err != nil {
				return false, err
			}
			state := e.block(below)
			if cold && region.InWorld(below) && state.Name == "minecraft:water" && hasSourceWater(state) {
				e.set(below, SurfaceBlockState{Name: "minecraft:ice"})
			}
			cold, err = BiomeColdAt(biome, bx, p.Y, bz)
			if err != nil {
				return false, err
			}
			state = e.block(p)
			if cold && region.InWorld(p) && (surfaceAir(state) || state.Name == "minecraft:snow") {
				ground := e.block(below)
				supported, err := snowSupported(ground)
				if err != nil {
					return false, err
				}
				if supported {
					e.set(p, SurfaceBlockState{Name: "minecraft:snow", Properties: map[string]string{"layers": "1"}})
					if _, ok := ground.Properties["snowy"]; ok || ground.Name == "minecraft:grass_block" || ground.Name == "minecraft:podzol" || ground.Name == "minecraft:mycelium" {
						e.set(below, stateWithProperty(ground, "snowy", "true"))
					}
				}
			}
		}
	}
	return e.err == nil, e.err
}
