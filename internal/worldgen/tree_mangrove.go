package worldgen

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type mangroveRootConfiguration struct {
	offset                                 featureIntProvider
	provider, muddyProvider, aboveProvider featureStateProvider
	growThrough, muddyIn                   map[string]bool
	width, length                          int32
	skew, aboveChance                      float32
}

func decodeTreeBlockSet(raw json.RawMessage) (map[string]bool, error) {
	var name string
	if json.Unmarshal(raw, &name) == nil {
		if strings.HasPrefix(name, "#minecraft:") {
			blocks, ok := treeTags[strings.TrimPrefix(name, "#minecraft:")]
			if !ok {
				return nil, fmt.Errorf("unsupported tree block tag %s", name)
			}
			return blocks, nil
		}
		if name == "" {
			return nil, fmt.Errorf("empty tree block set")
		}
		return map[string]bool{name: true}, nil
	}
	var names []string
	if err := json.Unmarshal(raw, &names); err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(names))
	for _, name := range names {
		result[name] = true
	}
	return result, nil
}

func decodeMangroveRoot(raw json.RawMessage) (*mangroveRootConfiguration, error) {
	var data struct {
		Type     string          `json:"type"`
		Offset   json.RawMessage `json:"trunk_offset_y"`
		Provider json.RawMessage `json:"root_provider"`
		Above    *struct {
			Provider json.RawMessage `json:"above_root_provider"`
			Chance   float32         `json:"above_root_placement_chance"`
		} `json:"above_root_placement"`
		Mangrove struct {
			GrowThrough   json.RawMessage `json:"can_grow_through"`
			MuddyIn       json.RawMessage `json:"muddy_roots_in"`
			MuddyProvider json.RawMessage `json:"muddy_roots_provider"`
			Width         int32           `json:"max_root_width"`
			Length        int32           `json:"max_root_length"`
			Skew          float32         `json:"random_skew_chance"`
		} `json:"mangrove_root_placement"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data.Type != "minecraft:mangrove_root_placer" {
		return nil, fmt.Errorf("unsupported root placer %s", data.Type)
	}
	config := &mangroveRootConfiguration{width: data.Mangrove.Width, length: data.Mangrove.Length, skew: data.Mangrove.Skew}
	if config.width < 1 || config.width > 8 || config.length < 1 || config.length > 15 || config.skew < 0 || config.skew > 1 {
		return nil, fmt.Errorf("invalid mangrove root bounds")
	}
	var err error
	if config.offset, err = decodeFeatureIntProvider(data.Offset, -64, 64); err != nil {
		return nil, err
	}
	if config.provider, err = decodeFeatureStateProvider(data.Provider, func(SurfaceBlockState) error { return nil }); err != nil {
		return nil, err
	}
	if config.muddyProvider, err = decodeFeatureStateProvider(data.Mangrove.MuddyProvider, func(SurfaceBlockState) error { return nil }); err != nil {
		return nil, err
	}
	if config.growThrough, err = decodeTreeBlockSet(data.Mangrove.GrowThrough); err != nil {
		return nil, err
	}
	if config.muddyIn, err = decodeTreeBlockSet(data.Mangrove.MuddyIn); err != nil {
		return nil, err
	}
	if data.Above != nil {
		config.aboveChance = data.Above.Chance
		if config.aboveChance < 0 || config.aboveChance > 1 {
			return nil, fmt.Errorf("invalid above-root placement chance")
		}
		if config.aboveProvider, err = decodeFeatureStateProvider(data.Above.Provider, func(SurfaceBlockState) error { return nil }); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (tree *treePlacement) validLog(s SurfaceBlockState) bool {
	return treeReplaceable(s) || tree.config.growThrough[s.Name]
}
func (tree *treePlacement) canRoot(p FeaturePosition) bool {
	s := tree.block(p)
	return treeReplaceable(s) || tree.config.root.growThrough[s.Name]
}

func (tree *treePlacement) mangroveRoots(origin, trunk FeaturePosition) bool {
	for p := origin; p.Y < trunk.Y; p.Y++ {
		if !tree.canRoot(p) {
			return false
		}
	}
	positions := []FeaturePosition{treeOffset(trunk, 0, -1, 0)}
	for _, dir := range []int{2, 5, 3, 4} {
		p := directionPosition(trunk, dir)
		var roots []FeaturePosition
		if !tree.simulateMangroveRoots(p, dir, trunk, &roots, 0) {
			return false
		}
		positions = append(positions, roots...)
		positions = append(positions, p)
	}
	for _, p := range positions {
		tree.placeMangroveRoot(p)
	}
	return tree.err == nil
}

func (tree *treePlacement) simulateMangroveRoots(p FeaturePosition, dir int, origin FeaturePosition, result *[]FeaturePosition, depth int32) bool {
	config := tree.config.root
	if depth == config.length || len(*result) > int(config.length) {
		return false
	}
	down, outward := treeOffset(p, 0, -1, 0), directionPosition(p, dir)
	distance := abs32(p.X-origin.X) + abs32(p.Y-origin.Y) + abs32(p.Z-origin.Z)
	var next []FeaturePosition
	if distance > config.width-3 && distance <= config.width {
		if tree.random.NextFloat() < config.skew {
			next = []FeaturePosition{down, treeOffset(outward, 0, -1, 0)}
		} else {
			next = []FeaturePosition{down}
		}
	} else if distance > config.width {
		next = []FeaturePosition{down}
	} else if tree.random.NextFloat() < config.skew {
		next = []FeaturePosition{down}
	} else if tree.random.NextBoolean() {
		next = []FeaturePosition{outward}
	} else {
		next = []FeaturePosition{down}
	}
	for _, q := range next {
		if !tree.canRoot(q) {
			continue
		}
		*result = append(*result, q)
		if !tree.simulateMangroveRoots(q, dir, origin, result, depth+1) {
			return false
		}
	}
	return tree.err == nil
}

func (tree *treePlacement) setRoot(p FeaturePosition, state SurfaceBlockState) {
	if tree.roots == nil {
		tree.roots = make(map[FeaturePosition]bool)
	}
	if _, ok := state.Properties["waterlogged"]; ok {
		old := tree.block(p)
		water := old.Name == "minecraft:water" || old.Properties["waterlogged"] == "true" || old.Name == "minecraft:kelp" || old.Name == "minecraft:kelp_plant" || old.Name == "minecraft:seagrass" || old.Name == "minecraft:tall_seagrass" || old.Name == "minecraft:bubble_column"
		state = stateWithProperty(state, "waterlogged", strconv.FormatBool(water))
	}
	tree.set(p, state, tree.roots)
	tree.rootOrder.add(p)
}

func (tree *treePlacement) placeMangroveRoot(p FeaturePosition) {
	config := tree.config.root
	if config.muddyIn[tree.block(p).Name] {
		tree.setRoot(p, config.muddyProvider(tree.random, p))
		return
	}
	if !tree.canRoot(p) {
		return
	}
	tree.setRoot(p, config.provider(tree.random, p))
	if config.aboveProvider != nil {
		above := treeOffset(p, 0, 1, 0)
		if tree.random.NextFloat() < config.aboveChance && tree.air(above) {
			tree.setRoot(above, config.aboveProvider(tree.random, above))
		}
	}
}

func (tree *treePlacement) upwardsTrunk(origin FeaturePosition, height int32) []treeFoliageAttachment {
	var attachments []treeFoliageAttachment
	for y := int32(0); y < height; y++ {
		p := treeOffset(origin, 0, y, 0)
		if tree.log(p, "y") && y < height-1 && tree.random.NextFloat() < tree.config.branchProbability {
			dir := featureHorizontalDirections[tree.random.NextIntBound(4)]
			length := tree.config.extraBranchLength(tree.random)
			start := max(int32(0), length-tree.config.extraBranchLength(tree.random)-1)
			steps := tree.config.extraBranchSteps(tree.random)
			top := p.Y + start
			for branchY := start; branchY < height && steps > 0; branchY, steps = branchY+1, steps-1 {
				if branchY < 1 {
					continue
				}
				p.X += dir.X
				p.Z += dir.Z
				p.Y = origin.Y + y + branchY
				top = p.Y
				if tree.log(p, "y") {
					top++
				}
				attachments = append(attachments, treeFoliageAttachment{position: p})
			}
			if top-(origin.Y+y) > 1 {
				p.Y = top
				attachments = append(attachments, treeFoliageAttachment{position: p}, treeFoliageAttachment{position: treeOffset(p, 0, -2, 0)})
			}
		}
		if y == height-1 {
			attachments = append(attachments, treeFoliageAttachment{position: treeOffset(origin, 0, y+1, 0)})
		}
	}
	return attachments
}
