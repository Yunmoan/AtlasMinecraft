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
	spawnAngle             float32
	flat                   *flatSettings
	normal                 *worldgen.OverworldGenerator
	dimensionGenerator     *worldgen.DimensionGenerator
	dimension              dimensionSettings
	regionDir              string
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
	return openVanillaDimension(dir, "minecraft:overworld")
}

func openVanillaDimension(dir, dimensionName string) (*VanillaSave, error) {
	dimension, err := builtinDimension(dimensionName)
	if err != nil {
		return nil, err
	}
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
	s := &VanillaSave{dir: dir, name: name, root: root, dimension: dimension, regionDir: filepath.Join(dir, dimension.regionSubdir, "region")}
	settings, ok := level.Object("WorldGenSettings")
	if !ok {
		return nil, errors.New("WorldGenSettings missing")
	}
	s.seed, _ = settings.Number("seed")
	x, _ := level.Number("SpawnX")
	y, _ := level.Number("SpawnY")
	z, _ := level.Number("SpawnZ")
	if tag, ok := level["SpawnAngle"]; ok {
		if angle, ok := tag.Value.(float32); ok {
			s.spawnAngle = angle
		}
	}
	s.spawnX, s.spawnY, s.spawnZ = int32(x), int32(y), int32(z)
	dims, ok := settings.Object("dimensions")
	if !ok {
		return nil, errors.New("world dimensions missing")
	}
	overworld, ok := dims.Object(dimension.name)
	if !ok {
		return nil, fmt.Errorf("%s generator missing", dimension.name)
	}
	gen, ok := overworld.Object("generator")
	if !ok {
		return nil, errors.New("overworld generator data missing")
	}
	if dimension.name != "minecraft:overworld" {
		kind, _ := gen.Text("type")
		if kind == "minecraft:noise" && supportsDimensionGenerator(gen, dimension) && usesBuiltinDatapacks(level) {
			generateStructures, _ := settings.Number("generate_features")
			s.dimensionGenerator, err = worldgen.NewDimensionGeneratorWithOptions(s.seed, dimension.generator, worldgen.DimensionGeneratorOptions{GenerateStructures: generateStructures != 0})
			if err != nil {
				return nil, fmt.Errorf("initialize %s generator: %w", dimension.name, err)
			}
		}
		return s, nil
	}
	if kind, _ := gen.Text("type"); kind == "minecraft:flat" {
		flat, err := parseFlatSettings(gen)
		if err != nil {
			return nil, err
		}
		s.flat = flat
	} else if kind == "minecraft:noise" && supportsNormalGenerator(gen) && usesBuiltinDatapacks(level) {
		salt, changed, saltErr := savedBeehiveShuffleSalt(level)
		if saltErr != nil {
			return nil, fmt.Errorf("initialize independent tree randomness: %w", saltErr)
		}
		generateStructures, _ := settings.Number("generate_features")
		s.normal, err = worldgen.NewOverworldGeneratorWithOptions(s.seed, worldgen.OverworldGeneratorOptions{BeehiveShuffleSalt: salt, GenerateStructures: generateStructures != 0})
		if err != nil {
			return nil, fmt.Errorf("initialize overworld generator: %w", err)
		}
		if changed {
			if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), s.name, s.root); err != nil {
				return nil, fmt.Errorf("persist independent tree randomness: %w", err)
			}
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

// spawnPoint returns the shared world spawn. setSpawn may replace it at any
// time, so the fields are read under the save mutex.
func (v *VanillaSave) spawnPoint() (int32, int32, int32, float32) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.spawnX, v.spawnY, v.spawnZ, v.spawnAngle
}

// spawnUninitialized reports whether this save still waits for its initial
// spawn scan: Atlas-created noise worlds carry an AtlasWorldgen marker with
// SpawnReady=0 until the boot-time scan has refined the spawn point. Imported
// and already-initialized saves never report uninitialized.
func (v *VanillaSave) spawnUninitialized() bool {
	if v.dimension.name != "minecraft:overworld" {
		return false
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	levelRoot, _ := v.root.Object()
	level, _ := levelRoot.Object("Data")
	marker, ok := level.Object("AtlasWorldgen")
	if !ok {
		return false
	}
	ready, ok := marker.Number("SpawnReady")
	return ok && ready == 0
}

// initializeSpawn stores the scanned initial spawn (with vanilla's zero
// facing angle) and completes the AtlasWorldgen SpawnReady marker. Unlike
// setSpawn it always writes, even when the position matches the placeholder.
func (v *VanillaSave) initializeSpawn(x, y, z int32) error {
	if v.dimension.name != "minecraft:overworld" {
		return errors.New("shared world spawn is owned by the overworld")
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.spawnX, v.spawnY, v.spawnZ, v.spawnAngle = x, y, z, 0
	if err := v.writeSpawnLocked(); err != nil {
		return err
	}
	return v.markSpawnReadyLocked()
}

func (v *VanillaSave) writeSpawnLocked() error {
	if v.dimension.name != "minecraft:overworld" {
		return errors.New("shared world spawn is owned by the overworld")
	}
	levelRoot, _ := v.root.Object()
	level, _ := levelRoot.Object("Data")
	level["SpawnX"] = nbt.Tag{Type: nbt.Int, Value: v.spawnX}
	level["SpawnY"] = nbt.Tag{Type: nbt.Int, Value: v.spawnY}
	level["SpawnZ"] = nbt.Tag{Type: nbt.Int, Value: v.spawnZ}
	level["SpawnAngle"] = nbt.Tag{Type: nbt.Float, Value: v.spawnAngle}
	return writeGzipNBTAtomic(filepath.Join(v.dir, "level.dat"), v.name, v.root)
}

func (v *VanillaSave) markSpawnReadyLocked() error {
	levelRoot, _ := v.root.Object()
	level, _ := levelRoot.Object("Data")
	marker, ok := level.Object("AtlasWorldgen")
	if !ok {
		return nil
	}
	if ready, _ := marker.Number("SpawnReady"); ready == 1 {
		return nil
	}
	marker["SpawnReady"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	return writeGzipNBTAtomic(filepath.Join(v.dir, "level.dat"), v.name, v.root)
}

// setSpawn stores an explicit shared world spawn in memory and level.dat,
// reporting whether the stored point changed. An explicit spawn also completes
// the AtlasWorldgen SpawnReady marker so the initial spawn scan can no longer
// override the chosen position.
func (v *VanillaSave) setSpawn(x, y, z int32, angle float32) (bool, error) {
	if v.dimension.name != "minecraft:overworld" {
		return false, errors.New("shared world spawn is owned by the overworld")
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if x == v.spawnX && y == v.spawnY && z == v.spawnZ && angle == v.spawnAngle {
		return false, v.markSpawnReadyLocked()
	}
	v.spawnX, v.spawnY, v.spawnZ, v.spawnAngle = x, y, z, angle
	if err := v.writeSpawnLocked(); err != nil {
		return true, err
	}
	return true, v.markSpawnReadyLocked()
}

func (v *VanillaSave) loadChunk(cx, cz int32) (*Chunk, error) {
	chunk, err := v.loadBlockChunk(cx, cz)
	if err != nil {
		return nil, err
	}
	if err = v.loadAndRealizeEntities(chunk); err != nil {
		return nil, fmt.Errorf("load entities for %d,%d: %w", cx, cz, err)
	}
	return chunk, nil
}

func (v *VanillaSave) loadBlockChunk(cx, cz int32) (*Chunk, error) {
	name, root, exists, err := v.regions.ReadChunk(v.regionDir, cx, cz)
	if err != nil {
		return nil, err
	}
	if exists {
		object, _ := root.Object()
		if status, _ := object.Text("Status"); status != "minecraft:full" && status != "full" {
			return nil, fmt.Errorf("%w: existing chunk %d,%d is unfinished (%s)", ErrVanillaGenerationUnavailable, cx, cz, status)
		}
		return chunkFromAnvilBounds(name, root, cx, cz, v.dimension.minY, v.dimension.maxY, v.dimension.hasSky)
	}
	if v.dimensionGenerator != nil {
		chunk, err := v.generateDimension(cx, cz)
		if err != nil {
			return nil, err
		}
		if err := v.persistChunk(chunk); err != nil {
			return nil, err
		}
		return chunk, nil
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
	if err := v.regions.WriteChunk(v.regionDir, chunk.cx, chunk.cz, name, root); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	chunk.dirty = false
	clear(chunk.dirtySections)
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
	clear(c.dirtySections)
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
	if v.dimension.name != "minecraft:overworld" {
		return nil
	}
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
