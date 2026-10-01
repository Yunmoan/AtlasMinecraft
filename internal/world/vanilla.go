package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"atlas/internal/registry"
	"atlas/internal/worldgen"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// VanillaSave retains all level.dat tags, including fields Atlas does not
// understand. The version check prevents accidental writes to other formats.
type VanillaSave struct {
	mu                     sync.Mutex
	regions                anvil.Store
	dir                    string
	name                   string
	root                   nbt.Tag
	seed                   int64
	spawnX, spawnY, spawnZ int32
	flat                   *flatSettings
	normal                 *worldgen.OverworldGenerator
}

type flatLayer struct {
	state  uint32
	height int32
}
type flatSettings struct {
	biome  int32
	layers []flatLayer
	exact  bool
}

var ErrVanillaGenerationUnavailable = errors.New("exact vanilla world generation is unavailable")

func openVanilla(dir string) (*VanillaSave, error) {
	data, err := readGzip(filepath.Join(dir, "level.dat"))
	if err != nil {
		return nil, err
	}
	name, root, err := nbt.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse level.dat: %w", err)
	}
	o, _ := root.Object()
	level, ok := o.Object("Data")
	if !ok {
		return nil, errors.New("level.dat has no Data compound")
	}
	version, ok := level.Number("DataVersion")
	if !ok || version != 3955 {
		return nil, fmt.Errorf("unsupported level.dat DataVersion %d (requires 1.21.1/3955)", version)
	}
	s := &VanillaSave{dir: dir, name: name, root: root}
	settings, ok := level.Object("WorldGenSettings")
	if !ok {
		return nil, errors.New("WorldGenSettings missing")
	}
	s.seed, _ = settings.Number("seed")
	x, _ := level.Number("SpawnX")
	y, _ := level.Number("SpawnY")
	z, _ := level.Number("SpawnZ")
	s.spawnX, s.spawnY, s.spawnZ = int32(x), int32(y), int32(z)
	dims, ok := settings.Object("dimensions")
	if !ok {
		return nil, errors.New("world dimensions missing")
	}
	overworld, ok := dims.Object("minecraft:overworld")
	if !ok {
		return nil, errors.New("overworld generator missing")
	}
	gen, ok := overworld.Object("generator")
	if !ok {
		return nil, errors.New("overworld generator data missing")
	}
	if kind, _ := gen.Text("type"); kind == "minecraft:flat" {
		flat, err := parseFlatSettings(gen)
		if err != nil {
			return nil, err
		}
		s.flat = flat
	} else if kind == "minecraft:noise" && supportsNormalGenerator(gen) && usesBuiltinDatapacks(level) {
		s.normal, err = worldgen.NewOverworldGenerator(s.seed)
		if err != nil {
			return nil, fmt.Errorf("initialize overworld generator: %w", err)
		}
	}
	return s, nil
}

func supportsNormalGenerator(generator nbt.Object) bool {
	settings, _ := generator.Text("settings")
	biome, ok := generator.Object("biome_source")
	if !ok || settings != "minecraft:overworld" {
		return false
	}
	kind, _ := biome.Text("type")
	preset, _ := biome.Text("preset")
	return kind == "minecraft:multi_noise" && preset == "minecraft:overworld"
}

func usesBuiltinDatapacks(level nbt.Object) bool {
	// A custom pack may override even minecraft:overworld. Loading existing
	// chunks is supported, but do not extend them using unrelated built-ins.
	if packs, ok := level.Object("DataPacks"); ok {
		if enabled, ok := packs["Enabled"].Elements(); ok {
			for _, pack := range enabled {
				name, ok := pack.Text()
				if !ok || name != "vanilla" {
					return false
				}
			}
		}
	}
	return true
}

func parseFlatSettings(generator nbt.Object) (*flatSettings, error) {
	settings, ok := generator.Object("settings")
	if !ok {
		return nil, errors.New("flat generator settings missing")
	}
	biomeName, ok := settings.Text("biome")
	if !ok {
		return nil, errors.New("flat biome missing")
	}
	biome, err := registry.BiomeID(biomeName)
	if err != nil {
		return nil, err
	}
	entries, ok := settings["layers"].Elements()
	if !ok {
		return nil, errors.New("flat layers missing")
	}
	f := &flatSettings{biome: biome, exact: true}
	for _, entry := range entries {
		layer, ok := entry.Object()
		if !ok {
			return nil, errors.New("flat layer is not compound")
		}
		name, ok := layer.Text("block")
		if !ok {
			return nil, errors.New("flat layer block missing")
		}
		state, err := BlockStateID(name, nil)
		if err != nil {
			return nil, err
		}
		height, ok := layer.Number("height")
		if !ok || height < 0 || height > WorldHeight {
			return nil, errors.New("invalid flat layer height")
		}
		f.layers = append(f.layers, flatLayer{state: state, height: int32(height)})
	}
	if features, _ := settings.Number("features"); features != 0 {
		f.exact = false
	}
	if lakes, _ := settings.Number("lakes"); lakes != 0 {
		f.exact = false
	}
	if structures, ok := settings["structure_overrides"].Elements(); ok && len(structures) > 0 {
		f.exact = false
	}
	return f, nil
}

func (v *VanillaSave) loadChunk(cx, cz int32) (*Chunk, error) {
	name, root, exists, err := v.regions.ReadChunk(filepath.Join(v.dir, "region"), cx, cz)
	if err != nil {
		return nil, err
	}
	if exists {
		object, _ := root.Object()
		if status, _ := object.Text("Status"); status != "minecraft:full" && status != "full" {
			return nil, fmt.Errorf("%w: existing chunk %d,%d is unfinished (%s)", ErrVanillaGenerationUnavailable, cx, cz, status)
		}
		return chunkFromAnvil(name, root, cx, cz)
	}
	if v.normal != nil {
		chunk, err := v.generateNormal(cx, cz)
		if err != nil {
			return nil, err
		}
		if err := v.persistChunk(chunk); err != nil {
			return nil, err
		}
		return chunk, nil
	}
	if v.flat == nil {
		return nil, fmt.Errorf("%w: missing noise-world chunk %d,%d", ErrVanillaGenerationUnavailable, cx, cz)
	}
	if !v.flat.exact {
		return nil, fmt.Errorf("%w: missing flat-world chunk %d,%d with structures/features", ErrVanillaGenerationUnavailable, cx, cz)
	}
	chunk, err := v.generateFlat(cx, cz)
	if err != nil {
		return nil, err
	}
	if err := v.persistChunk(chunk); err != nil {
		return nil, err
	}
	return chunk, nil
}

func (v *VanillaSave) persistChunk(chunk *Chunk) error {
	// Chunk state belongs to its actor (or its generating worker before
	// publication). Only metadata needs the save-wide mutex.
	name, root, err := chunk.updateAnvilBlocks()
	if err != nil {
		return err
	}
	if err := v.regions.WriteChunk(filepath.Join(v.dir, "region"), chunk.cx, chunk.cz, name, root); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	levelRoot, _ := v.root.Object()
	level, _ := levelRoot.Object("Data")
	if marker, ok := level.Object("AtlasWorldgen"); ok && chunk.cx == v.spawnX>>4 && chunk.cz == v.spawnZ>>4 {
		if ready, _ := marker.Number("SpawnReady"); ready == 0 {
			previousY, previousReady := level["SpawnY"], marker["SpawnReady"]
			level["SpawnY"] = nbt.Tag{Type: nbt.Int, Value: chunk.Height(int(v.spawnX&15), int(v.spawnZ&15))}
			marker["SpawnReady"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
			if err := writeGzipNBTAtomic(filepath.Join(v.dir, "level.dat"), v.name, v.root); err != nil {
				level["SpawnY"], marker["SpawnReady"] = previousY, previousReady
				return err
			}
		}
	}
	chunk.dirty = false
	chunk.dirtySections = [NumSections]bool{}
	return nil
}

func (v *VanillaSave) generateFlat(cx, cz int32) (*Chunk, error) {
	c := NewChunk(cx, cz)
	for si := range c.biomes {
		for i := range c.biomes[si] {
			c.biomes[si][i] = v.flat.biome
		}
	}
	y := int32(MinY)
	for _, layer := range v.flat.layers {
		for dy := int32(0); dy < layer.height && y < MaxY; dy++ {
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					c.SetBlock(x, z, y, layer.state)
				}
			}
			y++
		}
	}
	c.dirty = false
	c.dirtySections = [NumSections]bool{}
	if err := c.initializeAnvilRoot(); err != nil {
		return nil, err
	}
	return c, nil
}

func (v *VanillaSave) loadMeta() (worldMeta, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	o, _ := v.root.Object()
	data, _ := o.Object("Data")
	age, _ := data.Number("Time")
	day, _ := data.Number("DayTime")
	rain, _ := data.Number("raining")
	thunder, _ := data.Number("thundering")
	weather := "clear"
	if rain != 0 {
		weather = "rain"
	}
	if thunder != 0 {
		weather = "thunder"
	}
	return worldMeta{Age: age, DayTime: day, Weather: weather}, nil
}

func (v *VanillaSave) saveMeta(m worldMeta) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	o, _ := v.root.Object()
	data, _ := o.Object("Data")
	data["Time"] = nbt.Tag{Type: nbt.Long, Value: m.Age}
	data["DayTime"] = nbt.Tag{Type: nbt.Long, Value: m.DayTime}
	rain, thunder := int8(0), int8(0)
	if m.Weather != "clear" {
		rain = 1
	}
	if m.Weather == "thunder" {
		thunder = 1
	}
	data["raining"] = nbt.Tag{Type: nbt.Byte, Value: rain}
	data["thundering"] = nbt.Tag{Type: nbt.Byte, Value: thunder}
	return writeGzipNBTAtomic(filepath.Join(v.dir, "level.dat"), v.name, v.root)
}

func writeGzipNBTAtomic(path, name string, root nbt.Tag) error {
	encoded, err := nbt.Marshal(name, root)
	if err != nil {
		return err
	}
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	if _, err = zw.Write(encoded); err != nil {
		zw.Close()
		return err
	}
	if err = zw.Close(); err != nil {
		return err
	}
	return writeVanillaFileAtomic(path, compressed.Bytes())
}

// Keep the previous complete file as the vanilla .dat_old recovery copy.
func writeVanillaFileAtomic(path string, data []byte) error {
	if previous, err := os.ReadFile(path); err == nil {
		if err := replaceFileAtomic(path+"_old", previous); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return replaceFileAtomic(path, data)
}

func replaceFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "atlas-vanilla-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Sync()
	}
	if e := tmp.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func readGzip(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	b, err := io.ReadAll(io.LimitReader(zr, (32<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 32<<20 {
		return nil, errors.New("gzip NBT exceeds 32 MiB limit")
	}
	return b, nil
}
