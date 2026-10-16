package world

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

var ErrNotSingleChest = errors.New("block is not a single chest")
var ErrChestBlocked = errors.New("chest is obstructed")
var ErrChestLocked = errors.New("chest is locked")

type ChestInventory struct {
	Position   BlockPosition
	Slots      [27]item.Stack
	CustomName string
}

func (v ChestInventory) Clone() ChestInventory {
	for i := range v.Slots {
		v.Slots[i] = v.Slots[i].Clone()
	}
	return v
}

type containerQuery struct {
	position BlockPosition
	luck     float32
	change   func(*ChestInventory) error
	result   chan containerResult
}
type containerResult struct {
	inventory ChestInventory
	err       error
}

// ReadChest lazily unpacks the original loot table and durably saves Items
// before the first snapshot can be delivered to a client.
func (w *World) ReadChest(position BlockPosition, luck float32) (ChestInventory, error) {
	return w.UpdateChest(position, luck, nil)
}

// UpdateChest runs against the latest actor-owned contents. Multiple players
// cannot overwrite a chest using stale snapshots. The callback is trusted
// server code and must not call back into the world or block on a player lock.
func (w *World) UpdateChest(position BlockPosition, luck float32, change func(*ChestInventory) error) (ChestInventory, error) {
	if position.Y < w.MinBuildHeight() || position.Y >= w.MaxBuildHeight() || math.IsNaN(float64(luck)) || math.IsInf(float64(luck), 0) {
		return ChestInventory{}, ErrNotSingleChest
	}
	cx, cz := position.X>>4, position.Z>>4
	if err := w.ensureChunk(context.Background(), cx, cz); err != nil {
		return ChestInventory{}, err
	}
	result := make(chan containerResult, 1)
	if err := w.post(context.Background(), cx, cz, containerQuery{position, luck, change, result}); err != nil {
		return ChestInventory{}, err
	}
	select {
	case value := <-result:
		return value.inventory, value.err
	case <-w.done:
		return ChestInventory{}, ErrWorldClosed
	}
}

func (r *Region) containerQuery(query containerQuery) {
	value, err := r.updateChest(query)
	query.result <- containerResult{value, err}
}
func (r *Region) updateChest(query containerQuery) (ChestInventory, error) {
	p := query.position
	ch := r.chunks[chunkKey{p.X >> 4, p.Z >> 4}]
	if ch == nil {
		return ChestInventory{}, ErrChunkNotLoaded
	}
	ch.lastAccess = time.Now()
	state, err := BlockStateByID(ch.GetBlock(int(p.X&15), int(p.Z&15), p.Y))
	if err != nil || (state.Name != "minecraft:chest" && state.Name != "minecraft:trapped_chest") || state.Properties["type"] != "single" {
		return ChestInventory{}, ErrNotSingleChest
	}
	blocked, err := chestRedstoneConductor(ch.GetBlock(int(p.X&15), int(p.Z&15), p.Y+1))
	if err != nil {
		return ChestInventory{}, err
	}
	if blocked {
		return ChestInventory{}, ErrChestBlocked
	}
	root, _ := ch.anvilRoot.Object()
	entities, _ := root["block_entities"].Elements()
	index := -1
	for i, tag := range entities {
		data, _ := tag.Object()
		x, xOK := data.Number("x")
		y, yOK := data.Number("y")
		z, zOK := data.Number("z")
		if xOK && yOK && zOK && x == int64(p.X) && y == int64(p.Y) && z == int64(p.Z) {
			index = i
			break
		}
	}
	data := nbt.Object{"id": {Type: nbt.String, Value: state.Name}, "x": {Type: nbt.Int, Value: p.X}, "y": {Type: nbt.Int, Value: p.Y}, "z": {Type: nbt.Int, Value: p.Z}}
	if index >= 0 {
		original, _ := entities[index].Object()
		for key, value := range original {
			data[key] = value
		}
		if lock, _ := data.Text("Lock"); lock != "" {
			return ChestInventory{}, ErrChestLocked
		}
	}
	value := ChestInventory{Position: p}
	value.CustomName, _ = data.Text("CustomName")
	changed := index < 0
	if table, _ := data.Text("LootTable"); table != "" {
		seed, _ := data.Number("LootTableSeed")
		value.Slots, err = FillChestLoot(table, seed, query.luck)
		if err != nil {
			return ChestInventory{}, err
		}
		delete(data, "LootTable")
		delete(data, "LootTableSeed")
		changed = true
	} else if tag, exists := data["Items"]; exists {
		entries, ok := tag.Elements()
		if !ok {
			return ChestInventory{}, fmt.Errorf("invalid chest Items list")
		}
		seen := [27]bool{}
		for _, tag := range entries {
			entry, ok := tag.Object()
			slot, valid := entry.Number("Slot")
			if !ok || !valid || slot < 0 || slot >= 27 || seen[slot] {
				return ChestInventory{}, fmt.Errorf("invalid or duplicate chest item slot")
			}
			seen[slot] = true
			stack, err := item.FromNBT(tag)
			if err != nil {
				return ChestInventory{}, err
			}
			value.Slots[slot] = stack
		}
	}
	if query.change != nil {
		before := value.Clone()
		if err := query.change(&value); err != nil {
			return before, err
		}
		// Keep immutable ownership and names controlled by the block entity.
		value.Position, value.CustomName = before.Position, before.CustomName
		for i, stack := range value.Slots {
			if _, valid := item.Name(stack.ID); !stack.Empty() && !valid {
				return before, fmt.Errorf("invalid server chest item at slot %d", i)
			}
			if stack.Count < 0 || stack.Count > stack.MaxCount() || stack.Count > 99 {
				return before, fmt.Errorf("invalid server chest stack at slot %d", i)
			}
		}
		for i, stack := range value.Slots {
			old := before.Slots[i]
			if stack.Count != old.Count || !stack.Empty() && (!stack.SameItem(old) || !stack.SameSavedFields(old)) {
				changed = true
				break
			}
		}
	}
	if !changed {
		return value.Clone(), nil
	}
	items := []nbt.Tag{}
	for i, stack := range value.Slots {
		if stack.Empty() {
			continue
		}
		tag, err := stack.ToNBT()
		if err != nil {
			return ChestInventory{}, err
		}
		entry, _ := tag.Object()
		entry["Slot"] = nbt.Tag{Type: nbt.Byte, Value: int8(i)}
		items = append(items, tag)
	}
	data["Items"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: items}
	next := append([]nbt.Tag(nil), entities...)
	if index < 0 {
		next = append(next, nbt.Tag{Type: nbt.Compound, Value: data})
	} else {
		next[index] = nbt.Tag{Type: nbt.Compound, Value: data}
	}
	previous, dirty := root["block_entities"], ch.dirty
	root["block_entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: next}
	ch.dirty = true
	if r.world.store != nil {
		if err := r.world.store.vanilla.persistChunk(ch); err != nil {
			root["block_entities"] = previous
			ch.dirty = dirty
			return ChestInventory{}, err
		}
	}
	return value.Clone(), nil
}
