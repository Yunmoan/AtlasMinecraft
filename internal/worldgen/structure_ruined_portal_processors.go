package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed ruined_portal_classifiers_1.21.1.json.gz
var ruinedPortalClassifiers []byte

type ruinedPortalClassifierTable struct {
	Version int         `json:"version"`
	States  [][2]uint32 `json:"states"`
}

var loadRuinedPortalClassifiers = sync.OnceValues(func() (*ruinedPortalClassifierTable, error) {
	gz, err := gzip.NewReader(bytes.NewReader(ruinedPortalClassifiers))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	var result ruinedPortalClassifierTable
	if err := json.NewDecoder(gz).Decode(&result); err != nil {
		return nil, err
	}
	if result.Version != 3955 || len(result.States) != 26684 {
		return nil, fmt.Errorf("incomplete original ruined portal classifiers")
	}
	for _, row := range result.States {
		if row[0] >= 1<<17 || row[1] >= 26684 {
			return nil, fmt.Errorf("invalid ruined portal classifier row")
		}
	}
	return &result, nil
})

func ruinedPortalStateFlags(state SurfaceBlockState) (uint32, error) {
	table, err := loadRuinedPortalClassifiers()
	if err != nil {
		return 0, err
	}
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return 0, err
	}
	return table.States[id][0], nil
}

func ruinedPortalBlackstoneState(state SurfaceBlockState) (SurfaceBlockState, error) {
	table, err := loadRuinedPortalClassifiers()
	if err != nil {
		return SurfaceBlockState{}, err
	}
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return SurfaceBlockState{}, err
	}
	target, err := blockstate.BlockStateByID(table.States[id][1])
	if err != nil {
		return SurfaceBlockState{}, err
	}
	return SurfaceBlockState{Name: target.Name, Properties: target.Properties}, nil
}

func ruinedPortalAgeState(state SurfaceBlockState, p FeaturePosition, mossiness float32) (SurfaceBlockState, error) {
	random := NewLegacyRandom(positionSeed(p.X, p.Y, p.Z))
	flags, err := ruinedPortalStateFlags(state)
	if err != nil {
		return state, err
	}
	choose := func(normal, mossy []SurfaceBlockState) SurfaceBlockState {
		choices := normal
		if random.NextFloat() < mossiness {
			choices = mossy
		}
		return choices[random.NextIntBound(int32(len(choices)))]
	}
	randomStairs := func(name string) SurfaceBlockState {
		facing := []string{"north", "east", "south", "west"}[random.NextIntBound(4)]
		half := []string{"top", "bottom"}[random.NextIntBound(2)]
		return SurfaceBlockState{Name: "minecraft:" + name, Properties: map[string]string{"facing": facing, "half": half}}
	}
	switch {
	case state.Name == "minecraft:stone_bricks" || state.Name == "minecraft:stone" || state.Name == "minecraft:chiseled_stone_bricks":
		if random.NextFloat() >= 0.5 {
			return state, nil
		}
		normal := []SurfaceBlockState{structureState("cracked_stone_bricks"), randomStairs("stone_brick_stairs")}
		mossy := []SurfaceBlockState{structureState("mossy_stone_bricks"), randomStairs("mossy_stone_brick_stairs")}
		return choose(normal, mossy), nil
	case flags&2 != 0:
		if random.NextFloat() >= 0.5 {
			return state, nil
		}
		stairs := SurfaceBlockState{Name: "minecraft:mossy_stone_brick_stairs", Properties: map[string]string{"facing": state.Properties["facing"], "half": state.Properties["half"]}}
		return choose([]SurfaceBlockState{structureState("stone_slab"), structureState("stone_brick_slab")}, []SurfaceBlockState{stairs, structureState("mossy_stone_brick_slab")}), nil
	case flags&4 != 0:
		if random.NextFloat() < mossiness {
			return structureState("mossy_stone_brick_slab"), nil
		}
	case flags&8 != 0:
		if random.NextFloat() < mossiness {
			return structureState("mossy_stone_brick_wall"), nil
		}
	case state.Name == "minecraft:obsidian":
		if random.NextFloat() < 0.15 {
			return structureState("crying_obsidian"), nil
		}
	}
	return state, nil
}

// ProcessRuinedPortalTemplateBlocks reproduces the complete original chain
// before clipping or block-state rotation. Every processor reads the unchanged
// input region; placements must start only after this method returns.
func ProcessRuinedPortalTemplateBlocks(world FeatureBlockWorld, template *StructureTemplate, origin FeaturePosition, placement RuinedPortalVerticalPlacement, properties RuinedPortalProperties, mirror TemplateMirror, rotation TemplateRotation) ([]StructureTemplateBlock, error) {
	if world == nil || template == nil || len(template.Palettes) == 0 {
		return nil, fmt.Errorf("ruined portal processors require template and world")
	}
	pivot := FeaturePosition{template.Size.X / 2, 0, template.Size.Z / 2}
	positionRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	palette := template.Palettes[positionRandom.NextIntBound(int32(len(template.Palettes)))]
	result := make([]StructureTemplateBlock, 0, len(palette))
	for _, block := range palette {
		if block.State.Name == "minecraft:structure_block" || (!properties.AirPocket && block.State.Name == "minecraft:air") {
			continue
		}
		p, err := TransformTemplatePosition(block.Position, pivot, mirror, rotation)
		if err != nil {
			return nil, err
		}
		p = treeOffset(origin, p.X, p.Y, p.Z)
		state := block.State
		if block.NBT != nil {
			block.NBT = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object)
		}
		// RuleProcessor has its own position-seeded LegacyRandomSource.
		random := NewLegacyRandom(positionSeed(p.X, p.Y, p.Z))
		if state.Name == "minecraft:gold_block" && random.NextFloat() < 0.3 {
			state = structureState("air")
		} else if state.Name == "minecraft:lava" {
			if placement == RuinedPortalOnOceanFloor || (!properties.Cold && random.NextFloat() < 0.2) {
				state = structureState("magma_block")
			} else if properties.Cold {
				state = structureState("netherrack")
			}
		} else if !properties.Cold && state.Name == "minecraft:netherrack" && random.NextFloat() < 0.07 {
			state = structureState("magma_block")
		}
		state, err = ruinedPortalAgeState(state, p, properties.Mossiness)
		if err != nil {
			return nil, err
		}
		existing, err := world.BlockAt(p)
		if err != nil {
			return nil, err
		}
		flags, err := ruinedPortalStateFlags(existing)
		if err != nil {
			return nil, err
		}
		if flags&1 != 0 {
			continue
		}
		if existing.Name == "minecraft:lava" {
			flags, err := ruinedPortalStateFlags(state)
			if err != nil {
				return nil, err
			}
			if flags&32 == 0 {
				state = structureState("lava")
			}
		}
		if properties.ReplaceWithBlackstone {
			state, err = ruinedPortalBlackstoneState(state)
			if err != nil {
				return nil, err
			}
		}
		block.Position, block.State = p, state
		result = append(result, block)
	}
	return result, nil
}
