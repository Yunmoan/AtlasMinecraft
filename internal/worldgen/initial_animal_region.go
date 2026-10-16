package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"atlas/internal/light"
	"atlas/internal/nbt"
	"fmt"
	"math"
	"strings"
	"sync"
)

var initialAnimalTypeOrder = []string{"minecraft:sheep", "minecraft:pig", "minecraft:chicken", "minecraft:cow", "minecraft:horse", "minecraft:donkey", "minecraft:armadillo", "minecraft:fox", "minecraft:frog", "minecraft:goat", "minecraft:llama", "minecraft:mooshroom", "minecraft:panda", "minecraft:parrot", "minecraft:polar_bear", "minecraft:rabbit", "minecraft:turtle", "minecraft:wolf"}

type animalHeightKey struct {
	X, Z         int32
	Leaves       bool
	Unrestricted bool
}

type initialAnimalRegion struct {
	region  *OverworldFeatureRegion
	cx, cz  int32
	seed    int64
	tables  *initialAnimalTables
	heights map[animalHeightKey]int32
	lights  *light.Engine
}

func (w *initialAnimalRegion) SpawnPosition(id string, x, z int32) (FeaturePosition, error) {
	leaves := w.tables.Types[id].Heightmap == "MOTION_BLOCKING"
	key := animalHeightKey{x, z, leaves, id == "minecraft:fox" || id == "minecraft:panda"}
	if y, ok := w.heights[key]; ok {
		return FeaturePosition{x, y, z}, nil
	}
	for y := w.region.MaxBuildHeight() - 1; y >= w.region.MinBuildHeight(); y-- {
		state, err := w.region.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return FeaturePosition{}, err
		}
		blocking, err := motionBlocking(state)
		if err != nil {
			return FeaturePosition{}, err
		}
		if blocking && (leaves || !strings.HasSuffix(state.Name, "_leaves")) {
			tables, err := loadOtherInitialAnimalTables()
			if err != nil {
				return FeaturePosition{}, err
			}
			stateID, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				return FeaturePosition{}, err
			}
			spawnY := y + 1
			if id != "minecraft:fox" && id != "minecraft:panda" && tables.PlacementLower[stateID] {
				spawnY = y
			}
			w.heights[key] = spawnY
			return FeaturePosition{x, spawnY, z}, nil
		}
	}
	w.heights[key] = w.region.MinBuildHeight()
	return FeaturePosition{x, w.region.MinBuildHeight(), z}, nil
}

func initialAnimalBit(id string) uint32 {
	for i, name := range initialAnimalTypeOrder {
		if name == id {
			return 1 << i
		}
	}
	return 0
}
func (w *initialAnimalRegion) PositionOK(id string, p FeaturePosition) (bool, error) {
	if id == "minecraft:fox" || id == "minecraft:panda" {
		return true, nil
	}
	if p.X < -29999984 || p.X >= 29999984 || p.Z < -29999984 || p.Z >= 29999984 {
		return false, nil
	}
	bit := initialAnimalBit(id)
	for i, at := range []FeaturePosition{{p.X, p.Y - 1, p.Z}, p, {p.X, p.Y + 1, p.Z}} {
		state, err := w.region.BlockAt(at)
		if err != nil {
			return false, err
		}
		stateID, err := blockstate.BlockStateID(state.Name, state.Properties)
		if err != nil {
			return false, err
		}
		column := 1
		if i == 0 {
			column = 0
		}
		if w.tables.States[stateID][column]&bit == 0 {
			return false, nil
		}
	}
	return true, nil
}

func (w *initialAnimalRegion) NoCollision(id string, p [3]float64) (bool, error) {
	kind := w.tables.Types[id]
	half := float64(kind.Width) / 2
	box := blockrules.Box{p[0] - half, p[1], p[2] - half, p[0] + half, p[1] + float64(kind.Height), p[2] + half}
	for x := int32(math.Floor(box[0])); x <= int32(math.Floor(box[3])); x++ {
		for y := int32(math.Floor(box[1])); y <= int32(math.Floor(box[4])); y++ {
			for z := int32(math.Floor(box[2])); z <= int32(math.Floor(box[5])); z++ {
				at := FeaturePosition{x, y, z}
				state, err := w.region.BlockAt(at)
				if err != nil {
					return false, err
				}
				stateID, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					return false, err
				}
				rule, err := blockrules.Lookup(stateID)
				if err != nil {
					return false, err
				}
				offset := BlockOffset(state.Name, at)
				for _, shape := range rule.Boxes {
					if box[0] < float64(x)+shape[3]+offset[0] && box[3] > float64(x)+shape[0]+offset[0] && box[1] < float64(y)+shape[4]+offset[1] && box[4] > float64(y)+shape[1]+offset[1] && box[2] < float64(z)+shape[5]+offset[2] && box[5] > float64(z)+shape[2]+offset[2] {
						return false, nil
					}
				}
			}
		}
	}
	return true, nil
}

func (w *initialAnimalRegion) AnimalBiomeAt(p FeaturePosition) (string, error) {
	qx, qy, qz := BiomeQuartAtBlock(w.seed, p.X, p.Y, p.Z)
	return w.region.biomeAtQuart(qx, qy, qz)
}

func (w *initialAnimalRegion) SpawnRules(id string, p FeaturePosition, _ *Xoroshiro) (bool, error) {
	below, err := w.region.BlockAt(FeaturePosition{p.X, p.Y - 1, p.Z})
	if err != nil {
		return false, err
	}
	biome, err := w.AnimalBiomeAt(p)
	if err != nil {
		return false, err
	}
	allowed, err := animalSpawnGroundAllowed(id, biome, below.Name, p.Y, w.region.generationSeaLevel)
	if err != nil || !allowed {
		return false, err
	}
	brightness, err := w.rawBrightness(p)
	return brightness > 8, err
}
func (w *initialAnimalRegion) MobSpawnOK(id string, p [3]float64) (bool, error) {
	kind := w.tables.Types[id]
	half := float64(kind.Width) / 2
	for x := int32(math.Floor(p[0] - half)); x < int32(math.Ceil(p[0]+half)); x++ {
		for y := int32(math.Floor(p[1])); y < int32(math.Ceil(p[1]+float64(kind.Height))); y++ {
			for z := int32(math.Floor(p[2] - half)); z < int32(math.Ceil(p[2]+half)); z++ {
				state, err := w.region.BlockAt(FeaturePosition{x, y, z})
				if err != nil {
					return false, err
				}
				if state.Name == "minecraft:water" || state.Name == "minecraft:lava" || state.Name == "minecraft:bubble_column" || state.Properties["waterlogged"] == "true" {
					return false, nil
				}
			}
		}
	}
	return true, nil
}

func (w *initialAnimalRegion) rawBrightness(p FeaturePosition) (int32, error) {
	if w.lights == nil {
		for dx := int32(-1); dx <= 1; dx++ {
			for dz := int32(-1); dz <= 1; dz++ {
				if _, err := w.region.chunkAt((w.cx+dx)*16, (w.cz+dz)*16); err != nil {
					return 0, err
				}
			}
		}
	}
	if w.lights == nil && p.Y >= w.region.minY && p.Y < w.region.minY+w.region.height {
		chunk, err := w.region.chunkAt(p.X, p.Z)
		if err != nil {
			return 0, err
		}
		overlay := w.region.overlayAt(p.X, p.Z)
		var path [384]uint16
		count := 0
		for y := chunk.MinY + chunk.Height - 1; y >= p.Y; y-- {
			state := w.region.stateAt(chunk, overlay, p.X&15, y, p.Z&15)
			id, err := blockstate.BlockStateID(state.Name, state.Properties)
			if err != nil {
				return 0, err
			}
			path[count] = uint16(id)
			count++
		}
		visible, err := light.DirectSkylight(path[:count])
		if err != nil {
			return 0, err
		}
		if visible {
			return 15, nil
		}
	}
	if w.lights == nil {
		engine, err := light.New(int(w.region.minY), int(w.region.height), true)
		if err != nil {
			return 0, err
		}
		columns, err := w.lightColumns()
		if err != nil {
			return 0, err
		}
		if _, err := engine.InitializeArea(light.Position{X: w.cx, Z: w.cz}, columns); err != nil {
			return 0, err
		}
		w.lights = engine
	}
	data := w.lights.Snapshot(light.Position{X: p.X >> 4, Z: p.Z >> 4})
	if data == nil {
		return 0, nil
	}
	section := int((p.Y-w.region.minY)>>4) + 1
	if section < 0 || section >= len(data.Sky) {
		return 0, nil
	}
	at := int((p.Y&15)<<8 | (p.Z&15)<<4 | p.X&15)
	read := func(layers [][]byte) int32 {
		if layers[section] == nil {
			return 0
		}
		return int32(layers[section][at>>1] >> uint((at&1)*4) & 15)
	}
	return max(read(data.Sky), read(data.Block)), nil
}

var initialAnimalMaterialStates = sync.OnceValues(func() ([NoiseEndStone + 1]uint16, error) {
	var states [NoiseEndStone + 1]uint16
	for material := NoiseAir; material <= NoiseEndStone; material++ {
		id, err := blockstate.BlockStateID(material.BlockName(), nil)
		if err != nil {
			return states, err
		}
		states[material] = uint16(id)
	}
	return states, nil
})

func (w *initialAnimalRegion) lightColumns() (map[light.Position][]uint16, error) {
	materials, err := initialAnimalMaterialStates()
	if err != nil {
		return nil, err
	}
	ids := make(map[string]uint16)
	apply := func(states []uint16, overrides map[int]SurfaceBlockState) error {
		for at, state := range overrides {
			key := state.Name
			if len(state.Properties) > 0 {
				key += fmt.Sprint(state.Properties)
			}
			id, exists := ids[key]
			if !exists {
				value, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					return err
				}
				id = uint16(value)
				ids[key] = id
			}
			states[at] = id
		}
		return nil
	}
	columns := make(map[light.Position][]uint16, 9)
	for dx := int32(-1); dx <= 1; dx++ {
		for dz := int32(-1); dz <= 1; dz++ {
			cx, cz := w.cx+dx, w.cz+dz
			chunk, err := w.region.chunkAt(cx*16, cz*16)
			if err != nil {
				return nil, err
			}
			states := make([]uint16, len(chunk.Blocks))
			for at, material := range chunk.Blocks {
				if int(material) >= len(materials) {
					return nil, fmt.Errorf("unknown noise material %d", material)
				}
				states[at] = materials[material]
			}
			if err := apply(states, chunk.SurfaceBlocks); err != nil {
				return nil, err
			}
			if overlay := w.region.overlayAt(cx*16, cz*16); overlay != nil {
				if err := apply(states, overlay.blocks); err != nil {
					return nil, err
				}
			}
			columns[light.Position{X: cx, Z: cz}] = states
		}
	}
	return columns, nil
}

func (g *OverworldGenerator) generateInitialAnimals(region *OverworldFeatureRegion, cx, cz int32) ([]string, error) {
	qx, qy, qz := BiomeQuartAtBlock(g.seed, cx*16, region.MaxBuildHeight()-1, cz*16)
	biome, err := region.biomeAtQuart(qx, qy, qz)
	if err != nil {
		return nil, err
	}
	tables, err := loadInitialAnimalTables()
	if err != nil {
		return nil, err
	}
	random := NewLegacyRandom(g.seed)
	a, b := random.NextLong()|1, random.NextLong()|1
	seed := uint64(int64(cx*16))*a + uint64(int64(cz*16))*b ^ uint64(g.seed)
	random = NewLegacyRandom(int64(seed))
	env := &initialAnimalRegion{region: region, cx: cx, cz: cz, seed: g.seed, tables: tables, heights: make(map[animalHeightKey]int32)}
	animals, unsupported, err := SpawnInitialAnimals(biome, cx, cz, random, NewStructureRegionRandom(g.seed, cx, cz), env)
	if err != nil {
		return nil, err
	}
	for i, animal := range animals {
		at := FeaturePosition{int32(math.Floor(animal.Position[0])), int32(math.Floor(animal.Position[1])), int32(math.Floor(animal.Position[2]))}
		if animal.ID == "minecraft:chicken" {
			constructor := NewLegacyRandom(int64(positionSeed(at.X, at.Y, at.Z)) ^ g.seed)
			animal.Data["EggLayTime"] = nbt.Tag{Type: nbt.Int, Value: constructor.NextIntBound(6000) + 6000}
		}
		if animal.ID == "minecraft:armadillo" {
			constructor := NewLegacyRandom(int64(positionSeed(at.X, at.Y, at.Z)) ^ g.seed)
			animal.Data["scute_time"] = nbt.Tag{Type: nbt.Int, Value: constructor.NextIntBound(6000) + 6000}
		}
		if err := region.AddPendingEntity(PendingStructureEntity{ID: animal.ID, Position: at, SavedFlag: fmt.Sprintf("InitialAnimal:%d", i), Data: animal.Data}); err != nil {
			return nil, err
		}
	}
	return unsupported, nil
}
