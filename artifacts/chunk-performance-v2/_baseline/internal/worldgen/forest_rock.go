package worldgen

import (
	"encoding/json"
	"fmt"
)

var forestRockBaseStone = map[string]bool{
	"minecraft:stone": true, "minecraft:granite": true, "minecraft:diorite": true,
	"minecraft:andesite": true, "minecraft:tuff": true, "minecraft:deepslate": true,
}

var forestRockDirt = map[string]bool{
	"minecraft:dirt": true, "minecraft:grass_block": true, "minecraft:podzol": true,
	"minecraft:coarse_dirt": true, "minecraft:mycelium": true, "minecraft:rooted_dirt": true,
	"minecraft:moss_block": true, "minecraft:mud": true, "minecraft:muddy_mangrove_roots": true,
}

func LoadForestRockState(name string) (SurfaceBlockState, error) {
	definition, err := LoadConfiguredFeatureDefinition(name)
	if err != nil {
		return SurfaceBlockState{}, err
	}
	if definition.Type != "minecraft:forest_rock" {
		return SurfaceBlockState{}, fmt.Errorf("configured feature %s is %s, not minecraft:forest_rock", name, definition.Type)
	}
	var config struct {
		State SurfaceBlockState `json:"state"`
	}
	if err := json.Unmarshal(definition.Config, &config); err != nil {
		return SurfaceBlockState{}, err
	}
	if config.State.Name == "" {
		return SurfaceBlockState{}, fmt.Errorf("forest rock %s has no block state", name)
	}
	return config.State, nil
}

func PlaceForestRock(world FeatureBlockWorld, random *WorldgenRandom, origin FeaturePosition,
	state SurfaceBlockState) (bool, error) {
	if world == nil || random == nil || state.Name == "" {
		return false, fmt.Errorf("forest rock requires a world, random source, and block state")
	}
	center := origin
	for center.Y > -61 {
		below := FeaturePosition{center.X, center.Y - 1, center.Z}
		base, err := world.BlockAt(below)
		if err != nil {
			return false, err
		}
		if forestRockBaseStone[base.Name] || forestRockDirt[base.Name] {
			break
		}
		center.Y--
	}
	if center.Y <= -61 {
		return false, nil
	}
	for range 3 {
		radiusX := random.NextIntBound(2)
		radiusY := random.NextIntBound(2)
		radiusZ := random.NextIntBound(2)
		radius := float32(radiusX+radiusY+radiusZ)*0.333 + 0.5
		radiusSquared := float64(radius * radius)
		for blockY := center.Y - radiusY; blockY <= center.Y+radiusY; blockY++ {
			for blockZ := center.Z - radiusZ; blockZ <= center.Z+radiusZ; blockZ++ {
				for blockX := center.X - radiusX; blockX <= center.X+radiusX; blockX++ {
					deltaX, deltaY, deltaZ := blockX-center.X, blockY-center.Y, blockZ-center.Z
					if float64(deltaX*deltaX+deltaY*deltaY+deltaZ*deltaZ) > radiusSquared {
						continue
					}
					position := FeaturePosition{blockX, blockY, blockZ}
					if world.InWorld(position) {
						if err := world.SetBlock(position, state); err != nil {
							return false, err
						}
					}
				}
			}
		}
		center.X += -1 + random.NextIntBound(2)
		center.Y -= random.NextIntBound(2)
		center.Z += -1 + random.NextIntBound(2)
	}
	return true, nil
}

func ApplyOverworldForestRockFeature(region *OverworldFeatureRegion, worldSeed int64, chunkX, chunkZ int32,
	feature PlacedFeature, biomeAtQuart func(int32, int32, int32) (string, error)) (int, error) {
	if err := region.requireNeighborhood(chunkX, chunkZ); err != nil {
		return 0, err
	}
	if feature != (PlacedFeature{Step: 2, Index: 4, Name: "minecraft:forest_rock"}) {
		return 0, fmt.Errorf("not the 1.21.1 overworld forest rock feature: %+v", feature)
	}
	placed, err := LoadPlacedFeatureDefinition(feature.Name)
	if err != nil {
		return 0, err
	}
	state, err := LoadForestRockState(placed.Feature)
	if err != nil {
		return 0, err
	}
	successes := 0
	err = visitFeaturePositions(region, worldSeed, chunkX, chunkZ, feature, biomeAtQuart, func(position FeaturePosition, random *WorldgenRandom) error {
		changed, err := PlaceForestRock(region, random, position, state)
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
