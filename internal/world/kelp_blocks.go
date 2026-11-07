package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

const kelpBreakJournal = "AtlasPendingKelpBreaks"

// Block.destroyBlock emits the break event, drops the item, replaces the block
// with its fluid legacy state (flags 3), then emits BLOCK_DESTROY. The durable
// intent gives the item one identity across retries and source-chunk reloads.
func (r *fluidRuntime) DestroyKelpBlock(p worldgen.FeaturePosition, state worldgen.SurfaceBlockState) error {
	key := fmt.Sprintf("%d,%d,%d", p.X, p.Y, p.Z)
	var entry nbt.Object
	if err := r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(_ *Region, ch *Chunk) error {
		root, _ := ch.anvilRoot.Object()
		pending, _ := root.Object(kelpBreakJournal)
		entry, _ = pending.Object(key)
		return nil
	}); err != nil {
		return err
	}
	if entry == nil {
		v := r.world.store.vanilla
		v.mu.Lock()
		root, _ := v.root.Object()
		data, _ := root.Object("Data")
		rules, _ := data.Object("GameRules")
		value, exists := rules.Text("doTileDrops")
		v.mu.Unlock()
		drops := !exists || strings.EqualFold(value, "true")
		if err := r.world.ensureTickRandom(); err != nil {
			return err
		}
		var entityRandom *worldgen.LegacyRandom
		if drops {
			var seed [8]byte
			if _, err := rand.Read(seed[:]); err != nil {
				return err
			}
			entityRandom = worldgen.NewLegacyRandom(int64(binary.LittleEndian.Uint64(seed[:])))
		}
		item, err := worldgen.NewKelpBlockDropEntity(p, drops, r.world.tickRandom, entityRandom)
		if err != nil {
			return err
		}
		id, err := BlockStateID(state.Name, state.Properties)
		if err != nil {
			return err
		}
		entry = nbt.Object{"Source": {Type: nbt.IntArray, Value: []int32{p.X, p.Y, p.Z}}, "Expected": {Type: nbt.Int, Value: int32(id)}}
		if item != nil {
			entry["Entity"] = nbt.Tag{Type: nbt.Compound, Value: item}
		}
		if err := r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(actor *Region, ch *Chunk) error {
			return updateKelpJournal(actor, ch, key, entry)
		}); err != nil {
			return err
		}
	}
	return r.finishKelpBreak(key, entry)
}

func updateKelpJournal(actor *Region, ch *Chunk, key string, entry nbt.Object) error {
	root, _ := ch.anvilRoot.Object()
	old, existed := root[kelpBreakJournal]
	pending, _ := old.Object()
	next := make(nbt.Object, len(pending)+1)
	for k, v := range pending {
		next[k] = v
	}
	if entry == nil {
		delete(next, key)
	} else {
		next[key] = nbt.Tag{Type: nbt.Compound, Value: entry}
	}
	if len(next) == 0 {
		delete(root, kelpBreakJournal)
	} else {
		root[kelpBreakJournal] = nbt.Tag{Type: nbt.Compound, Value: next}
	}
	if err := actor.world.store.vanilla.persistChunk(ch); err != nil {
		if existed {
			root[kelpBreakJournal] = old
		} else {
			delete(root, kelpBreakJournal)
		}
		return err
	}
	return nil
}

func (r *fluidRuntime) finishKelpBreak(key string, entry nbt.Object) error {
	source, _ := entry["Source"].Value.([]int32)
	expected, ok := entry.Number("Expected")
	if len(source) != 3 || !ok || expected < 0 || expected > math.MaxUint32 {
		return fmt.Errorf("invalid kelp break intent")
	}
	p := worldgen.FeaturePosition{X: source[0], Y: source[1], Z: source[2]}
	if !r.InWorld(p) {
		return fmt.Errorf("kelp journal is outside build bounds")
	}
	if key != fmt.Sprintf("%d,%d,%d", p.X, p.Y, p.Z) {
		return fmt.Errorf("kelp journal source differs from key")
	}
	state, err := BlockStateByID(uint32(expected))
	if err != nil {
		return err
	}
	if state.Name != "minecraft:kelp" && state.Name != "minecraft:kelp_plant" {
		return fmt.Errorf("kelp break intent has another block")
	}
	var entity *WorldEntity
	if tag, exists := entry["Entity"]; exists {
		data, ok := tag.Object()
		if !ok {
			return fmt.Errorf("invalid kelp drop NBT")
		}
		e, err := worldEntityFromNBT(data)
		if err != nil {
			return err
		}
		item, _ := data.Object("Item")
		id, _ := item.Text("id")
		count, _ := item.Number("count")
		if e.Type != "minecraft:item" || id != "minecraft:kelp" || count != 1 || math.Floor(e.Position[0]) != float64(p.X) || math.Floor(e.Position[1]) != float64(p.Y) || math.Floor(e.Position[2]) != float64(p.Z) {
			return fmt.Errorf("kelp drop does not match source")
		}
		entity = &e
	}
	current, err := r.BlockAt(p)
	if err != nil {
		return err
	}
	id, err := BlockStateID(current.Name, current.Properties)
	if err != nil {
		return err
	}
	if id == uint32(expected) {
		if err := r.FluidLevelEvent(p, 2001, int32(expected)); err != nil {
			return err
		}
	}
	if entity != nil {
		if err := r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(actor *Region, ch *Chunk) error {
			for _, existing := range ch.entities {
				if existing.UUID == entity.UUID {
					return nil
				}
			}
			return persistFallingEntities(actor, ch, append(append([]WorldEntity(nil), ch.entities...), *entity))
		}); err != nil {
			return err
		}
	}
	if id == uint32(expected) {
		fluid, err := worldgen.FluidStateForBlock(worldgen.SurfaceBlockState{Name: state.Name, Properties: state.Properties})
		if err != nil {
			return err
		}
		if err := r.SetFluidBlock(p, fluid.LegacyBlock(), 3); err != nil {
			return err
		}
		if err := r.FallingGameEvent(p, "minecraft:block_destroy"); err != nil {
			return err
		}
	}
	return r.world.fallingEntityOperation(p.X>>4, p.Z>>4, func(actor *Region, ch *Chunk) error { return updateKelpJournal(actor, ch, key, nil) })
}

func (r *fluidRuntime) recoverKelpBreaks() error {
	o := r.world.scheduledTicks
	o.mu.Lock()
	keys := make([]chunkKey, 0, len(o.loadOrder))
	for _, ch := range o.loadOrder {
		keys = append(keys, chunkKey{ch.x, ch.z})
	}
	o.mu.Unlock()
	for _, key := range keys {
		var pending nbt.Object
		if err := r.world.fallingEntityOperation(key.x, key.z, func(_ *Region, ch *Chunk) error {
			root, _ := ch.anvilRoot.Object()
			items, _ := root.Object(kelpBreakJournal)
			var err error
			pending, err = cloneEntityData(items)
			return err
		}); err != nil {
			return err
		}
		for id, tag := range pending {
			entry, ok := tag.Object()
			if !ok {
				return fmt.Errorf("invalid kelp journal record")
			}
			source, _ := entry["Source"].Value.([]int32)
			if len(source) != 3 || source[0]>>4 != key.x || source[2]>>4 != key.z {
				return fmt.Errorf("kelp journal belongs to another chunk")
			}
			if err := r.finishKelpBreak(id, entry); err != nil {
				return err
			}
		}
	}
	return nil
}
