package worldgen

import (
	"atlas/internal/nbt"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type treeDecorator struct {
	kind         string
	probability  float32
	xz, y, empty int32
	directions   []int
	provider     featureStateProvider
}

func decodeTreeDecorator(raw json.RawMessage) (treeDecorator, error) {
	var data struct {
		Type          string          `json:"type"`
		Probability   float32         `json:"probability"`
		XZ            int32           `json:"exclusion_radius_xz"`
		Y             int32           `json:"exclusion_radius_y"`
		Empty         int32           `json:"required_empty_blocks"`
		Directions    []string        `json:"directions"`
		Provider      json.RawMessage `json:"provider"`
		BlockProvider json.RawMessage `json:"block_provider"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return treeDecorator{}, err
	}
	d := treeDecorator{kind: data.Type, probability: data.Probability, xz: data.XZ, y: data.Y, empty: data.Empty}
	if data.Probability < 0 || data.Probability > 1 {
		return d, fmt.Errorf("invalid tree decorator probability")
	}
	var err error
	switch data.Type {
	case "minecraft:beehive", "minecraft:cocoa", "minecraft:leave_vine", "minecraft:trunk_vine":
	case "minecraft:alter_ground":
		d.provider, err = decodeFeatureStateProvider(data.Provider, func(SurfaceBlockState) error { return nil })
	case "minecraft:attached_to_leaves":
		if data.XZ < 0 || data.XZ > 16 || data.Y < 0 || data.Y > 16 || data.Empty < 1 || data.Empty > 16 || len(data.Directions) == 0 {
			return d, fmt.Errorf("invalid attached-to-leaves decorator bounds")
		}
		for _, name := range data.Directions {
			found := false
			for i, dir := range directionNames {
				if dir == name {
					d.directions = append(d.directions, i)
					found = true
					break
				}
			}
			if !found {
				return d, fmt.Errorf("invalid tree decorator direction %s", name)
			}
		}
		d.provider, err = decodeFeatureStateProvider(data.BlockProvider, func(SurfaceBlockState) error { return nil })
	default:
		return d, fmt.Errorf("unsupported tree decorator %s", data.Type)
	}
	return d, err
}

// HasBeehiveDecorator identifies configurations that require an independent
// shuffle source. Collections.shuffle in vanilla does not use the world seed.
func (config TreeConfiguration) HasBeehiveDecorator() bool {
	for _, d := range config.decorators {
		if d.kind == "minecraft:beehive" {
			return true
		}
	}
	return false
}

// WithBeehiveShuffle supplies the independent java.util.Collections shuffle.
// Region replay must freeze its result for the same canonical tree invocation.
func (config TreeConfiguration) WithBeehiveShuffle(shuffle func([]FeaturePosition)) TreeConfiguration {
	config.beehiveShuffle = shuffle
	return config
}

func sortedTreePositions(set *javaPositionSet) []FeaturePosition {
	positions := set.positions()
	// ObjectArrayList.sort is stable: equal-Y entries retain HashSet order.
	sort.SliceStable(positions, func(i, j int) bool { return positions[i].Y < positions[j].Y })
	return positions
}

func (tree *treePlacement) decoration(p FeaturePosition, state SurfaceBlockState) {
	if tree.decorations == nil {
		tree.decorations = make(map[FeaturePosition]bool)
	}
	tree.set(p, state, tree.decorations)
}

func (tree *treePlacement) air(p FeaturePosition) bool {
	s := tree.block(p)
	return s.Name == "minecraft:air" || s.Name == "minecraft:cave_air" || s.Name == "minecraft:void_air"
}

func (tree *treePlacement) vine(p FeaturePosition, face int) {
	tree.decoration(p, SurfaceBlockState{Name: "minecraft:vine", Properties: map[string]string{
		"up": "false", "north": "false", "south": "false", "east": "false", "west": "false", directionNames[face]: "true",
	}})
}

func (tree *treePlacement) placeDecorators() {
	if len(tree.config.decorators) == 0 {
		return
	}
	logs, leaves := sortedTreePositions(&tree.trunkOrder), sortedTreePositions(&tree.leafOrder)
	roots := sortedTreePositions(&tree.rootOrder)
	for _, d := range tree.config.decorators {
		switch d.kind {
		case "minecraft:trunk_vine", "minecraft:leave_vine":
			positions := logs
			if d.kind == "minecraft:leave_vine" {
				positions = leaves
			}
			for _, p := range positions {
				// West, east, north, south; the vine faces toward the tree.
				for _, dir := range []int{4, 5, 2, 3} {
					place := false
					if d.kind == "minecraft:trunk_vine" {
						place = tree.random.NextIntBound(3) > 0
					} else {
						place = tree.random.NextFloat() < d.probability
					}
					q := directionPosition(p, dir)
					if place && tree.air(q) {
						tree.vine(q, dir^1)
						if d.kind == "minecraft:leave_vine" {
							q.Y--
							for left := 4; tree.air(q) && left > 0; left-- {
								tree.vine(q, dir^1)
								q.Y--
							}
						}
					}
				}
			}
		case "minecraft:cocoa":
			if tree.random.NextFloat() >= d.probability || len(logs) == 0 {
				continue
			}
			for _, p := range logs {
				if p.Y-logs[0].Y > 2 {
					continue
				}
				for _, dir := range []int{2, 5, 3, 4} {
					if tree.random.NextFloat() > 0.25 {
						continue
					}
					q := directionPosition(p, dir^1)
					if tree.air(q) {
						tree.decoration(q, SurfaceBlockState{Name: "minecraft:cocoa", Properties: map[string]string{"age": strconv.Itoa(int(tree.random.NextIntBound(3))), "facing": directionNames[dir]}})
					}
				}
			}
		case "minecraft:attached_to_leaves":
			positions := append([]FeaturePosition(nil), leaves...)
			for i := len(positions); i > 1; i-- {
				j := tree.random.NextIntBound(int32(i))
				positions[i-1], positions[j] = positions[j], positions[i-1]
			}
			excluded := make(map[FeaturePosition]bool)
			for _, p := range positions {
				dir := d.directions[tree.random.NextIntBound(int32(len(d.directions)))]
				q := directionPosition(p, dir)
				if excluded[q] || tree.random.NextFloat() >= d.probability {
					continue
				}
				clear := true
				for i := int32(1); i <= d.empty; i++ {
					v := featureDirections[dir]
					if !tree.air(treeOffset(p, v.X*i, v.Y*i, v.Z*i)) {
						clear = false
						break
					}
				}
				if !clear {
					continue
				}
				for x := -d.xz; x <= d.xz; x++ {
					for y := -d.y; y <= d.y; y++ {
						for z := -d.xz; z <= d.xz; z++ {
							excluded[treeOffset(q, x, y, z)] = true
						}
					}
				}
				tree.decoration(q, d.provider(tree.random, q))
			}
		case "minecraft:alter_ground":
			positions := logs
			if len(roots) != 0 {
				positions = roots
				if len(logs) != 0 && roots[0].Y == logs[0].Y {
					positions = append(append([]FeaturePosition(nil), logs...), roots...)
				}
			}
			if len(positions) == 0 {
				continue
			}
			for _, p := range positions {
				if p.Y != positions[0].Y {
					continue
				}
				for _, delta := range []FeaturePosition{{-1, 0, -1}, {2, 0, -1}, {-1, 0, 2}, {2, 0, 2}} {
					tree.groundCircle(treeOffset(p, delta.X, 0, delta.Z), d.provider)
				}
				for i := 0; i < 5; i++ {
					n := tree.random.NextIntBound(64)
					x, z := n%8, n/8
					if x == 0 || x == 7 || z == 0 || z == 7 {
						tree.groundCircle(treeOffset(p, x-3, 0, z-3), d.provider)
					}
				}
			}
		case "minecraft:beehive":
			tree.beehive(logs, leaves, d.probability)
		}
		if tree.err != nil {
			return
		}
	}
}

func (tree *treePlacement) groundCircle(p FeaturePosition, provider featureStateProvider) {
	for x := int32(-2); x <= 2; x++ {
		for z := int32(-2); z <= 2; z++ {
			if abs32(x) == 2 && abs32(z) == 2 {
				continue
			}
			q := treeOffset(p, x, 0, z)
			for y := int32(2); y >= -3; y-- {
				n := treeOffset(q, 0, y, 0)
				if treeTags["dirt"][tree.block(n).Name] {
					tree.decoration(n, provider(tree.random, q))
					break
				}
				if !tree.air(n) && y < 0 {
					break
				}
			}
		}
	}
}

type treeBlockEntityWorld interface {
	SetBlockEntity(FeaturePosition, string, nbt.Object) error
}

func (tree *treePlacement) beehive(logs, leaves []FeaturePosition, probability float32) {
	if tree.random.NextFloat() >= probability || len(logs) == 0 {
		return
	}
	y := int32(0)
	if len(leaves) != 0 {
		y = max(leaves[0].Y-1, logs[0].Y+1)
	} else {
		y = min(logs[0].Y+1+tree.random.NextIntBound(3), logs[len(logs)-1].Y)
	}
	var candidates []FeaturePosition
	for _, p := range logs {
		if p.Y == y {
			for _, dir := range []int{5, 3, 4} {
				candidates = append(candidates, directionPosition(p, dir))
			}
		}
	}
	if len(candidates) == 0 {
		return
	}
	if tree.config.beehiveShuffle == nil {
		tree.err = fmt.Errorf("beehive decoration requires a frozen independent shuffle for region replay")
		return
	}
	tree.config.beehiveShuffle(candidates)
	for _, p := range candidates {
		if !tree.air(p) || !tree.air(directionPosition(p, 3)) {
			continue
		}
		tree.decoration(p, SurfaceBlockState{Name: "minecraft:bee_nest", Properties: map[string]string{"facing": "south", "honey_level": "0"}})
		if world, ok := tree.world.(treeBlockEntityWorld); ok && tree.err == nil && tree.world.InWorld(p) {
			count := 2 + tree.random.NextIntBound(2)
			bees := make([]nbt.Tag, count)
			for i := range bees {
				bees[i] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
					"entity_data":       {Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:bee"}}},
					"ticks_in_hive":     {Type: nbt.Int, Value: tree.random.NextIntBound(599)},
					"min_ticks_in_hive": {Type: nbt.Int, Value: int32(600)},
				}}
			}
			tree.err = world.SetBlockEntity(p, "minecraft:bee_nest", nbt.Object{"id": {Type: nbt.String, Value: "minecraft:beehive"}, "bees": {Type: nbt.List, ListType: nbt.Compound, Value: bees}})
		}
		return
	}
}
