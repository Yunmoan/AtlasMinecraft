package world

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
)

// The default 1.21.1 resource tables only. Optional trade-rebalance datapacks
// are not part of this runtime's active resource configuration.
//
//go:embed container_loot_tables_1.21.1.json
var chestLootAssets []byte

type chestLootFunction struct {
	Function string          `json:"function"`
	Count    json.RawMessage `json:"count"`
	Add      bool            `json:"add"`
	Options  string          `json:"options"`
}
type chestLootEntry struct {
	Type, Name string
	Weight     *int32
	Quality    int32
	Functions  []chestLootFunction
}
type chestLootPool struct {
	Rolls      json.RawMessage
	BonusRolls float32 `json:"bonus_rolls"`
	Entries    []chestLootEntry
	Functions  []chestLootFunction
}
type chestLootTable struct {
	Pools     []chestLootPool
	Functions []chestLootFunction
}
type chestLootTag struct{ Values []string }
type chestLootData struct {
	Tables       map[string]chestLootTable
	Enchantments map[string]struct {
		MaxLevel int32 `json:"max_level"`
	}
	EnchantmentTag chestLootTag `json:"enchantment_tag"`
	InstrumentTag  chestLootTag `json:"instrument_tag"`
}

var readChestLootData = sync.OnceValues(func() (*chestLootData, error) {
	var data chestLootData
	if err := json.Unmarshal(chestLootAssets, &data); err != nil {
		return nil, err
	}
	if len(data.Tables) != 17 || len(data.InstrumentTag.Values) != 4 {
		return nil, fmt.Errorf("incomplete original chest loot assets")
	}
	for _, name := range data.EnchantmentTag.Values {
		if strings.HasPrefix(name, "#") || data.Enchantments[name].MaxLevel <= 0 {
			return nil, fmt.Errorf("unresolved original enchantment tag %s", name)
		}
	}
	return &data, nil
})

// FillChestLoot ports LootTable.fill for these original tables: pool expansion,
// weighted selection, function order, maximum-stack splitting, slot shuffling
// and shuffleAndSplitItems all consume the original LegacyRandomSource stream.
// Nonzero seeds are the explicit LootTableSeed path used by generated chests.
func FillChestLoot(table string, seed int64, luck float32) ([27]item.Stack, error) {
	var slots [27]item.Stack
	if seed == 0 {
		return slots, fmt.Errorf("unseeded loot requires the world's saved random_sequence stream")
	}
	if math.IsNaN(float64(luck)) || math.IsInf(float64(luck), 0) {
		return slots, fmt.Errorf("invalid loot luck")
	}
	assets, err := readChestLootData()
	if err != nil {
		return slots, err
	}
	definition, ok := assets.Tables[table]
	if !ok {
		return slots, fmt.Errorf("unsupported original chest loot table %s", table)
	}
	random := worldgen.NewLegacyRandom(seed)
	var stacks []item.Stack
	for _, pool := range definition.Pools {
		rolls, err := chestLootInt(pool.Rolls, random)
		if err != nil {
			return slots, err
		}
		rolls += int32(math.Floor(float64(pool.BonusRolls * luck)))
		for roll := int32(0); roll < rolls; roll++ {
			var entries []chestLootEntry
			var weights []int32
			var total int32
			for _, entry := range pool.Entries {
				weight := int32(1)
				if entry.Weight != nil {
					weight = *entry.Weight
				}
				weight = max(0, int32(math.Floor(float64(float32(weight)+float32(entry.Quality)*luck))))
				if weight > 0 {
					entries = append(entries, entry)
					weights = append(weights, weight)
					total += weight
				}
			}
			if len(entries) == 0 {
				continue
			}
			index := 0
			if len(entries) > 1 {
				pick := random.NextIntBound(total)
				for i, weight := range weights {
					pick -= weight
					if pick < 0 {
						index = i
						break
					}
				}
			}
			entry := entries[index]
			if entry.Type == "minecraft:empty" {
				continue
			}
			if entry.Type != "minecraft:item" {
				return slots, fmt.Errorf("unsupported chest entry %s", entry.Type)
			}
			id, ok := item.ID(entry.Name)
			if !ok {
				return slots, fmt.Errorf("unknown original loot item %s", entry.Name)
			}
			stack := item.New(id, 1)
			for _, functions := range [][]chestLootFunction{entry.Functions, pool.Functions, definition.Functions} {
				for _, function := range functions {
					stack, err = applyChestLootFunction(assets, stack, function, random)
					if err != nil {
						return slots, err
					}
				}
			}
			for !stack.Empty() {
				count := min(stack.Count, stack.MaxCount())
				stacks = append(stacks, stack.WithCount(count))
				stack = stack.WithCount(stack.Count - count)
			}
		}
	}
	available := make([]int, 27)
	for i := range available {
		available[i] = i
	}
	shuffleChestList(available, random)
	var singles, split []item.Stack
	for _, stack := range stacks {
		if stack.Count > 1 {
			split = append(split, stack)
		} else if !stack.Empty() {
			singles = append(singles, stack)
		}
	}
	for len(available)-len(singles)-len(split) > 0 && len(split) > 0 {
		index := int(chestLootRange(random, 0, int32(len(split)-1)))
		stack := split[index]
		split = append(split[:index], split[index+1:]...)
		count := chestLootRange(random, 1, stack.Count/2)
		part := stack.WithCount(count)
		stack = stack.WithCount(stack.Count - count)
		if stack.Count > 1 && random.NextBoolean() {
			split = append(split, stack)
		} else {
			singles = append(singles, stack)
		}
		if part.Count > 1 && random.NextBoolean() {
			split = append(split, part)
		} else {
			singles = append(singles, part)
		}
	}
	singles = append(singles, split...)
	shuffleChestList(singles, random)
	for _, stack := range singles {
		if len(available) == 0 {
			break
		}
		last := len(available) - 1
		slots[available[last]] = stack
		available = available[:last]
	}
	return slots, nil
}

func chestLootRange(random *worldgen.LegacyRandom, minValue, maxValue int32) int32 {
	if minValue >= maxValue {
		return minValue
	}
	return random.NextIntBound(maxValue-minValue+1) + minValue
}
func chestLootInt(raw json.RawMessage, random *worldgen.LegacyRandom) (int32, error) {
	var constant float32
	if err := json.Unmarshal(raw, &constant); err == nil {
		return int32(math.Floor(float64(constant))), nil
	}
	var uniform struct {
		Type     string
		Min, Max float32
	}
	if err := json.Unmarshal(raw, &uniform); err != nil || uniform.Type != "minecraft:uniform" {
		return 0, fmt.Errorf("unsupported original chest number provider %s", raw)
	}
	return chestLootRange(random, int32(math.Floor(float64(uniform.Min))), int32(math.Floor(float64(uniform.Max)))), nil
}
func shuffleChestList[T any](values []T, random *worldgen.LegacyRandom) {
	for i := len(values); i > 1; i-- {
		j := int(random.NextIntBound(int32(i)))
		values[i-1], values[j] = values[j], values[i-1]
	}
}
func applyChestLootFunction(assets *chestLootData, stack item.Stack, function chestLootFunction, random *worldgen.LegacyRandom) (item.Stack, error) {
	switch function.Function {
	case "minecraft:set_count":
		count, err := chestLootInt(function.Count, random)
		if err != nil {
			return stack, err
		}
		if function.Add {
			count += stack.Count
		}
		return stack.WithCount(count), nil
	case "minecraft:enchant_randomly":
		name, _ := item.Name(stack.ID)
		if name != "minecraft:book" || function.Options != "#minecraft:on_random_loot" {
			return stack, fmt.Errorf("unsupported original chest enchantment function")
		}
		options := assets.EnchantmentTag.Values
		choice := options[random.NextIntBound(int32(len(options)))]
		level := chestLootRange(random, 1, assets.Enchantments[choice].MaxLevel)
		return item.FromNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
			"id": {Type: nbt.String, Value: "minecraft:enchanted_book"}, "count": {Type: nbt.Int, Value: int32(1)},
			"components": {Type: nbt.Compound, Value: nbt.Object{"minecraft:stored_enchantments": {Type: nbt.Compound, Value: nbt.Object{"levels": {Type: nbt.Compound, Value: nbt.Object{choice: {Type: nbt.Int, Value: level}}}}}}},
		}})
	case "minecraft:set_instrument":
		if function.Options != "#minecraft:regular_goat_horns" {
			return stack, fmt.Errorf("unsupported original chest instrument tag")
		}
		choice := assets.InstrumentTag.Values[random.NextIntBound(int32(len(assets.InstrumentTag.Values)))]
		name, _ := item.Name(stack.ID)
		return item.FromNBT(nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: name}, "count": {Type: nbt.Int, Value: stack.Count}, "components": {Type: nbt.Compound, Value: nbt.Object{"minecraft:instrument": {Type: nbt.String, Value: choice}}}}})
	default:
		return stack, fmt.Errorf("unsupported original chest function %s", function.Function)
	}
}
