package worldgen

import (
	"encoding/json"
	"fmt"
	"math"
)

type OreTarget struct {
	Predicate string
	Block     string
	Tag       string
	State     SurfaceBlockState
}

type OreConfiguration struct {
	Size            int32
	DiscardOnAir    float32
	Targets         []OreTarget
	replacementTags map[string]map[string]bool
}

type OreWorld interface {
	FeatureBlockWorld
	OceanFloorHeightAt(blockX, blockZ int32) (int32, error)
}

func LoadOreConfiguration(name string) (OreConfiguration, error) {
	definition, err := LoadConfiguredFeatureDefinition(name)
	if err != nil {
		return OreConfiguration{}, err
	}
	if definition.Type != "minecraft:ore" {
		return OreConfiguration{}, fmt.Errorf("configured feature %s is %s, not minecraft:ore", name, definition.Type)
	}
	var raw struct {
		Size         int32   `json:"size"`
		DiscardOnAir float32 `json:"discard_chance_on_air_exposure"`
		Targets      []struct {
			Target struct {
				Predicate string `json:"predicate_type"`
				Block     string `json:"block"`
				Tag       string `json:"tag"`
			} `json:"target"`
			State SurfaceBlockState `json:"state"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(definition.Config, &raw); err != nil {
		return OreConfiguration{}, err
	}
	if raw.Size < 0 || raw.Size > 64 || raw.DiscardOnAir < 0 || raw.DiscardOnAir > 1 || len(raw.Targets) == 0 {
		return OreConfiguration{}, fmt.Errorf("invalid 1.21.1 ore configuration %s", name)
	}
	definitions, err := loadFeatureDefinitions()
	if err != nil {
		return OreConfiguration{}, err
	}
	config := OreConfiguration{Size: raw.Size, DiscardOnAir: raw.DiscardOnAir,
		replacementTags: make(map[string]map[string]bool)}
	for _, target := range raw.Targets {
		oreTarget := OreTarget{Predicate: target.Target.Predicate, Block: target.Target.Block,
			Tag: target.Target.Tag, State: target.State}
		if oreTarget.State.Name == "" {
			return OreConfiguration{}, fmt.Errorf("ore %s has a target without a block state", name)
		}
		switch oreTarget.Predicate {
		case "minecraft:block_match":
			if oreTarget.Block == "" {
				return OreConfiguration{}, fmt.Errorf("ore %s has an empty block matcher", name)
			}
		case "minecraft:tag_match":
			values, exists := definitions.OreTags[oreTarget.Tag]
			if !exists {
				return OreConfiguration{}, fmt.Errorf("ore %s uses unknown replacement tag %s", name, oreTarget.Tag)
			}
			if config.replacementTags[oreTarget.Tag] == nil {
				config.replacementTags[oreTarget.Tag] = make(map[string]bool)
				for _, block := range values {
					config.replacementTags[oreTarget.Tag][block] = true
				}
			}
		default:
			return OreConfiguration{}, fmt.Errorf("ore %s uses unsupported target predicate %s", name, oreTarget.Predicate)
		}
		config.Targets = append(config.Targets, oreTarget)
	}
	return config, nil
}

type oreSphere struct {
	centerX, centerY, centerZ, radius float64
}

func (config OreConfiguration) targetFor(state SurfaceBlockState) (SurfaceBlockState, bool) {
	for _, target := range config.Targets {
		if target.Predicate == "minecraft:block_match" && state.Name == target.Block ||
			target.Predicate == "minecraft:tag_match" && config.replacementTags[target.Tag][state.Name] {
			return target.State, true
		}
	}
	return SurfaceBlockState{}, false
}

func oreAdjacentToAir(world OreWorld, position FeaturePosition) (bool, error) {
	neighbors := [...]FeaturePosition{
		{X: position.X - 1, Y: position.Y, Z: position.Z},
		{X: position.X + 1, Y: position.Y, Z: position.Z},
		{X: position.X, Y: position.Y - 1, Z: position.Z},
		{X: position.X, Y: position.Y + 1, Z: position.Z},
		{X: position.X, Y: position.Y, Z: position.Z - 1},
		{X: position.X, Y: position.Y, Z: position.Z + 1},
	}
	for _, neighbor := range neighbors {
		state, err := world.BlockAt(neighbor)
		if err != nil {
			return false, err
		}
		if surfaceAir(state) {
			return true, nil
		}
	}
	return false, nil
}

func oreMayReplace(world OreWorld, random *WorldgenRandom, position FeaturePosition,
	state SurfaceBlockState, config OreConfiguration) (SurfaceBlockState, bool, error) {
	replacement, matches := config.targetFor(state)
	if !matches {
		return SurfaceBlockState{}, false, nil
	}
	if config.DiscardOnAir > 0 {
		checkAir := config.DiscardOnAir >= 1 || random.NextFloat() < config.DiscardOnAir
		if checkAir {
			adjacent, err := oreAdjacentToAir(world, position)
			if err != nil {
				return SurfaceBlockState{}, false, err
			}
			if adjacent {
				return SurfaceBlockState{}, false, nil
			}
		}
	}
	return replacement, true, nil
}

// PlaceOreFeature ports the vanilla OreFeature algorithm. The supplied world
// must expose all chunks the vein can touch; missing neighbors are an error.
func PlaceOreFeature(world OreWorld, random *WorldgenRandom, origin FeaturePosition, config OreConfiguration) (bool, error) {
	if world == nil || random == nil || config.Size < 0 || config.Size > 64 || len(config.Targets) == 0 {
		return false, fmt.Errorf("ore placement requires a world, random source, and valid configuration")
	}
	angle := random.NextFloat() * float32(math.Pi)
	reach := float32(config.Size) / 8
	padding := int32(math.Ceil(float64(((float32(config.Size)/16)*2 + 1) / 2)))
	startX, startY, startZ := origin.X-int32(math.Ceil(float64(reach)))-padding,
		origin.Y-2-padding, origin.Z-int32(math.Ceil(float64(reach)))-padding
	width := 2 * (int32(math.Ceil(float64(reach))) + padding)
	xStart := float64(origin.X) + math.Sin(float64(angle))*float64(reach)
	xEnd := float64(origin.X) - math.Sin(float64(angle))*float64(reach)
	zStart := float64(origin.Z) + math.Cos(float64(angle))*float64(reach)
	zEnd := float64(origin.Z) - math.Cos(float64(angle))*float64(reach)
	yStart := float64(origin.Y + random.NextIntBound(3) - 2)
	yEnd := float64(origin.Y + random.NextIntBound(3) - 2)
	possible := false
	for blockX := startX; blockX <= startX+width && !possible; blockX++ {
		for blockZ := startZ; blockZ <= startZ+width; blockZ++ {
			columnHeight, err := world.OceanFloorHeightAt(blockX, blockZ)
			if err != nil {
				return false, err
			}
			if startY <= columnHeight {
				possible = true
				break
			}
		}
	}
	if !possible {
		return false, nil
	}
	spheres := make([]oreSphere, config.Size)
	for index := range spheres {
		fraction := float32(index) / float32(config.Size)
		phase := float32(math.Pi) * fraction
		spread := random.NextDouble() * float64(config.Size) / 16
		spheres[index] = oreSphere{
			centerX: xStart + float64(fraction)*(xEnd-xStart),
			centerY: yStart + float64(fraction)*(yEnd-yStart),
			centerZ: zStart + float64(fraction)*(zEnd-zStart),
			radius:  (float64(carverSin(phase)+1)*spread + 1) / 2,
		}
	}
	for first := range spheres {
		if spheres[first].radius <= 0 {
			continue
		}
		for second := first + 1; second < len(spheres); second++ {
			if spheres[second].radius <= 0 {
				continue
			}
			deltaRadius := spheres[first].radius - spheres[second].radius
			deltaX := spheres[first].centerX - spheres[second].centerX
			deltaY := spheres[first].centerY - spheres[second].centerY
			deltaZ := spheres[first].centerZ - spheres[second].centerZ
			if deltaRadius*deltaRadius > deltaX*deltaX+deltaY*deltaY+deltaZ*deltaZ {
				if deltaRadius > 0 {
					spheres[second].radius = -1
				} else {
					spheres[first].radius = -1
				}
			}
		}
	}
	seen := make(map[FeaturePosition]bool)
	placed := false
	for _, sphere := range spheres {
		if sphere.radius < 0 {
			continue
		}
		minX := max(int32(math.Floor(sphere.centerX-sphere.radius)), startX)
		minY := max(int32(math.Floor(sphere.centerY-sphere.radius)), startY)
		minZ := max(int32(math.Floor(sphere.centerZ-sphere.radius)), startZ)
		maxX := max(int32(math.Floor(sphere.centerX+sphere.radius)), minX)
		maxY := max(int32(math.Floor(sphere.centerY+sphere.radius)), minY)
		maxZ := max(int32(math.Floor(sphere.centerZ+sphere.radius)), minZ)
		for blockX := minX; blockX <= maxX; blockX++ {
			xDistance := (float64(blockX) + 0.5 - sphere.centerX) / sphere.radius
			if xDistance*xDistance >= 1 {
				continue
			}
			for blockY := minY; blockY <= maxY; blockY++ {
				yDistance := (float64(blockY) + 0.5 - sphere.centerY) / sphere.radius
				if xDistance*xDistance+yDistance*yDistance >= 1 {
					continue
				}
				for blockZ := minZ; blockZ <= maxZ; blockZ++ {
					zDistance := (float64(blockZ) + 0.5 - sphere.centerZ) / sphere.radius
					position := FeaturePosition{blockX, blockY, blockZ}
					if xDistance*xDistance+yDistance*yDistance+zDistance*zDistance >= 1 ||
						!world.InWorld(position) || seen[position] {
						continue
					}
					seen[position] = true
					state, err := world.BlockAt(position)
					if err != nil {
						return false, err
					}
					replacement, matches, err := oreMayReplace(world, random, position, state, config)
					if err != nil {
						return false, err
					}
					if !matches {
						continue
					}
					if err := world.SetBlock(position, replacement); err != nil {
						return false, err
					}
					placed = true
				}
			}
		}
	}
	return placed, nil
}
