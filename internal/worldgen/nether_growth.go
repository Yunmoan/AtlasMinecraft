package worldgen

import (
	"encoding/json"
	"fmt"
)

// Nether vines and glowstone use BlockState.isAir, not replaceability: water,
// vegetation and other nonair states block these generation attempts.
func netherWartSupport(s SurfaceBlockState) bool {
	return s.Name == "minecraft:netherrack" || s.Name == "minecraft:nether_wart_block"
}

func compileTwistingVinesPlan(def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	var c struct {
		Width  int32 `json:"spread_width"`
		Height int32 `json:"spread_height"`
		Length int32 `json:"max_height"`
	}
	if err := json.Unmarshal(def.Config, &c); err != nil {
		return nil, err
	}
	if c.Width <= 0 || c.Width > 64 || c.Height <= 0 || c.Height > 64 || c.Length <= 0 || c.Length > 256 {
		return nil, fmt.Errorf("invalid twisting vines spread")
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		valid := func(p FeaturePosition) bool {
			if !surfaceAir(e.block(p)) {
				return false
			}
			s := e.block(treeOffset(p, 0, -1, 0))
			return s.Name == "minecraft:netherrack" || s.Name == "minecraft:warped_nylium" || s.Name == "minecraft:warped_wart_block"
		}
		if !valid(origin) {
			return false, e.err
		}
		for i := int32(0); i < c.Width*c.Width; i++ {
			p := treeOffset(origin, randomBetweenInclusive(r, -c.Width, c.Width), randomBetweenInclusive(r, -c.Height, c.Height), randomBetweenInclusive(r, -c.Width, c.Width))
			found := false
			for {
				p.Y--
				if p.Y < region.MinBuildHeight() || p.Y >= region.MaxBuildHeight() {
					break
				}
				if !surfaceAir(e.block(p)) {
					p.Y++
					found = true
					break
				}
			}
			if !found || !valid(p) {
				continue
			}
			height := randomBetweenInclusive(r, 1, c.Length)
			if r.NextIntBound(6) == 0 {
				height *= 2
			}
			if r.NextIntBound(5) == 0 {
				height = 1
			}
			for j := int32(1); j <= height; j++ {
				if surfaceAir(e.block(p)) {
					if j == height || !surfaceAir(e.block(treeOffset(p, 0, 1, 0))) {
						e.set(p, SurfaceBlockState{Name: "minecraft:twisting_vines", Properties: map[string]string{"age": fmt.Sprint(randomBetweenInclusive(r, 17, 25))}})
						break
					}
					e.set(p, SurfaceBlockState{Name: "minecraft:twisting_vines_plant"})
				}
				p.Y++
			}
		}
		return e.err == nil, e.err
	}, nil
}

func compileWeepingVinesPlan(ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !surfaceAir(e.block(origin)) || !netherWartSupport(e.block(treeOffset(origin, 0, 1, 0))) {
			return false, e.err
		}
		e.set(origin, SurfaceBlockState{Name: "minecraft:nether_wart_block"})
		for i := 0; i < 200; i++ {
			p := treeOffset(origin, r.NextIntBound(6)-r.NextIntBound(6), r.NextIntBound(2)-r.NextIntBound(5), r.NextIntBound(6)-r.NextIntBound(6))
			if !surfaceAir(e.block(p)) {
				continue
			}
			neighbors := 0
			for _, d := range featureDirections {
				if netherWartSupport(e.block(treeOffset(p, d.X, d.Y, d.Z))) {
					neighbors++
				}
				if neighbors > 1 {
					break
				}
			}
			if neighbors == 1 {
				e.set(p, SurfaceBlockState{Name: "minecraft:nether_wart_block"})
			}
		}
		for i := 0; i < 100; i++ {
			p := treeOffset(origin, r.NextIntBound(8)-r.NextIntBound(8), r.NextIntBound(2)-r.NextIntBound(7), r.NextIntBound(8)-r.NextIntBound(8))
			if !surfaceAir(e.block(p)) || !netherWartSupport(e.block(treeOffset(p, 0, 1, 0))) {
				continue
			}
			height := randomBetweenInclusive(r, 1, 8)
			if r.NextIntBound(6) == 0 {
				height *= 2
			}
			if r.NextIntBound(5) == 0 {
				height = 1
			}
			placeWeepingVineColumn(&e, r, p, height, 17, 25)
		}
		return e.err == nil, e.err
	}, nil
}

func compileGlowstonePlan(ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		e := featureEditor{world: region}
		if !surfaceAir(e.block(origin)) {
			return false, e.err
		}
		s := e.block(treeOffset(origin, 0, 1, 0))
		if s.Name != "minecraft:netherrack" && s.Name != "minecraft:basalt" && s.Name != "minecraft:blackstone" {
			return false, e.err
		}
		e.set(origin, SurfaceBlockState{Name: "minecraft:glowstone"})
		for i := 0; i < 1500; i++ {
			p := treeOffset(origin, r.NextIntBound(8)-r.NextIntBound(8), -r.NextIntBound(12), r.NextIntBound(8)-r.NextIntBound(8))
			if !surfaceAir(e.block(p)) {
				continue
			}
			neighbors := 0
			for _, d := range featureDirections {
				if e.block(treeOffset(p, d.X, d.Y, d.Z)).Name == "minecraft:glowstone" {
					neighbors++
				}
				if neighbors > 1 {
					break
				}
			}
			if neighbors == 1 {
				e.set(p, SurfaceBlockState{Name: "minecraft:glowstone"})
			}
		}
		return e.err == nil, e.err
	}, nil
}
