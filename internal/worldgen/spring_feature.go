package worldgen

import (
	"encoding/json"
	"fmt"
)

// SpringConfiguration stores the original fluid state as its legacy block.
// Built-in springs use source fluids; their falling flag does not change the
// source fluid's legacy block level (0).
type SpringConfiguration struct {
	State                SurfaceBlockState
	RequiresBlockBelow   bool
	RockCount, HoleCount int32
	ValidBlocks          map[string]bool
}

func LoadSpringConfiguration(name string) (SpringConfiguration, error) {
	definition, err := LoadConfiguredFeatureDefinition(name)
	if err != nil {
		return SpringConfiguration{}, err
	}
	if definition.Type != "minecraft:spring_feature" {
		return SpringConfiguration{}, fmt.Errorf("%s is not a spring", name)
	}
	var raw struct {
		State  SurfaceBlockState `json:"state"`
		Below  bool              `json:"requires_block_below"`
		Rocks  int32             `json:"rock_count"`
		Holes  int32             `json:"hole_count"`
		Blocks json.RawMessage   `json:"valid_blocks"`
	}
	if err = json.Unmarshal(definition.Config, &raw); err != nil {
		return SpringConfiguration{}, err
	}
	var blocks []string
	var single string
	if json.Unmarshal(raw.Blocks, &single) == nil {
		blocks = []string{single}
	} else if err = json.Unmarshal(raw.Blocks, &blocks); err != nil {
		return SpringConfiguration{}, err
	}
	if (raw.State.Name != "minecraft:water" && raw.State.Name != "minecraft:lava") || raw.Rocks < 0 || raw.Rocks > 5 || raw.Holes < 0 || raw.Holes > 5 || len(blocks) == 0 {
		return SpringConfiguration{}, fmt.Errorf("unsupported spring configuration %s", name)
	}
	config := SpringConfiguration{State: SurfaceBlockState{Name: raw.State.Name, Properties: map[string]string{"level": "0"}}, RequiresBlockBelow: raw.Below, RockCount: raw.Rocks, HoleCount: raw.Holes, ValidBlocks: make(map[string]bool)}
	for _, name := range blocks {
		config.ValidBlocks[name] = true
	}
	return config, nil
}

func PlaceSpringFeature(world FeatureBlockWorld, origin FeaturePosition, config SpringConfiguration) (bool, error) {
	if world == nil || len(config.ValidBlocks) == 0 {
		return false, fmt.Errorf("spring requires a world and valid configuration")
	}
	above, err := world.BlockAt(treeOffset(origin, 0, 1, 0))
	if err != nil {
		return false, err
	}
	if !config.ValidBlocks[above.Name] {
		return false, nil
	}
	below, err := world.BlockAt(treeOffset(origin, 0, -1, 0))
	if err != nil {
		return false, err
	}
	if config.RequiresBlockBelow && !config.ValidBlocks[below.Name] {
		return false, nil
	}
	current, err := world.BlockAt(origin)
	if err != nil {
		return false, err
	}
	if !surfaceAir(current) && !config.ValidBlocks[current.Name] {
		return false, nil
	}
	rocks, holes := int32(0), int32(0)
	for _, offset := range []FeaturePosition{{-1, 0, 0}, {1, 0, 0}, {0, 0, -1}, {0, 0, 1}, {0, -1, 0}} {
		state, err := world.BlockAt(treeOffset(origin, offset.X, offset.Y, offset.Z))
		if err != nil {
			return false, err
		}
		if config.ValidBlocks[state.Name] {
			rocks++
		}
		if surfaceAir(state) {
			holes++
		}
	}
	if rocks != config.RockCount || holes != config.HoleCount {
		return false, nil
	}
	if err = setFeatureBlock(world, origin, config.State); err != nil {
		return false, err
	}
	// WorldGenRegion records the fluid tick; it does not simulate flow during
	// decoration. Persist it with the generated chunk's fluid postprocessing.
	if scheduler, ok := world.(interface{ ScheduleFluidUpdate(FeaturePosition) error }); ok {
		err = scheduler.ScheduleFluidUpdate(origin)
	}
	return err == nil, err
}
