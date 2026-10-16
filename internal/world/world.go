package world

import (
	"atlas/internal/item"
	"atlas/internal/light"
	"atlas/internal/proto"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// World shards ownership of chunk data across region actors and runs a pool
// of chunk-generation workers. Callers never touch chunk state directly;
// they post messages to the owning region's mailbox.
type World struct {
	mu            sync.Mutex
	regions       map[chunkKey]*Region
	generation    *generationQueue
	hub           HubView
	store         *Store
	stateMu       sync.RWMutex
	saveMu        sync.Mutex
	blockUpdateMu sync.Mutex // edits plus neighbor propagation; actors never acquire it
	meta          worldMeta
	seed          int64
	workers       int
	loadChunk     func(int32, int32) (*Chunk, error)
	done          chan struct{}
	wg            sync.WaitGroup
	closeOnce     sync.Once
	closeErr      error
	lighting      *light.Engine
	lightErr      error
	loadedChunks  atomic.Int64
	activeJobs    atomic.Int64
	dimension     dimensionSettings
}

var ErrWorldClosed = errors.New("world is closed")

type genJob struct {
	cx, cz   int32
	region   *Region
	canceled bool
}

func New(hub HubView, genWorkers int) *World {
	return newWorld(hub, genWorkers, nil, worldMeta{DayTime: 6000, Weather: "clear"})
}

// Open loads an existing Java Edition save. An incompatible or corrupt save
// is an error; it is never silently regenerated.
func Open(hub HubView, genWorkers int, dir string) (*World, error) {
	s, err := OpenStore(dir)
	if err != nil {
		return nil, err
	}
	m, err := s.LoadMeta()
	if err != nil {
		return nil, err
	}
	return newWorld(hub, genWorkers, s, m), nil
}

func newWorld(hub HubView, genWorkers int, store *Store, meta worldMeta) *World {
	if genWorkers < 1 {
		genWorkers = 1
	}
	w := &World{
		regions:    make(map[chunkKey]*Region),
		generation: newGenerationQueue(),
		hub:        hub,
		store:      store,
		meta:       meta,
		workers:    genWorkers,
		done:       make(chan struct{}),
	}
	w.dimension, _ = builtinDimension("minecraft:overworld")
	if store != nil {
		w.dimension = store.vanilla.dimension
		w.seed = store.vanilla.seed
		w.loadChunk = store.vanilla.loadChunk
		if store.vanilla.normal != nil {
			store.vanilla.normal.SetTerrainWorkers(genWorkers)
			store.vanilla.normal.SetStructureDayTimeSource(func() int64 { _, day := w.Time(); return day })
			store.vanilla.normal.SetStructureGameTimeSource(func() int64 { age, _ := w.Time(); return age })
		}
		if store.vanilla.dimensionGenerator != nil {
			store.vanilla.dimensionGenerator.SetTerrainWorkers(genWorkers)
		}
	} else {
		w.loadChunk = func(int32, int32) (*Chunk, error) { return nil, ErrVanillaGenerationUnavailable }
	}
	w.lighting, w.lightErr = light.New(int(w.dimension.minY), int(w.dimension.maxY-w.dimension.minY), w.dimension.hasSky)
	for i := 0; i < genWorkers; i++ {
		w.wg.Add(1)
		go w.genWorker()
	}
	return w
}

func (w *World) genWorker() {
	defer w.wg.Done()
	for {
		j, ok := w.generation.next(w.done)
		if !ok {
			return
		}
		select {
		case <-w.done:
			return
		default:
		}
		if j.canceled {
			select {
			case j.region.mailbox <- generationCanceled{j.cx, j.cz}:
			case <-w.done:
				return
			}
			continue
		}
		w.activeJobs.Add(1)
		started := time.Now()
		ch, err := w.loadChunk(j.cx, j.cz)
		terrainReady := time.Now()
		if err == nil && w.lighting != nil {
			position := light.Position{X: j.cx, Z: j.cz}
			var updates map[light.Position]*light.Data
			updates, err = w.lighting.Install(position, denseLightStates(ch), ch.lights)
			if err == nil {
				ch.lights = w.lighting.Snapshot(position)
				if updates[position] != nil {
					ch.dirty = true
				}
				if ch.dirty && w.store != nil {
					err = w.store.vanilla.persistChunk(ch)
				}
				w.notifyLight(updates, position)
			}
		} else if err == nil && w.lightErr != nil {
			err = w.lightErr
		}
		slog.Debug("chunk ready", "x", j.cx, "z", j.cz, "terrain_ms", terrainReady.Sub(started).Milliseconds(), "lighting_and_save_ms", time.Since(terrainReady).Milliseconds(), "error", err)
		w.activeJobs.Add(-1)
		select {
		case j.region.mailbox <- chunkGenerated{cx: j.cx, cz: j.cz, chunk: ch, err: err}:
		case <-w.done:
			return
		}
	}
}

// post starts the owning actor under the lifecycle lock. Once Close begins,
// no new actors can be added to the wait group or messages accepted.
func (w *World) post(ctx context.Context, cx, cz int32, message any) error {
	k := chunkKey{cx >> RegionChunkBits, cz >> RegionChunkBits}
	w.mu.Lock()
	select {
	case <-w.done:
		w.mu.Unlock()
		return ErrWorldClosed
	case <-ctx.Done():
		w.mu.Unlock()
		return ctx.Err()
	default:
	}
	r := w.regions[k]
	if r == nil {
		r = newRegion(w, k.x, k.z)
		w.regions[k] = r
		w.wg.Add(1)
		go r.run()
	}
	w.mu.Unlock()
	select {
	case r.mailbox <- message:
		return nil
	case <-w.done:
		return ErrWorldClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RequestChunk asynchronously delivers the chunk's packet to sink.
func (w *World) RequestChunk(sink PacketSink, cx, cz int32) {
	_ = w.post(context.Background(), cx, cz, chunkRequest{cx: cx, cz: cz, to: sink})
}

// LoadChunk waits for the owning region to return the current chunk packet.
func (w *World) LoadChunk(cx, cz int32) ([]byte, error) {
	result := make(chan chunkResult, 1)
	if err := w.post(context.Background(), cx, cz, chunkLoad{cx: cx, cz: cz, result: result}); err != nil {
		return nil, err
	}
	select {
	case res := <-result:
		// Public callers own the returned bytes; network sinks consume immutable
		// shared cache buffers without an extra copy.
		return append([]byte(nil), res.payload...), res.err
	case <-w.done:
		return nil, ErrWorldClosed
	}
}

// SurfaceY returns the first air block above the highest block in a column.
func (w *World) SurfaceY(x, z int32) (int32, error) {
	result := make(chan surfaceResult, 1)
	if err := w.post(context.Background(), x>>4, z>>4, surfaceQuery{x: x, z: z, result: result}); err != nil {
		return 0, err
	}
	select {
	case res := <-result:
		return res.y, res.err
	case <-w.done:
		return 0, ErrWorldClosed
	}
}

// SpawnXZ returns the shared world spawn column. New players without saved
// data join there; /setworldspawn can move it at runtime.
func (w *World) SpawnXZ() (int32, int32) {
	x, _, z, _ := w.SpawnPoint()
	return x, z
}

// SpawnPoint returns the shared world spawn block and its facing angle. The
// Y coordinate is the player's feet position.
func (w *World) SpawnPoint() (int32, int32, int32, float32) {
	if w.store == nil {
		return 0, 0, 0, 0
	}
	return w.store.vanilla.spawnPoint()
}

// SetSpawn moves the shared world spawn and persists it in level.dat. It
// reports whether the stored point changed so callers can skip redundant
// spawn-position broadcasts.
func (w *World) SetSpawn(x, y, z int32, angle float32) (bool, error) {
	if w.store == nil {
		return false, errors.New("no Java Edition save is loaded")
	}
	if x < -30000000 || x >= 30000000 || z < -30000000 || z >= 30000000 {
		return false, errors.New("X/Z must be >= -30000000 and < 30000000")
	}
	if y < w.MinBuildHeight() || y >= w.MaxBuildHeight() {
		return false, fmt.Errorf("Y must be >= %d and < %d", w.MinBuildHeight(), w.MaxBuildHeight())
	}
	return w.store.vanilla.setSpawn(x, y, z, angle)
}

func (w *World) Seed() int64 { return w.seed }

func (w *World) IsFlat() bool {
	return w.store != nil && w.store.vanilla.flat != nil
}

// Vanilla's BiomeManager.obfuscateSeed hashes the little-endian seed and
// sends the first eight hash bytes as a little-endian signed long.
func (w *World) HashedSeed() int64 {
	var seed [8]byte
	binary.LittleEndian.PutUint64(seed[:], uint64(w.seed))
	hash := sha256.Sum256(seed[:])
	return int64(binary.LittleEndian.Uint64(hash[:8]))
}

func (w *World) GenerationWarning() string {
	if w.store == nil {
		return "no Java Edition save is loaded"
	}
	flat := w.store.vanilla.flat
	if flat == nil {
		if w.store.vanilla.dimensionGenerator != nil {
			structures := "structure generation is disabled in this save"
			if w.store.vanilla.dimensionGenerator.StructuresEnabled() {
				structures = "remaining structures are incomplete"
				if w.DimensionName() == "minecraft:the_nether" {
					structures = "fortresses, Nether fossils, ruined portals and Bastions are generated"
				} else if w.DimensionName() == "minecraft:the_end" {
					structures = "End cities with chests, Shulkers and Elytra item frames are generated"
				}
			}
			return "dimension terrain, surface, biomes, Nether caves and all dimension feature plans are generated (Nether 37/37, End 5/5); " + structures + "; entities, dragon fight, neighboring decoration ordering and generation-time lighting/ticks remain incomplete"
		}
		if w.store.vanilla.normal != nil {
			structures := "supported pyramids, swamp huts, igloos, buried treasure, shipwrecks, ocean ruins, mineshafts, trail ruins, trial chambers, ancient cities and six ruined-portal variants are generated"
			if !w.store.vanilla.normal.StructuresEnabled() {
				structures = "structure generation is disabled in this save"
			}
			return "noise, biomes, surface, caves, all 151 Overworld feature plans and 18 initial animal types are generated; " + structures + "; remaining structures, entity behavior, neighboring decoration ordering and generation-time lighting/ticks remain incomplete; beehives use persisted independent randomness"
		}
		return "missing noise-world chunks cannot be generated exactly"
	}
	if !flat.exact {
		return "missing flat-world chunks with structures or features cannot be generated exactly"
	}
	return ""
}

// GenerationMode describes the path actually used for a missing chunk.
func (w *World) GenerationMode() string {
	if w.store == nil {
		return "no save loaded"
	}
	flat := w.store.vanilla.flat
	if flat == nil {
		if w.store.vanilla.dimensionGenerator != nil {
			features := "5/5 End placed features"
			structures := "structure generation disabled"
			if w.store.vanilla.dimensionGenerator.StructuresEnabled() {
				structures = "remaining structures incomplete"
			}
			if w.DimensionName() == "minecraft:the_nether" {
				features = "37/37 Nether placed features"
				if w.store.vanilla.dimensionGenerator.StructuresEnabled() {
					structures = "fortresses, fossils, ruined portals and Bastions"
				}
			} else if w.DimensionName() == "minecraft:the_end" && w.store.vanilla.dimensionGenerator.StructuresEnabled() {
				structures = "End cities, chests, Shulkers and Elytra item frames"
			}
			return fmt.Sprintf("Minecraft 1.21.1 %s (pure Go terrain, surface, biomes, Nether caves and %s; %s; entities/ticks incomplete)", w.DimensionName(), features, structures)
		}
		if w.store.vanilla.normal != nil {
			return "Minecraft 1.21.1 Overworld (pure Go terrain, caves, 151/151 feature plans and supported structures; remaining structure families and generation stages incomplete)"
		}
		return "Minecraft 1.21.1 noise save (existing chunks only)"
	}
	if !flat.exact {
		return "Minecraft 1.21.1 flat save (existing chunks only; features or structures enabled)"
	}
	return "Minecraft 1.21.1 flat layers (features and structures disabled)"
}

// PreGenerateSpawn generates the spawn chunk up front. Without this, a joining
// client out-races the parallel
// generation queue: the player spawns over not-yet-loaded ground, falls
// through it, and ends up buried inside solid stone (all-black screen).
func (w *World) PreGenerateSpawn(cx, cz int32) error {
	// Go through the actor so repeated preload calls cannot replace a newer
	// edited chunk with an independently loaded snapshot.
	return w.ensureChunk(context.Background(), cx, cz)
}

// SpawnChunk delivers the player's current chunk first, then marks it sent.
func (w *World) SpawnChunk(p PacketSink, cx, cz int32, mark func(cx, cz int32)) error {
	payload, err := w.LoadChunk(cx, cz)
	if err != nil {
		return err
	}
	if deliverChunk(p, cx, cz, payload) {
		mark(cx, cz)
	}
	return nil
}

// SetBlock posts a block edit; the owning region applies it and broadcasts
// the change to every viewer of that position.
func (w *World) SetBlock(x, y, z int32, state uint32) error {
	if y < w.MinBuildHeight() || y >= w.MaxBuildHeight() {
		return errors.New("block Y outside world bounds")
	}
	if state > MaxBlockState {
		return errors.New("unknown 1.21.1 block state")
	}
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	changed, err := w.editBlock(setBlockMsg{x: x, y: y, z: z, state: state})
	if err != nil {
		return err
	}
	return w.updateBlockNeighbors(changed)
}

func (w *World) Time() (age, dayTime int64) {
	w.stateMu.RLock()
	defer w.stateMu.RUnlock()
	return w.meta.Age, w.meta.DayTime
}

func (w *World) AdvanceTime(ticks int64) {
	w.stateMu.Lock()
	w.meta.Age += ticks
	w.meta.DayTime += ticks
	w.stateMu.Unlock()
}

func (w *World) SetDayTime(ticks int64) {
	w.stateMu.Lock()
	w.meta.DayTime = ticks
	w.stateMu.Unlock()
}

func (w *World) Weather() string {
	w.stateMu.RLock()
	defer w.stateMu.RUnlock()
	return w.meta.Weather
}

func (w *World) LoadPlayer(uuid [16]byte) (PlayerPosition, bool, error) {
	if w.store == nil {
		return PlayerPosition{}, false, nil
	}
	return w.store.LoadPlayer(uuid)
}

func (w *World) LoadInventory(uuid [16]byte) (item.Inventory, error) {
	if w.store == nil {
		return item.Inventory{}, nil
	}
	return w.store.LoadInventory(uuid)
}
func (w *World) SavePlayerState(uuid [16]byte, p PlayerPosition, inventory item.Inventory) error {
	if w.store == nil {
		return nil
	}
	return w.store.SavePlayerState(uuid, p, inventory)
}

func (w *World) SavePlayer(uuid [16]byte, p PlayerPosition) error {
	if w.store == nil {
		return nil
	}
	return w.store.SavePlayer(uuid, p)
}

// SetPlayerSpawn persists a player's personal respawn point as the vanilla
// SpawnX/SpawnY/SpawnZ/SpawnAngle/SpawnForced/SpawnDimension tags.
func (w *World) SetPlayerSpawn(uuid [16]byte, x, y, z int32, angle float32) error {
	if w.store == nil {
		return errors.New("no Java Edition save is loaded")
	}
	return w.store.SavePlayerSpawn(uuid, x, y, z, angle)
}

func (w *World) SetWeather(weather string) error {
	if weather != "clear" && weather != "rain" && weather != "thunder" {
		return errors.New("weather must be clear, rain, or thunder")
	}
	w.stateMu.Lock()
	w.meta.Weather = weather
	w.stateMu.Unlock()
	return nil
}

// Save flushes world metadata. Block edits are durable when SetBlock returns.
func (w *World) Save() error {
	w.saveMu.Lock()
	defer w.saveMu.Unlock()
	if w.store == nil {
		return nil
	}
	w.mu.Lock()
	regions := make([]*Region, 0, len(w.regions))
	for _, r := range w.regions {
		regions = append(regions, r)
	}
	w.mu.Unlock()
	for _, r := range regions {
		result := make(chan error, 1)
		select {
		case r.mailbox <- saveRegion{result: result}:
		case <-w.done:
			return ErrWorldClosed
		}
		select {
		case err := <-result:
			if err != nil {
				return err
			}
		case <-w.done:
			return ErrWorldClosed
		}
	}
	w.stateMu.RLock()
	m := w.meta
	w.stateMu.RUnlock()
	return w.store.SaveMeta(m)
}

func (w *World) Close() error {
	w.closeOnce.Do(func() {
		w.mu.Lock()
		close(w.done)
		w.mu.Unlock()
		// Generators may finish their current chunk, but no writes/goroutines
		// may outlive Close. Actors stop before their maps are read directly.
		w.wg.Wait()
		w.saveMu.Lock()
		defer w.saveMu.Unlock()
		if w.store != nil {
			for _, r := range w.regions {
				w.closeErr = errors.Join(w.closeErr, r.saveChunks())
			}
			w.stateMu.RLock()
			meta := w.meta
			w.stateMu.RUnlock()
			w.closeErr = errors.Join(w.closeErr, w.store.SaveMeta(meta), w.store.Close())
		}
	})
	return w.closeErr
}

// buildChunkDataPacket encodes a full SChunkData (0x27) payload: header,
// heightmaps NBT, section data, block entities, and sky light.
func buildChunkDataPacket(cx, cz int32, ch *Chunk) []byte {
	data := ch.EncodeSections()
	w := proto.NewWriterCap(len(data) + 6*2048 + 512)
	w.I32(cx).I32(cz)
	w.Raw(ch.HeightmapNBT())
	w.VarInt(int32(len(data))).Raw(data)
	writeChunkBlockEntities(w, ch)

	lighting := ch.lights
	if lighting == nil {
		lighting = detachedLight(ch)
	}
	writeLightData(w, lighting)
	return w.Payload()
}
