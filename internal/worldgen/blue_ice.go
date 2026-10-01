package worldgen

import "fmt"

var blueIceDirections = [...]FeaturePosition{
	{Y: -1}, {Y: 1}, {Z: -1}, {Z: 1}, {X: -1}, {X: 1},
}

func PlaceBlueIce(world FeatureBlockWorld, random *WorldgenRandom, origin FeaturePosition) (bool, error) {
	if world == nil || random == nil {
		return false, fmt.Errorf("blue ice requires a world and random source")
	}
	if origin.Y > 319 {
		return false, nil
	}
	state, err := world.BlockAt(origin)
	if err != nil {
		return false, err
	}
	below, err := world.BlockAt(FeaturePosition{origin.X, origin.Y - 1, origin.Z})
	if err != nil {
		return false, err
	}
	if state.Name != "minecraft:water" && below.Name != "minecraft:water" {
		return false, nil
	}
	connected := false
	for _, direction := range blueIceDirections[1:] {
		neighbor, err := world.BlockAt(FeaturePosition{origin.X + direction.X,
			origin.Y + direction.Y, origin.Z + direction.Z})
		if err != nil {
			return false, err
		}
		if neighbor.Name == "minecraft:packed_ice" {
			connected = true
			break
		}
	}
	if !connected {
		return false, nil
	}
	if err := world.SetBlock(origin, SurfaceBlockState{Name: "minecraft:blue_ice"}); err != nil {
		return false, err
	}
	for range 200 {
		offsetY := random.NextIntBound(5) - random.NextIntBound(6)
		spread := int32(3)
		if offsetY < 2 {
			spread += offsetY / 2
		}
		if spread < 1 {
			continue
		}
		candidate := FeaturePosition{
			X: origin.X + random.NextIntBound(spread) - random.NextIntBound(spread),
			Y: origin.Y + offsetY,
			Z: origin.Z + random.NextIntBound(spread) - random.NextIntBound(spread),
		}
		current, err := world.BlockAt(candidate)
		if err != nil {
			return false, err
		}
		if !surfaceAir(current) && current.Name != "minecraft:water" &&
			current.Name != "minecraft:packed_ice" && current.Name != "minecraft:ice" {
			continue
		}
		for _, direction := range blueIceDirections {
			neighbor, err := world.BlockAt(FeaturePosition{candidate.X + direction.X,
				candidate.Y + direction.Y, candidate.Z + direction.Z})
			if err != nil {
				return false, err
			}
			if neighbor.Name == "minecraft:blue_ice" {
				if world.InWorld(candidate) {
					if err := world.SetBlock(candidate, SurfaceBlockState{Name: "minecraft:blue_ice"}); err != nil {
						return false, err
					}
				}
				break
			}
		}
	}
	return true, nil
}

func ApplyOverworldBlueIceFeature(region *OverworldFeatureRegion, worldSeed int64, chunkX, chunkZ int32,
	feature PlacedFeature, biomeAtQuart func(int32, int32, int32) (string, error)) (int, error) {
	if err := region.requireNeighborhood(chunkX, chunkZ); err != nil {
		return 0, err
	}
	if feature != (PlacedFeature{Step: 4, Index: 3, Name: "minecraft:blue_ice"}) {
		return 0, fmt.Errorf("not the 1.21.1 overworld blue ice feature: %+v", feature)
	}
	placed, err := LoadPlacedFeatureDefinition(feature.Name)
	if err != nil {
		return 0, err
	}
	configured, err := LoadConfiguredFeatureDefinition(placed.Feature)
	if err != nil {
		return 0, err
	}
	if configured.Type != "minecraft:blue_ice" {
		return 0, fmt.Errorf("configured feature %s is %s, not minecraft:blue_ice", placed.Feature, configured.Type)
	}
	successes := 0
	err = visitFeaturePositions(region, worldSeed, chunkX, chunkZ, feature, biomeAtQuart, func(position FeaturePosition, random *WorldgenRandom) error {
		changed, err := PlaceBlueIce(region, random, position)
		if err != nil {
			return err
		}
		if changed {
			successes++
		}
		return nil
	})
	return successes, err
}
