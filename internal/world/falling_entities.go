package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"sync"
)

type fallingEntityKey struct {
	uuid [16]byte
	x, z int32
}
type fallingEntityOrder struct {
	mu      sync.Mutex
	entries []fallingEntityKey
}

func (o *fallingEntityOrder) register(ch *Chunk) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, e := range ch.entities {
		if e.Type != "minecraft:falling_block" {
			continue
		}
		found := false
		for _, key := range o.entries {
			if key.uuid == e.UUID {
				found = true
				break
			}
		}
		if !found {
			o.entries = append(o.entries, fallingEntityKey{e.UUID, ch.cx, ch.cz})
		}
	}
}
func (o *fallingEntityOrder) unregister(x, z int32) {
	o.mu.Lock()
	defer o.mu.Unlock()
	out := o.entries[:0]
	for _, e := range o.entries {
		if e.x != x || e.z != z {
			out = append(out, e)
		}
	}
	o.entries = out
}
func (o *fallingEntityOrder) remove(uuid [16]byte) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i, key := range o.entries {
		if key.uuid == uuid {
			o.entries = append(o.entries[:i], o.entries[i+1:]...)
			return
		}
	}
}

// Actor operations are synchronous under World.blockUpdateMu. Entity NBT and
// chunk journals are only read or changed by the owning region actor.
type fallingEntityOperation struct {
	x, z   int32
	run    func(*Region, *Chunk) error
	result chan error
}

func (w *World) fallingEntityOperation(x, z int32, run func(*Region, *Chunk) error) error {
	result := make(chan error, 1)
	if err := w.post(context.Background(), x, z, fallingEntityOperation{x, z, run, result}); err != nil {
		return err
	}
	select {
	case err := <-result:
		return err
	case <-w.done:
		return ErrWorldClosed
	}
}
func (r *Region) runFallingEntityOperation(op fallingEntityOperation) {
	ch := r.chunks[chunkKey{op.x, op.z}]
	if ch == nil {
		op.result <- ErrChunkNotLoaded
		return
	}
	op.result <- op.run(r, ch)
}

func persistFallingEntities(r *Region, ch *Chunk, entities []WorldEntity) error {
	v := r.world.store.vanilla
	name, root, _, err := v.readEntityChunk(ch.cx, ch.cz)
	if err != nil {
		return err
	}
	list := make([]nbt.Tag, len(entities))
	for i, e := range entities {
		list[i] = nbt.Tag{Type: nbt.Compound, Value: e.Data}
	}
	root["Entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: list}
	if err := v.regions.WriteChunk(v.entityDirectory(), ch.cx, ch.cz, name, nbt.Tag{Type: nbt.Compound, Value: root}); err != nil {
		return err
	}
	ch.entities = entities
	r.world.fallingEntities.register(ch)
	return nil
}

const fallingJournal = "AtlasPendingFallingEntities"

func (r *fluidRuntime) SpawnFallingBlock(spawn worldgen.FallingBlockSpawn) error {
	if r.world.store == nil {
		return ErrVanillaGenerationUnavailable
	}
	p := spawn.Source
	var entry nbt.Object
	err := r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(actor *Region, ch *Chunk) error {
		root, _ := ch.anvilRoot.Object()
		pending, _ := root[fallingJournal].Elements()
		for _, tag := range pending {
			obj, _ := tag.Object()
			source, _ := obj["Source"].Value.([]int32)
			if len(source) == 3 && source[0] == p.X && source[1] == p.Y && source[2] == p.Z {
				entry = obj
				return nil
			}
		}
		data, err := cloneEntityData(spawn.Data)
		if err != nil {
			return err
		}
		var uuid [16]byte
		if _, err := rand.Read(uuid[:]); err != nil {
			return err
		}
		uuid[6] = uuid[6]&15 | 64
		uuid[8] = uuid[8]&63 | 128
		words := make([]int32, 4)
		for i := range words {
			words[i] = int32(binary.BigEndian.Uint32(uuid[i*4:]))
		}
		data["UUID"] = nbt.Tag{Type: nbt.IntArray, Value: words}
		data["id"] = nbt.Tag{Type: nbt.String, Value: "minecraft:falling_block"}
		replacement, err := BlockStateID(spawn.Replacement.Name, spawn.Replacement.Properties)
		if err != nil {
			return err
		}
		entry = nbt.Object{"Source": {Type: nbt.IntArray, Value: []int32{p.X, p.Y, p.Z}}, "Expected": {Type: nbt.Int, Value: int32(ch.GetBlock(int(p.X&15), int(p.Z&15), p.Y))}, "Replacement": {Type: nbt.Int, Value: int32(replacement)}, "Entity": {Type: nbt.Compound, Value: data}}
		old, existed := root[fallingJournal]
		root[fallingJournal] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: append(pending, nbt.Tag{Type: nbt.Compound, Value: entry})}
		if err := actor.world.store.vanilla.persistChunk(ch); err != nil {
			if existed {
				root[fallingJournal] = old
			} else {
				delete(root, fallingJournal)
			}
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return r.finishFallingSpawn(entry)
}

// The durable intent precedes source removal. A retry/reload reuses its UUID,
// and a committed entity is deduplicated before clearing the journal.
func (r *fluidRuntime) finishFallingSpawn(entry nbt.Object) error {
	source, _ := entry["Source"].Value.([]int32)
	if len(source) != 3 {
		return fmt.Errorf("invalid falling source journal")
	}
	p := worldgen.FeaturePosition{X: source[0], Y: source[1], Z: source[2]}
	expected, expectedOK := entry.Number("Expected")
	replacement, replacementOK := entry.Number("Replacement")
	if !expectedOK || !replacementOK || expected < 0 || replacement < 0 || expected > math.MaxUint32 || replacement > math.MaxUint32 {
		return fmt.Errorf("invalid falling journal block states")
	}
	if _, err := BlockStateByID(uint32(expected)); err != nil {
		return err
	}
	replacementState, err := BlockStateByID(uint32(replacement))
	if err != nil {
		return err
	}
	data, _ := entry.Object("Entity")
	entity, err := worldEntityFromNBT(data)
	if err != nil {
		return err
	}
	if entity.Type != "minecraft:falling_block" || entity.Position != [3]float64{float64(p.X) + 0.5, float64(p.Y), float64(p.Z) + 0.5} {
		return fmt.Errorf("falling journal entity does not match source")
	}
	// Validate the entire durable intent before its first world mutation. A
	// malformed entity must not remove the source and leave an unrecoverable hole.
	current, err := r.BlockAt(p)
	if err != nil {
		return err
	}
	id, err := BlockStateID(current.Name, current.Properties)
	if err != nil {
		return err
	}
	if id == uint32(expected) {
		if err := r.SetFluidBlock(p, worldgen.SurfaceBlockState{Name: replacementState.Name, Properties: replacementState.Properties}, 3); err != nil {
			return err
		}
	}
	return r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(actor *Region, ch *Chunk) error {
		found := false
		for _, e := range ch.entities {
			if e.UUID == entity.UUID {
				found = true
				break
			}
		}
		if !found {
			entities := append(append([]WorldEntity(nil), ch.entities...), entity)
			if err := persistFallingEntities(actor, ch, entities); err != nil {
				return err
			}
		}
		root, _ := ch.anvilRoot.Object()
		old := root[fallingJournal]
		pending, _ := old.Elements()
		var remaining []nbt.Tag
		for _, tag := range pending {
			item, _ := tag.Object()
			eData, _ := item.Object("Entity")
			e, err := worldEntityFromNBT(eData)
			if err != nil {
				return err
			}
			if e.UUID != entity.UUID {
				remaining = append(remaining, tag)
			}
		}
		if len(remaining) == 0 {
			delete(root, fallingJournal)
		} else {
			root[fallingJournal] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: remaining}
		}
		if err := actor.world.store.vanilla.persistChunk(ch); err != nil {
			root[fallingJournal] = old
			return err
		}
		return nil
	})
}

func (r *fluidRuntime) recoverFallingSpawns() error {
	o := r.world.scheduledTicks
	o.mu.Lock()
	keys := make([]chunkKey, 0, len(o.loadOrder))
	for _, ch := range o.loadOrder {
		keys = append(keys, chunkKey{ch.x, ch.z})
	}
	o.mu.Unlock()
	for _, key := range keys {
		var pending []nbt.Object
		if err := r.world.fallingEntityOperation(key.x, key.z, func(_ *Region, ch *Chunk) error {
			root, _ := ch.anvilRoot.Object()
			items, _ := root[fallingJournal].Elements()
			for _, tag := range items {
				obj, _ := tag.Object()
				data, err := cloneEntityData(obj)
				if err != nil {
					return err
				}
				pending = append(pending, data)
			}
			return nil
		}); err != nil {
			return err
		}
		for _, entry := range pending {
			if err := r.finishFallingSpawn(entry); err != nil {
				return err
			}
		}
	}
	return nil
}

// Entity ticking has its own ticket predicate: a block-ticking halo is not
// necessarily entity-ticking. Invoke after TickScheduled for each elapsed tick.
func (w *World) TickFallingEntities(eligible func(int32, int32) bool) error {
	if eligible == nil {
		return fmt.Errorf("falling entities require explicit entity-ticking eligibility")
	}
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	r := &fluidRuntime{world: w, settings: w.fluidSettings()}
	if err := r.recoverFallingSpawns(); err != nil {
		return err
	}
	w.fallingEntities.mu.Lock()
	keys := append([]fallingEntityKey(nil), w.fallingEntities.entries...)
	w.fallingEntities.mu.Unlock()
	for _, key := range keys {
		if !eligible(key.x, key.z) {
			continue
		}
		var entity WorldEntity
		found := false
		if err := w.fallingEntityOperation(key.x, key.z, func(_ *Region, ch *Chunk) error {
			for _, e := range ch.entities {
				if e.UUID == key.uuid {
					var err error
					entity = e
					entity.Data, err = cloneEntityData(e.Data)
					found = true
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		if !found {
			continue
		}
		motion, _ := entity.Data["Motion"].Elements()
		if len(motion) != 3 {
			return fmt.Errorf("invalid falling entity motion")
		}
		for _, axis := range []int{0, 2} {
			speed, ok := motion[axis].Value.(float64)
			if !ok {
				return fmt.Errorf("invalid falling motion")
			}
			coordinate := int32(math.Floor((entity.Position[axis] + speed) / 16))
			expected := key.x
			if axis == 2 {
				expected = key.z
			}
			if coordinate != expected {
				return fmt.Errorf("falling entity chunk migration is not implemented")
			}
		}
		r.fallingDrops = nil
		data, removed, err := worldgen.TickFallingEntity(r, entity.Data)
		if err != nil {
			return err
		}
		if !removed && len(r.fallingDrops) > 0 {
			return fmt.Errorf("falling entity requested an item without removal")
		}
		next, err := worldEntityFromNBT(data)
		if err != nil {
			return err
		}
		if int32(math.Floor(next.Position[0]/16)) != key.x || int32(math.Floor(next.Position[2]/16)) != key.z {
			return fmt.Errorf("falling entity chunk migration is not implemented")
		}
		next.Revision = entity.Revision + 1
		if err := w.fallingEntityOperation(key.x, key.z, func(actor *Region, ch *Chunk) error {
			entities := make([]WorldEntity, 0, len(ch.entities))
			for _, e := range ch.entities {
				if e.UUID != key.uuid {
					entities = append(entities, e)
				} else if !removed {
					entities = append(entities, next)
				}
			}
			// One native entity-region replacement commits both source entity
			// removal and its item. Failed writes retain the airborne snapshot;
			// retry/reload cannot append a second committed item.
			entities = append(entities, r.fallingDrops...)
			return persistFallingEntities(actor, ch, entities)
		}); err != nil {
			return err
		}
		if removed {
			w.fallingEntities.remove(key.uuid)
		}
	}
	return nil
}

func (r *fluidRuntime) MaxBuildHeight() int32                                   { return r.world.MaxBuildHeight() }
func (r *fluidRuntime) FallingGameEvent(worldgen.FeaturePosition, string) error { return nil } // No vibration consumer yet; same boundary as fluid game events.
func (r *fluidRuntime) DropFallingBlockItem(position [3]float64, state worldgen.SurfaceBlockState) error {
	enabled, err := r.FallingEntityDropsEnabled()
	if err != nil || !enabled {
		return err
	}
	if err := r.world.ensureTickRandom(); err != nil {
		return err
	}
	var seed [8]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return err
	}
	entityRandom := worldgen.NewLegacyRandom(int64(binary.LittleEndian.Uint64(seed[:])))
	tag, err := worldgen.NewFallingBlockDropEntity(position, state, r.world.tickRandom, entityRandom)
	if err != nil {
		return err
	}
	entity, err := worldEntityFromNBT(tag)
	if err != nil {
		return err
	}
	r.fallingDrops = append(r.fallingDrops, entity)
	return nil
}

func (r *fluidRuntime) FallingEntityDropsEnabled() (bool, error) {
	v := r.world.store.vanilla
	v.mu.Lock()
	root, _ := v.root.Object()
	data, _ := root.Object("Data")
	rules, _ := data.Object("GameRules")
	drops, exists := rules.Text("doEntityDrops")
	v.mu.Unlock()
	return !exists || strings.EqualFold(drops, "true"), nil
}
