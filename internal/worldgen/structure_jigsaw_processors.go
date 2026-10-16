package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
)

type jigsawRulePredicate struct {
	Type  string `json:"predicate_type"`
	Block string `json:"block"`
	State struct {
		Name       string            `json:"Name"`
		Properties map[string]string `json:"Properties"`
	} `json:"block_state"`
	Probability          float32 `json:"probability"`
	Axis                 string  `json:"axis"`
	Tag                  string  `json:"tag"`
	MinChance, MaxChance float32
	MinDist, MaxDist     int32
}

func (p *jigsawRulePredicate) UnmarshalJSON(data []byte) error {
	type plain jigsawRulePredicate
	var value struct {
		plain
		MinChance float32 `json:"min_chance"`
		MaxChance float32 `json:"max_chance"`
		MinDist   int32   `json:"min_dist"`
		MaxDist   int32   `json:"max_dist"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = jigsawRulePredicate(value.plain)
	p.MinChance, p.MaxChance, p.MinDist, p.MaxDist = value.MinChance, value.MaxChance, value.MinDist, value.MaxDist
	return nil
}

type jigsawProcessorRule struct {
	Input    jigsawRulePredicate `json:"input_predicate"`
	Location jigsawRulePredicate `json:"location_predicate"`
	Position jigsawRulePredicate `json:"position_predicate"`
	Output   struct {
		Name       string            `json:"Name"`
		Properties map[string]string `json:"Properties"`
	} `json:"output_state"`
	Modifier *struct {
		Type      string `json:"type"`
		LootTable string `json:"loot_table"`
	} `json:"block_entity_modifier"`
}
type jigsawProcessor struct {
	Type         string                `json:"processor_type"`
	Rules        []jigsawProcessorRule `json:"rules"`
	Integrity    float32               `json:"integrity"`
	Rottable     string                `json:"rottable_blocks"`
	Protected    string                `json:"value"`
	Delegate     *jigsawProcessor      `json:"delegate"`
	Limit        int32                 `json:"limit"`
	readLocation bool
}
type jigsawProcessors struct {
	Processors []jigsawProcessor `json:"processors"`
	hasCapped  bool
}
type jigsawProcessorEntry struct {
	load func() (*jigsawProcessors, error)
}

var jigsawProcessorCache sync.Map

//go:embed jigsaw_processor_tags_1.21.1.json
var jigsawProcessorTagsJSON []byte

var loadJigsawProcessorTags = sync.OnceValues(func() (map[string]map[string]bool, error) {
	var data struct {
		Version int                 `json:"version"`
		Tags    map[string][]string `json:"tags"`
	}
	if err := json.Unmarshal(jigsawProcessorTagsJSON, &data); err != nil {
		return nil, err
	}
	if data.Version != 3955 || len(data.Tags) != 4 {
		return nil, fmt.Errorf("incomplete original Jigsaw processor tags")
	}
	result := make(map[string]map[string]bool, len(data.Tags))
	for name, values := range data.Tags {
		set := make(map[string]bool, len(values))
		for _, value := range values {
			set[value] = true
		}
		result[name] = set
	}
	return result, nil
})

func jigsawTaggedBlock(tag, name string) (bool, error) {
	tags, err := loadJigsawProcessorTags()
	if err != nil {
		return false, err
	}
	set, ok := tags[strings.TrimPrefix(tag, "#")]
	if !ok {
		return false, fmt.Errorf("unknown original Jigsaw processor block tag %q", tag)
	}
	return set[name], nil
}

func compileJigsawProcessor(processor *jigsawProcessor) error {
	switch processor.Type {
	case "minecraft:capped":
		if processor.Delegate == nil || processor.Limit < 0 {
			return fmt.Errorf("invalid original capped processor")
		}
		return compileJigsawProcessor(processor.Delegate)
	case "minecraft:block_rot":
		if processor.Rottable != "" {
			_, err := jigsawTaggedBlock(processor.Rottable, "")
			return err
		}
		return nil
	case "minecraft:protected_blocks":
		_, err := jigsawTaggedBlock(processor.Protected, "")
		return err
	case "minecraft:rule":
	default:
		return fmt.Errorf("Jigsaw processor %q is not implemented", processor.Type)
	}
	for _, rule := range processor.Rules {
		if rule.Location.Type != "minecraft:always_true" && rule.Location.Type != "" {
			processor.readLocation = true
		}
		for _, predicate := range []jigsawRulePredicate{rule.Input, rule.Location, rule.Position} {
			switch predicate.Type {
			case "", "minecraft:always_true", "minecraft:block_match", "minecraft:random_block_match", "minecraft:blockstate_match", "minecraft:random_blockstate_match", "minecraft:axis_aligned_linear_pos":
			case "minecraft:tag_match":
				if _, err := jigsawTaggedBlock(predicate.Tag, ""); err != nil {
					return err
				}
			default:
				return fmt.Errorf("Jigsaw rule predicate %q is not implemented", predicate.Type)
			}
		}
		if rule.Modifier != nil && rule.Modifier.Type != "minecraft:append_loot" && rule.Modifier.Type != "minecraft:passthrough" {
			return fmt.Errorf("Jigsaw rule block entity modifier %q is not implemented", rule.Modifier.Type)
		}
		if _, err := blockstate.BlockStateID(rule.Output.Name, rule.Output.Properties); err != nil {
			return err
		}
	}
	return nil
}

func loadJigsawProcessors(encoded json.RawMessage) (*jigsawProcessors, error) {
	if len(encoded) == 0 {
		return &jigsawProcessors{}, nil
	}
	key := string(encoded)
	if actual, found := jigsawProcessorCache.Load(key); found {
		return actual.(*jigsawProcessorEntry).load()
	}
	entry := &jigsawProcessorEntry{load: sync.OnceValues(func() (*jigsawProcessors, error) {
		data := encoded
		var name string
		if json.Unmarshal(encoded, &name) == nil {
			var err error
			data, err = JigsawProcessorJSON(name)
			if err != nil {
				return nil, err
			}
		}
		var result jigsawProcessors
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		for i := range result.Processors {
			if err := compileJigsawProcessor(&result.Processors[i]); err != nil {
				return nil, err
			}
			result.hasCapped = result.hasCapped || result.Processors[i].Type == "minecraft:capped" && result.Processors[i].Limit > 0
		}
		return &result, nil
	})}
	actual, _ := jigsawProcessorCache.LoadOrStore(key, entry)
	return actual.(*jigsawProcessorEntry).load()
}

func (predicate jigsawRulePredicate) test(state SurfaceBlockState, current, pivot FeaturePosition, random *LegacyRandom) (bool, error) {
	switch predicate.Type {
	case "", "minecraft:always_true":
		return true, nil
	case "minecraft:block_match":
		return state.Name == predicate.Block, nil
	case "minecraft:random_block_match":
		return state.Name == predicate.Block && random.NextFloat() < predicate.Probability, nil
	case "minecraft:tag_match":
		return jigsawTaggedBlock(predicate.Tag, state.Name)
	case "minecraft:blockstate_match", "minecraft:random_blockstate_match":
		left, err := blockstate.BlockStateID(state.Name, state.Properties)
		if err != nil {
			return false, err
		}
		right, err := blockstate.BlockStateID(predicate.State.Name, predicate.State.Properties)
		if err != nil {
			return false, err
		}
		return left == right && (predicate.Type == "minecraft:blockstate_match" || random.NextFloat() < predicate.Probability), nil
	case "minecraft:axis_aligned_linear_pos":
		delta := current.Y - pivot.Y
		if predicate.Axis == "x" {
			delta = current.X - pivot.X
		}
		if predicate.Axis == "z" {
			delta = current.Z - pivot.Z
		}
		distance := int32(float32(math.Abs(float64(delta))))
		if predicate.MinDist >= predicate.MaxDist {
			return false, fmt.Errorf("invalid Jigsaw linear position range")
		}
		// Both inverseLerp and clampedLerp use the float overloads here.
		t := (float32(distance) - float32(predicate.MinDist)) / (float32(predicate.MaxDist) - float32(predicate.MinDist))
		chance := predicate.MinChance
		if t > 1 {
			chance = predicate.MaxChance
		} else if t >= 0 {
			increment := float32(t * float32(predicate.MaxChance-predicate.MinChance))
			chance = float32(predicate.MinChance + increment)
		}
		return random.NextFloat() <= chance, nil
	default:
		return false, fmt.Errorf("unsupported Jigsaw predicate %q", predicate.Type)
	}
}

func jigsawReplacementState(text string) (SurfaceBlockState, error) {
	name, properties, hasProperties := strings.Cut(text, "[")
	state := SurfaceBlockState{Name: densityName(name)}
	if hasProperties {
		// BlockStateParser consumes through the first closing bracket without
		// requiring end-of-input. The original savanna lamp-post template has
		// an extra trailing ']'; rejecting it removes its acacia-fence support.
		var closed bool
		properties, _, closed = strings.Cut(properties, "]")
		if !closed {
			return state, fmt.Errorf("invalid Jigsaw final state %q", text)
		}
		state.Properties = make(map[string]string)
		var items []string
		if properties != "" {
			items = strings.Split(properties, ",")
		}
		for _, item := range items {
			key, value, found := strings.Cut(item, "=")
			if !found {
				return state, fmt.Errorf("invalid Jigsaw final state %q", text)
			}
			state.Properties[key] = value
		}
	}
	_, err := blockstate.BlockStateID(state.Name, state.Properties)
	return state, err
}

func (processor *jigsawProcessor) process(world FeatureBlockWorld, original, block StructureTemplateBlock, pivot FeaturePosition) (StructureTemplateBlock, bool, error) {
	random := NewLegacyRandom(positionSeed(block.Position.X, block.Position.Y, block.Position.Z))
	switch processor.Type {
	case "minecraft:capped":
		return block, true, nil
	case "minecraft:block_rot":
		if processor.Rottable != "" {
			member, err := jigsawTaggedBlock(processor.Rottable, original.State.Name)
			if err != nil {
				return block, false, err
			}
			if !member {
				return block, true, nil
			}
		}
		return block, random.NextFloat() <= processor.Integrity, nil
	case "minecraft:protected_blocks":
		existing, err := world.BlockAt(block.Position)
		if err != nil {
			return block, false, err
		}
		protected, err := jigsawTaggedBlock(processor.Protected, existing.Name)
		return block, !protected, err
	case "minecraft:rule":
	default:
		return block, false, fmt.Errorf("unknown compiled Jigsaw processor %q", processor.Type)
	}
	var existing SurfaceBlockState
	if processor.readLocation {
		var err error
		existing, err = world.BlockAt(block.Position)
		if err != nil {
			return block, false, err
		}
	}
	for _, rule := range processor.Rules {
		matched, err := rule.Input.test(block.State, block.Position, pivot, random)
		if err != nil {
			return block, false, err
		}
		if !matched {
			continue
		}
		matched, err = rule.Location.test(existing, block.Position, pivot, random)
		if err != nil {
			return block, false, err
		}
		if !matched {
			continue
		}
		matched, err = rule.Position.test(block.State, block.Position, pivot, random)
		if err != nil {
			return block, false, err
		}
		if !matched {
			continue
		}
		block.State = SurfaceBlockState{Name: rule.Output.Name, Properties: rule.Output.Properties}
		if rule.Modifier != nil && rule.Modifier.Type == "minecraft:append_loot" {
			if block.NBT == nil {
				block.NBT = make(nbt.Object)
			} else {
				block.NBT = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object)
			}
			block.NBT["LootTable"] = nbt.Tag{Type: nbt.String, Value: rule.Modifier.LootTable}
			block.NBT["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(random.NextLong())}
		}
		break
	}
	return block, true, nil
}

func finalizeJigsawProcessors(world FeatureBlockWorld, processors *jigsawProcessors, original, result []StructureTemplateBlock, origin, pivot FeaturePosition) ([]StructureTemplateBlock, error) {
	for _, processor := range processors.Processors {
		if processor.Type != "minecraft:capped" || processor.Limit == 0 || len(result) == 0 {
			continue
		}
		seedSource, ok := world.(interface{ WorldSeed() int64 })
		if !ok {
			return nil, fmt.Errorf("capped Jigsaw processor requires world seed")
		}
		if len(original) != len(result) {
			return nil, fmt.Errorf("original/processed Jigsaw block lists are not synchronized")
		}
		random := NewLegacyRandom(seedSource.WorldSeed()).ForkPositional().At(origin.X, origin.Y, origin.Z)
		indices := make([]int, len(result))
		for i := range indices {
			indices[i] = i
		}
		jigsawShuffle(indices, random)
		remaining := min(int(processor.Limit), len(result))
		for _, i := range indices {
			if remaining == 0 {
				break
			}
			processed, keep, err := processor.Delegate.process(world, original[i], result[i], pivot)
			if err != nil {
				return nil, err
			}
			if !keep || jigsawTemplateBlockEqual(result[i], processed) {
				continue
			}
			result[i] = processed
			remaining--
		}
	}
	return result, nil
}

func jigsawTemplateBlockEqual(left, right StructureTemplateBlock) bool {
	if left.Position != right.Position {
		return false
	}
	lid, lerr := blockstate.BlockStateID(left.State.Name, left.State.Properties)
	rid, rerr := blockstate.BlockStateID(right.State.Name, right.State.Properties)
	return lerr == nil && rerr == nil && lid == rid && reflect.DeepEqual(left.NBT, right.NBT)
}

// ProcessJigsawTemplateBlocks runs the original position-seeded processor
// chain before clipping and state rotation. All reads observe the unmodified
// input world, including blocks which will later fall outside the chunk clip.
func ProcessJigsawTemplateBlocks(world FeatureBlockWorld, template *StructureTemplate, element JigsawPoolElement, origin, pivot FeaturePosition, rotation TemplateRotation) ([]StructureTemplateBlock, error) {
	if world == nil || template == nil {
		return nil, fmt.Errorf("Jigsaw processors require a template and world")
	}
	if len(template.Palettes) == 0 {
		return nil, nil
	}
	processors, err := loadJigsawProcessors(element.Processors)
	if err != nil {
		return nil, err
	}
	palette := template.Palettes[NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z)).NextIntBound(int32(len(template.Palettes)))]
	result := make([]StructureTemplateBlock, 0, len(palette))
	var retainedOriginal []StructureTemplateBlock
	for _, original := range palette {
		block := original
		p, err := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
		if err != nil {
			return nil, err
		}
		block.Position = treeOffset(origin, p.X, p.Y, p.Z)
		legacy := element.Type == "minecraft:legacy_single_pool_element"
		if !legacy && block.State.Name == "minecraft:structure_block" {
			continue
		}
		if block.State.Name == "minecraft:jigsaw" && block.NBT != nil {
			final, _ := block.NBT.Text("final_state")
			state, err := jigsawReplacementState(final)
			if err != nil {
				continue
			}
			if state.Name == "minecraft:structure_void" {
				continue
			}
			block.State, block.NBT = state, nil
		}
		if block.NBT != nil {
			block.NBT = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object)
		}
		keep := true
		for i := range processors.Processors {
			block, keep, err = processors.Processors[i].process(world, original, block, pivot)
			if err != nil {
				return nil, err
			}
			if !keep {
				break
			}
		}
		if !keep {
			continue
		}
		if element.Projection == "terrain_matching" {
			height, err := structureTemplateHeight(world, "WORLD_SURFACE_WG", block.Position.X, block.Position.Z)
			if err != nil {
				return nil, err
			}
			block.Position.Y = height - 1 + original.Position.Y
		}
		if legacy && (block.State.Name == "minecraft:air" || block.State.Name == "minecraft:structure_block") {
			continue
		}
		result = append(result, block)
		if processors.hasCapped {
			retainedOriginal = append(retainedOriginal, original)
		}
	}
	return finalizeJigsawProcessors(world, processors, retainedOriginal, result, origin, pivot)
}
