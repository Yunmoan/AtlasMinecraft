package world

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"atlas/internal/proto"
)

// Numeric registry IDs and update tags are checked against the 1.21.1 JAR's
// ClientboundLevelChunkPacketData.BlockEntityInfo. Loot tables stay server-side.
var generatedBlockEntityTypes = map[string]int32{
	"minecraft:chest": 1, "minecraft:mob_spawner": 9, "minecraft:brushable_block": 39,
	"minecraft:dispenser": 5, "minecraft:brewing_stand": 11, "minecraft:sign": 7,
	"minecraft:beehive": 33, "minecraft:sculk_sensor": 34,
	"minecraft:sculk_catalyst": 36, "minecraft:sculk_shrieker": 37,
	"minecraft:end_portal": 13, "minecraft:end_gateway": 21,
}

func blockEntityUpdateTag(entity nbt.Object) nbt.Tag {
	id, _ := entity.Text("id")
	if id == "minecraft:sign" {
		text := func() nbt.Tag {
			messages := make([]nbt.Tag, 4)
			for i := range messages {
				messages[i] = nbt.Tag{Type: nbt.String, Value: `""`}
			}
			return nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
				"has_glowing_text": {Type: nbt.Byte, Value: int8(0)},
				"color":            {Type: nbt.String, Value: "black"},
				"messages":         {Type: nbt.List, ListType: nbt.String, Value: messages},
			}}
		}
		data := nbt.Object{"front_text": text(), "back_text": text(), "is_waxed": {Type: nbt.Byte, Value: int8(0)}}
		for _, name := range []string{"front_text", "back_text", "is_waxed"} {
			if tag, ok := entity[name]; ok {
				data[name] = tag
			}
		}
		return nbt.Tag{Type: nbt.Compound, Value: data}
	}
	if id == "minecraft:end_gateway" {
		data := nbt.Object{"Age": {Type: nbt.Long, Value: int64(0)}}
		for _, name := range []string{"Age", "exit_portal", "ExactTeleport"} {
			if tag, ok := entity[name]; ok {
				data[name] = tag
			}
		}
		return nbt.Tag{Type: nbt.Compound, Value: data}
	}
	if id != "minecraft:mob_spawner" {
		return nbt.Tag{Type: nbt.End}
	}
	data := make(nbt.Object)
	for _, name := range []string{"Delay", "MinSpawnDelay", "MaxSpawnDelay", "SpawnCount", "MaxNearbyEntities", "RequiredPlayerRange", "SpawnRange", "SpawnData"} {
		if tag, ok := entity[name]; ok {
			data[name] = tag
		}
	}
	if len(data) == 0 {
		return nbt.Tag{Type: nbt.End}
	}
	return nbt.Tag{Type: nbt.Compound, Value: data}
}

func writeChunkBlockEntities(w *proto.Writer, ch *Chunk) {
	root, _ := ch.anvilRoot.Object()
	entities, _ := root["block_entities"].Elements()
	var entries [][]byte
	for _, tag := range entities {
		entity, ok := tag.Object()
		if !ok {
			continue
		}
		id, _ := entity.Text("id")
		typeID, supported := generatedBlockEntityTypes[id]
		x, xOK := entity.Number("x")
		y, yOK := entity.Number("y")
		z, zOK := entity.Number("z")
		if !supported || !xOK || !yOK || !zOK || x>>4 != int64(ch.cx) || z>>4 != int64(ch.cz) || y < int64(ch.minY) || y >= int64(ch.maxY) {
			continue
		}
		update, err := nbt.MarshalNetwork(blockEntityUpdateTag(entity))
		if err != nil {
			continue
		}
		entry := proto.NewWriter().U8(byte((x&15)<<4 | z&15)).I16(int16(y)).VarInt(typeID).Raw(update)
		entries = append(entries, entry.Payload())
	}
	w.VarInt(int32(len(entries)))
	for _, entry := range entries {
		w.Raw(entry)
	}
}

// Vanilla removes an entity when a different block replaces its owning block.
// State changes within the same block (for example chest facing) retain it.
func (c *Chunk) removeReplacedBlockEntity(lx, lz int, y int32, old, next uint32) {
	root, _ := c.anvilRoot.Object()
	entities, _ := root["block_entities"].Elements()
	if len(entities) == 0 {
		return
	}
	a, errA := blockstate.BlockStateByID(old)
	b, errB := blockstate.BlockStateByID(next)
	if errA != nil || errB != nil || a.Name == b.Name {
		return
	}
	var kept []nbt.Tag
	for _, tag := range entities {
		e, _ := tag.Object()
		x, xOK := e.Number("x")
		ey, yOK := e.Number("y")
		z, zOK := e.Number("z")
		if xOK && yOK && zOK && x == int64(c.cx)*16+int64(lx) && ey == int64(y) && z == int64(c.cz)*16+int64(lz) {
			continue
		}
		kept = append(kept, tag)
	}
	if kept == nil {
		kept = []nbt.Tag{}
	}
	root["block_entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: kept}
}
