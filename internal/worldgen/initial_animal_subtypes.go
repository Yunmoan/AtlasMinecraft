package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
)

//go:embed other_initial_animal_data.json
var otherInitialAnimalJSON []byte

type otherInitialAnimalTables struct {
	Types          map[string]initialAnimalType `json:"types"`
	States         [][2]uint32                  `json:"states"`
	PlacementLower []bool                       `json:"placement_lower"`
	BiomeTags      map[string][]string          `json:"biome_tags"`
	BlockTags      map[string][]string          `json:"block_tags"`
}

var loadOtherInitialAnimalTables = sync.OnceValues(func() (*otherInitialAnimalTables, error) {
	var table otherInitialAnimalTables
	err := json.Unmarshal(otherInitialAnimalJSON, &table)
	if err == nil && (len(table.Types) != 12 || len(table.States) != 26684) {
		err = fmt.Errorf("incomplete other initial animal tables")
	}
	return &table, err
})

func animalBiomeTag(biome, tag string) bool {
	tables, err := loadOtherInitialAnimalTables()
	return err == nil && slices.Contains(tables.BiomeTags["minecraft:"+tag], biome)
}

func animalSpawnGroundAllowed(id, biome, below string, y, seaLevel int32) (bool, error) {
	tables, err := loadOtherInitialAnimalTables()
	if err != nil {
		return false, err
	}
	tag := "animals_spawnable_on"
	switch id {
	case "minecraft:armadillo":
		tag = "armadillo_spawnable_on"
	case "minecraft:fox":
		tag = "foxes_spawnable_on"
	case "minecraft:frog":
		tag = "frogs_spawnable_on"
	case "minecraft:goat":
		tag = "goats_spawnable_on"
	case "minecraft:mooshroom":
		tag = "mooshrooms_spawnable_on"
	case "minecraft:parrot":
		tag = "parrots_spawnable_on"
	case "minecraft:rabbit":
		tag = "rabbits_spawnable_on"
	case "minecraft:wolf":
		tag = "wolves_spawnable_on"
	case "minecraft:polar_bear":
		if animalBiomeTag(biome, "polar_bears_spawn_on_alternate_blocks") {
			tag = "polar_bears_spawnable_on_alternate"
		}
	case "minecraft:turtle":
		return y < seaLevel+4 && (below == "minecraft:sand" || below == "minecraft:red_sand"), nil
	}
	return slices.Contains(tables.BlockTags[tag], below), nil
}

func animalBrainMemory(values map[string]int32) nbt.Tag {
	memories := nbt.Object{}
	for name, value := range values {
		memories["minecraft:"+name] = nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"value": {Type: nbt.Int, Value: value}}}
	}
	return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"memories": {Type: nbt.Compound, Value: memories}}}
}

func pandaGene(draw int32) string {
	switch {
	case draw == 0:
		return "lazy"
	case draw == 1:
		return "worried"
	case draw == 2:
		return "playful"
	case draw == 4:
		return "aggressive"
	case draw < 9:
		return "weak"
	case draw < 11:
		return "brown"
	default:
		return "normal"
	}
}

// The subtype draws precede AgeableMob and Mob finalization. The group retains
// only data shared by the official pack, and does not use constructor randomness.
func finalizeOtherInitialAnimal(id, biome string, position FeaturePosition, random *Xoroshiro, group *initialAnimalGroup, result nbt.Object) (attributes []nbt.Tag, babyChance float32, noBaby, forceBaby bool, err error) {
	babyChance = .05
	setInt := func(name string, value int32) { result[name] = nbt.Tag{Type: nbt.Int, Value: value} }
	setText := func(name, value string) { result[name] = nbt.Tag{Type: nbt.String, Value: value} }
	setBool := func(name string, value bool) {
		var v int8
		if value {
			v = 1
		}
		result[name] = nbt.Tag{Type: nbt.Byte, Value: v}
	}
	switch id {
	case "minecraft:armadillo", "minecraft:mooshroom":
	case "minecraft:polar_bear":
		babyChance = 1
	case "minecraft:wolf":
		noBaby = true
		if group.Count == 0 {
			variant := "pale"
			switch {
			case biome == "minecraft:snowy_taiga":
				variant = "ashen"
			case biome == "minecraft:old_growth_pine_taiga":
				variant = "black"
			case biome == "minecraft:old_growth_spruce_taiga":
				variant = "chestnut"
			case animalBiomeTag(biome, "is_jungle"):
				variant = "rusty"
			case biome == "minecraft:grove":
				variant = "snowy"
			case animalBiomeTag(biome, "is_savanna"):
				variant = "spotted"
			case animalBiomeTag(biome, "is_badlands"):
				variant = "striped"
			case biome == "minecraft:forest":
				variant = "woods"
			}
			group.VariantName = "minecraft:" + variant
		}
		setText("variant", group.VariantName)
	case "minecraft:fox":
		noBaby, forceBaby = true, group.Count >= 2
		if group.Count == 0 {
			group.VariantName = "red"
			if animalBiomeTag(biome, "spawns_snow_foxes") {
				group.VariantName = "snow"
			}
		}
		setText("Type", group.VariantName)
		if random.NextFloat() < .2 {
			value, item := random.NextFloat(), "feather"
			switch {
			case value < .05:
				item = "emerald"
			case value < .2:
				item = "egg"
			case value < .4:
				item = "rabbit_hide"
				if random.NextBoolean() {
					item = "rabbit_foot"
				}
			case value < .6:
				item = "wheat"
			case value < .8:
				item = "leather"
			}
			result["HandItems"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:" + item}, "count": {Type: nbt.Int, Value: int32(1)}}}, {Type: nbt.Compound, Value: nbt.Object{}}}}
		}
	case "minecraft:rabbit":
		babyChance = 1
		draw := random.NextIntBound(100)
		if group.Count == 0 {
			switch {
			case animalBiomeTag(biome, "spawns_white_rabbits"):
				group.Variant = 1
				if draw >= 80 {
					group.Variant = 3
				}
			case animalBiomeTag(biome, "spawns_gold_rabbits"):
				group.Variant = 4
			case draw < 50:
				group.Variant = 0
			case draw < 90:
				group.Variant = 5
			default:
				group.Variant = 2
			}
		}
		setInt("RabbitType", group.Variant)
		attributes = append(attributes, animalAttribute("attack_damage", 3))
	case "minecraft:parrot":
		noBaby = true
		setInt("Variant", random.NextIntBound(5))
	case "minecraft:panda":
		babyChance = .2
		main, hidden := pandaGene(random.NextIntBound(16)), pandaGene(random.NextIntBound(16))
		setText("MainGene", main)
		setText("HiddenGene", hidden)
		expressed := main
		if (main == "weak" || main == "brown") && hidden != main {
			expressed = "normal"
		}
		if expressed == "weak" {
			attributes = append(attributes, animalAttribute("max_health", 10))
		}
		if expressed == "lazy" {
			attributes = append(attributes, animalAttribute("movement_speed", float64(float32(.07))))
		}
	case "minecraft:llama":
		bound := int32(3)
		if random.NextFloat() < .04 {
			bound = 5
		}
		setInt("Strength", random.NextIntBound(bound)+1)
		if group.Count == 0 {
			group.Variant = random.NextIntBound(4)
		}
		setInt("Variant", group.Variant)
		health := float64(float32(15) + float32(random.NextIntBound(8)) + float32(random.NextIntBound(9)))
		attributes = append(attributes, animalAttribute("max_health", health))
	case "minecraft:frog":
		variant := "temperate"
		if animalBiomeTag(biome, "spawns_cold_variant_frogs") {
			variant = "cold"
		} else if animalBiomeTag(biome, "spawns_warm_variant_frogs") {
			variant = "warm"
		}
		setText("variant", "minecraft:"+variant)
		result["Brain"] = animalBrainMemory(map[string]int32{"long_jump_cooling_down": random.NextIntBound(41) + 100})
	case "minecraft:goat":
		attributes = append(attributes, animalAttribute("attack_damage", 2))
		jump, ram := random.NextIntBound(601)+600, random.NextIntBound(5401)+600
		result["Brain"] = animalBrainMemory(map[string]int32{"long_jump_cooling_down": jump, "ram_cooldown_ticks": ram})
		setBool("IsScreamingGoat", random.NextDouble() < .02)
		left, right := true, true
		if random.NextFloat() < .1 {
			if random.NextBoolean() {
				left = false
			} else {
				right = false
			}
		}
		setBool("HasLeftHorn", left)
		setBool("HasRightHorn", right)
	case "minecraft:turtle":
		setInt("HomePosX", position.X)
		setInt("HomePosY", position.Y)
		setInt("HomePosZ", position.Z)
		setInt("TravelPosX", 0)
		setInt("TravelPosY", 0)
		setInt("TravelPosZ", 0)
	default:
		err = fmt.Errorf("unsupported animal finalization %s", id)
	}
	return
}
