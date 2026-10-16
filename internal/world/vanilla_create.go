package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CreateFlatWorld creates a Java 1.21.1 save with the stock flat layers and
// disabled structures/features.
func CreateFlatWorld(dir string, requestedSeed *int64) error {
	return createJavaWorld(dir, requestedSeed, true)
}

// CreateNormalWorld creates a Java 1.21.1 Overworld noise save.
func CreateNormalWorld(dir string, requestedSeed *int64) error {
	return createJavaWorld(dir, requestedSeed, false)
}

func createJavaWorld(dir string, requestedSeed *int64, flat bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("refusing to create a world in nonempty directory %s", dir)
	}
	seed := int64(0)
	if requestedSeed != nil {
		seed = *requestedSeed
	} else {
		var raw [8]byte
		if _, err := rand.Read(raw[:]); err != nil {
			return err
		}
		seed = int64(binary.LittleEndian.Uint64(raw[:]))
	}
	// Vanilla selects the initial spawn chunk with the climate-based search
	// (Climate$SpawnFinder) before any chunk exists; the boot-time scan then
	// refines the exact column. Flat worlds keep column (0,0), where the
	// spawn scan always succeeds on the stock layers.
	spawnX, spawnZ := int32(0), int32(0)
	if !flat {
		x, z, err := worldgen.FindSpawnPosition(seed)
		if err != nil {
			return err
		}
		// setInitialSpawn's fallback stores the center block of the selected
		// chunk at generator spawn height (sea level + 1).
		spawnX, spawnZ = (x>>4)*16+8, (z>>4)*16+8
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	root := newWorldLevelAt(filepath.Base(filepath.Clean(dir)), seed, flat, spawnX, spawnZ)
	if !flat {
		levelRoot, _ := root.Object()
		level, _ := levelRoot.Object("Data")
		if _, _, err := savedBeehiveShuffleSalt(level); err != nil {
			return err
		}
	}
	return writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), "", root)
}

func newWorldLevel(name string, seed int64, flat bool) nbt.Tag {
	return newWorldLevelAt(name, seed, flat, 0, 0)
}

func newWorldLevelAt(name string, seed int64, flat bool, spawnX, spawnZ int32) nbt.Tag {
	stringTag := func(value string) nbt.Tag { return nbt.Tag{Type: nbt.String, Value: value} }
	compoundTag := func(value nbt.Object) nbt.Tag { return nbt.Tag{Type: nbt.Compound, Value: value} }
	byteTag := func(value int8) nbt.Tag { return nbt.Tag{Type: nbt.Byte, Value: value} }
	emptyStrings := nbt.Tag{Type: nbt.List, ListType: nbt.String, Value: []nbt.Tag{}}
	stringList := func(values ...string) nbt.Tag {
		items := make([]nbt.Tag, 0, len(values))
		for _, value := range values {
			items = append(items, stringTag(value))
		}
		return nbt.Tag{Type: nbt.List, ListType: nbt.String, Value: items}
	}
	layers := []nbt.Tag{}
	for _, layer := range []struct {
		block  string
		height int32
	}{{"minecraft:bedrock", 1}, {"minecraft:dirt", 2}, {"minecraft:grass_block", 1}} {
		layers = append(layers, compoundTag(nbt.Object{
			"block": stringTag(layer.block), "height": {Type: nbt.Int, Value: layer.height},
		}))
	}
	overworldGenerator := nbt.Object{
		"type": stringTag("minecraft:flat"),
		"settings": compoundTag(nbt.Object{
			"biome": stringTag("minecraft:plains"), "features": byteTag(0), "lakes": byteTag(0),
			"layers":              {Type: nbt.List, ListType: nbt.Compound, Value: layers},
			"structure_overrides": emptyStrings,
		}),
	}
	if !flat {
		overworldGenerator = nbt.Object{
			"type": stringTag("minecraft:noise"), "settings": stringTag("minecraft:overworld"),
			"biome_source": compoundTag(nbt.Object{
				"type": stringTag("minecraft:multi_noise"), "preset": stringTag("minecraft:overworld"),
			}),
		}
	}
	dimensions := nbt.Object{
		"minecraft:overworld": compoundTag(nbt.Object{
			"type": stringTag("minecraft:overworld"), "generator": compoundTag(overworldGenerator),
		}),
		"minecraft:the_nether": compoundTag(nbt.Object{
			"type": stringTag("minecraft:the_nether"),
			"generator": compoundTag(nbt.Object{
				"type": stringTag("minecraft:noise"), "settings": stringTag("minecraft:nether"),
				"biome_source": compoundTag(nbt.Object{
					"type": stringTag("minecraft:multi_noise"), "preset": stringTag("minecraft:nether"),
				}),
			}),
		}),
		"minecraft:the_end": compoundTag(nbt.Object{
			"type": stringTag("minecraft:the_end"),
			"generator": compoundTag(nbt.Object{
				"type": stringTag("minecraft:noise"), "settings": stringTag("minecraft:end"),
				"biome_source": compoundTag(nbt.Object{"type": stringTag("minecraft:the_end")}),
			}),
		}),
	}
	spawnY := int32(64)
	// New normal worlds enable supported starts and pieces. Coverage remains
	// explicit in AtlasWorldgen until all original structure families exist.
	features := int8(1)
	if flat {
		spawnY = -60
		features = 0
	}
	data := nbt.Object{
		"DataVersion": {Type: nbt.Int, Value: int32(3955)},
		"version":     {Type: nbt.Int, Value: int32(19133)},
		"Version": compoundTag(nbt.Object{
			"Id": {Type: nbt.Int, Value: int32(3955)}, "Name": stringTag("1.21.1"),
			"Series": stringTag("main"), "Snapshot": byteTag(0),
		}),
		"LevelName": stringTag(name), "initialized": byteTag(1),
		"LastPlayed": {Type: nbt.Long, Value: time.Now().UnixMilli()},
		"Time":       {Type: nbt.Long, Value: int64(0)}, "DayTime": {Type: nbt.Long, Value: int64(6000)},
		"SpawnX": {Type: nbt.Int, Value: spawnX}, "SpawnY": {Type: nbt.Int, Value: spawnY},
		"SpawnZ": {Type: nbt.Int, Value: spawnZ}, "SpawnAngle": {Type: nbt.Float, Value: float32(0)},
		"raining": byteTag(0), "thundering": byteTag(0),
		"Difficulty": byteTag(2), "DifficultyLocked": byteTag(0),
		"GameType": {Type: nbt.Int, Value: int32(1)}, "hardcore": byteTag(0),
		"allowCommands": byteTag(1), "GameRules": compoundTag(nbt.Object{}),
		"DataPacks": compoundTag(nbt.Object{
			"Enabled": stringList("vanilla"), "Disabled": emptyStrings,
		}),
		"enabled_features": stringList("minecraft:vanilla"),
		"WorldGenSettings": compoundTag(nbt.Object{
			"seed": {Type: nbt.Long, Value: seed}, "generate_features": byteTag(features),
			"bonus_chest": byteTag(0), "dimensions": compoundTag(dimensions),
		}),
	}
	if !flat {
		data["AtlasWorldgen"] = compoundTag(nbt.Object{
			"Version":    {Type: nbt.Int, Value: worldgen.OverworldGenerationVersion},
			"Stages":     stringList("noise", "biomes", "surface", "carvers", "supported_features", "supported_trees", "supported_vegetation"),
			"Complete":   byteTag(0),
			"SpawnReady": byteTag(0),
		})
	}
	return compoundTag(nbt.Object{"Data": compoundTag(data)})
}
