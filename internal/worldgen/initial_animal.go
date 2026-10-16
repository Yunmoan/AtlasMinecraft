package worldgen

import (
	"atlas/internal/nbt"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"sync"
)

//go:embed initial_animal_data.json
var initialAnimalJSON []byte

type InitialAnimalEntry struct {
	ID     string `json:"id"`
	Weight int32  `json:"weight"`
	Min    int32  `json:"min"`
	Max    int32  `json:"max"`
}
type InitialAnimalSettings struct {
	Probability float32              `json:"probability"`
	Entries     []InitialAnimalEntry `json:"entries"`
}
type initialAnimalType struct {
	Width     float32 `json:"width"`
	Height    float32 `json:"height"`
	Summon    bool    `json:"summon"`
	Heightmap string  `json:"heightmap"`
}
type initialAnimalTables struct {
	Biomes map[string]InitialAnimalSettings `json:"biomes"`
	Types  map[string]initialAnimalType     `json:"types"`
	States [][2]uint32                      `json:"states"`
}

var loadInitialAnimalTables = sync.OnceValues(func() (*initialAnimalTables, error) {
	var t initialAnimalTables
	err := json.Unmarshal(initialAnimalJSON, &t)
	if err == nil && (len(t.Biomes) != 64 || len(t.States) != 26684) {
		err = fmt.Errorf("incomplete initial animal tables")
	}
	if err == nil {
		other, loadErr := loadOtherInitialAnimalTables()
		err = loadErr
		if err == nil {
			for id, kind := range other.Types {
				t.Types[id] = kind
			}
			for i, state := range other.States {
				t.States[i][0] |= state[0] << 6
				t.States[i][1] |= state[1] << 6
			}
		}
	}
	return &t, err
})

// InitialAnimalEnvironment supplies the official placement and collision
// predicates. SpawnRules receives the independent WorldGenRegion random stream.
type InitialAnimalEnvironment interface {
	SpawnPosition(id string, x, z int32) (FeaturePosition, error)
	PositionOK(id string, p FeaturePosition) (bool, error)
	NoCollision(id string, p [3]float64) (bool, error)
	SpawnRules(id string, p FeaturePosition, random *Xoroshiro) (bool, error)
	MobSpawnOK(id string, p [3]float64) (bool, error)
}

type InitialAnimal struct {
	ID       string
	Position [3]float64
	Yaw      float32
	Data     nbt.Object
}
type initialAnimalGroup struct {
	Count        int
	HorseVariant int32
	Variant      int32
	VariantName  string
}

// SpawnInitialAnimals transcribes NaturalSpawner.spawnMobsForChunkGeneration.
// Unsupported species are reported before consuming either random stream.
func SpawnInitialAnimals(biome string, cx, cz int32, random *LegacyRandom, worldRandom *Xoroshiro, world InitialAnimalEnvironment) ([]InitialAnimal, []string, error) {
	tables, err := loadInitialAnimalTables()
	if err != nil {
		return nil, nil, err
	}
	settings, exists := tables.Biomes[biome]
	if !exists {
		return nil, nil, fmt.Errorf("unknown initial animal biome %s", biome)
	}
	var unsupported []string
	for _, entry := range settings.Entries {
		if _, ok := tables.Types[entry.ID]; !ok {
			unsupported = append(unsupported, entry.ID)
		}
	}
	if len(unsupported) > 0 {
		return nil, unsupported, nil
	}
	if len(settings.Entries) == 0 {
		return nil, nil, nil
	}
	var total int32
	for _, entry := range settings.Entries {
		total += entry.Weight
	}
	minX, minZ := cx*16, cz*16
	var result []InitialAnimal
	for random.NextFloat() < settings.Probability {
		if total == 0 {
			continue
		}
		weight := random.NextIntBound(total)
		var selected InitialAnimalEntry
		for _, entry := range settings.Entries {
			weight -= entry.Weight
			if weight < 0 {
				selected = entry
				break
			}
		}
		count := selected.Min + random.NextIntBound(1+selected.Max-selected.Min)
		x, z := minX+random.NextIntBound(16), minZ+random.NextIntBound(16)
		anchorX, anchorZ := x, z
		group := initialAnimalGroup{}
		for member := int32(0); member < count; member++ {
			success := false
			for attempt := 0; !success && attempt < 4; attempt++ {
				p, err := world.SpawnPosition(selected.ID, x, z)
				if err != nil {
					return nil, nil, err
				}
				kind := tables.Types[selected.ID]
				ok, err := world.PositionOK(selected.ID, p)
				if err != nil {
					return nil, nil, err
				}
				if kind.Summon && ok {
					width := float64(kind.Width)
					position := [3]float64{max(float64(minX)+width, min(float64(minX+16)-width, float64(x))), float64(p.Y), max(float64(minZ)+width, min(float64(minZ+16)-width, float64(z)))}
					ok, err = world.NoCollision(selected.ID, position)
					if err != nil {
						return nil, nil, err
					}
					if !ok {
						continue
					}
					rulePosition := FeaturePosition{int32(math.Floor(position[0])), p.Y, int32(math.Floor(position[2]))}
					ok, err = world.SpawnRules(selected.ID, rulePosition, worldRandom)
					if err != nil {
						return nil, nil, err
					}
					if !ok {
						continue
					}
					yaw := random.NextFloat() * 360
					ok, err = world.MobSpawnOK(selected.ID, position)
					if err != nil {
						return nil, nil, err
					}
					if ok {
						actualBiome := biome
						if lookup, ok := world.(interface {
							AnimalBiomeAt(FeaturePosition) (string, error)
						}); ok {
							actualBiome, err = lookup.AnimalBiomeAt(rulePosition)
							if err != nil {
								return nil, nil, err
							}
						}
						data, err := finalizeInitialAnimalAt(selected.ID, actualBiome, rulePosition, worldRandom, &group)
						if err != nil {
							return nil, nil, err
						}
						data["Pos"] = animalPositionNBT(position)
						data["Rotation"] = nbt.Tag{Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{{Type: nbt.Float, Value: yaw}, {Type: nbt.Float, Value: float32(0)}}}
						result = append(result, InitialAnimal{selected.ID, position, yaw, data})
						success = true
					}
				}
				// Failed collision/rule/constructor branches above deliberately skip
				// these jitter draws; the official loop uses continue in those branches.
				x += random.NextIntBound(5) - random.NextIntBound(5)
				z += random.NextIntBound(5) - random.NextIntBound(5)
				for x < minX || x >= minX+16 || z < minZ || z >= minZ+16 {
					x = anchorX + random.NextIntBound(5) - random.NextIntBound(5)
					z = anchorZ + random.NextIntBound(5) - random.NextIntBound(5)
				}
			}
		}
	}
	return result, nil, nil
}

func animalPositionNBT(p [3]float64) nbt.Tag {
	values := make([]nbt.Tag, 3)
	for i, v := range p {
		values[i] = nbt.Tag{Type: nbt.Double, Value: v}
	}
	return nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: values}
}

func animalAttribute(id string, base float64) nbt.Tag {
	return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:generic." + id}, "base": {Type: nbt.Double, Value: base}}}
}

func finalizeInitialAnimal(id string, random *Xoroshiro, group *initialAnimalGroup) (nbt.Object, error) {
	return finalizeInitialAnimalAt(id, "minecraft:plains", FeaturePosition{}, random, group)
}

func finalizeInitialAnimalAt(id, biome string, position FeaturePosition, random *Xoroshiro, group *initialAnimalGroup) (nbt.Object, error) {
	result := nbt.Object{}
	attributes := []nbt.Tag{}
	babyChance := float32(.05)
	noBaby, forceBaby := false, false
	switch id {
	case "minecraft:sheep":
		color := int8(0)
		draw := random.NextIntBound(100)
		switch {
		case draw < 5:
			color = 15
		case draw < 10:
			color = 7
		case draw < 15:
			color = 8
		case draw < 18:
			color = 12
		default:
			if random.NextIntBound(500) == 0 {
				color = 6
			}
		}
		result["Color"] = nbt.Tag{Type: nbt.Byte, Value: color}
	case "minecraft:horse":
		if group.Count == 0 {
			group.HorseVariant = random.NextIntBound(7)
		}
		marking := random.NextIntBound(5)
		result["Variant"] = nbt.Tag{Type: nbt.Int, Value: group.HorseVariant | marking<<8}
		health := float64(float32(15) + float32(random.NextIntBound(8)) + float32(random.NextIntBound(9)))
		speed := (float64(float32(.45)) + random.NextDouble()*.3 + random.NextDouble()*.3 + random.NextDouble()*.3) * .25
		jump := float64(float32(.4)) + random.NextDouble()*.2 + random.NextDouble()*.2 + random.NextDouble()*.2
		attributes = append(attributes, animalAttribute("max_health", health), animalAttribute("movement_speed", speed), animalAttribute("jump_strength", jump))
	case "minecraft:donkey":
		babyChance = .2
		health := float64(float32(15) + float32(random.NextIntBound(8)) + float32(random.NextIntBound(9)))
		attributes = append(attributes, animalAttribute("max_health", health))
	case "minecraft:pig", "minecraft:chicken", "minecraft:cow":
	default:
		var err error
		attributes, babyChance, noBaby, forceBaby, err = finalizeOtherInitialAnimal(id, biome, position, random, group, result)
		if err != nil {
			return nil, err
		}
	}
	age := int32(0)
	if forceBaby || !noBaby && group.Count > 0 && random.NextFloat() <= babyChance {
		age = -24000
	}
	group.Count++
	result["Age"] = nbt.Tag{Type: nbt.Int, Value: age}
	if id == "minecraft:goat" && age < 0 {
		result["HasLeftHorn"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		result["HasRightHorn"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		attributes[0] = animalAttribute("attack_damage", 1)
	}
	bonus := 0.0 + .11485000000000001*(random.NextDouble()-random.NextDouble())
	left := int8(0)
	if random.NextFloat() < .05 {
		left = 1
	}
	modifier := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:random_spawn_bonus"}, "amount": {Type: nbt.Double, Value: bonus}, "operation": {Type: nbt.String, Value: "add_multiplied_base"}}
	followRange := 16.0
	if id == "minecraft:fox" {
		followRange = 32
	}
	if id == "minecraft:llama" {
		followRange = 40
	}
	if id == "minecraft:polar_bear" {
		followRange = 20
	}
	follow := animalAttribute("follow_range", followRange)
	follow.Value.(nbt.Object)["modifiers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: modifier}}}
	attributes = append(attributes, follow)
	result["attributes"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: attributes}
	result["LeftHanded"] = nbt.Tag{Type: nbt.Byte, Value: left}
	return result, nil
}
