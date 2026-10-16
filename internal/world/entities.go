package world

import (
	"atlas/internal/nbt"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"sync"
)

// WorldEntity is an immutable snapshot. Data contains the complete saved NBT,
// including unsupported fields, rather than only the client-visible fields.
type WorldEntity struct {
	UUID     [16]byte
	Type     string
	Position [3]float64
	Rotation [2]float32
	Data     nbt.Object
	// LoadedFromDisk distinguishes NBT load side effects from a freshly
	// finalized actor. It is runtime state and is never written into saved NBT.
	LoadedFromDisk bool
	Revision       uint64 // runtime mutation order; never saved to entity NBT
}

//go:embed entity_defaults_1.21.1.txt
var officialEntityDefaults string

//go:embed animal_defaults_1.21.1.txt
var officialAnimalDefaults string

//go:embed other_animal_defaults_1.21.1.txt
var officialOtherAnimalDefaults string

//go:embed bastion_actor_defaults_1.21.1.txt
var officialBastionActorDefaults string

//go:embed drowned_actor_defaults_1.21.1.txt
var officialDrownedActorDefaults string

//go:embed outpost_actor_defaults_1.21.1.txt
var officialOutpostActorDefaults string

//go:embed vehicle_defaults_1.21.1.txt
var officialVehicleDefaults string

//go:embed end_city_actor_defaults_1.21.1.txt
var officialEndCityActorDefaults string

//go:embed village_actor_defaults_1.21.1.txt
var officialVillageActorDefaults string

type entityQuery struct {
	cx, cz int32
	result chan entityResult
}
type entityResult struct {
	entities []WorldEntity
	err      error
}

func (w *World) EntitiesInChunk(cx, cz int32) ([]WorldEntity, error) {
	if err := w.ensureChunk(context.Background(), cx, cz); err != nil {
		return nil, err
	}
	result := make(chan entityResult, 1)
	if err := w.post(context.Background(), cx, cz, entityQuery{cx, cz, result}); err != nil {
		return nil, err
	}
	select {
	case res := <-result:
		return res.entities, res.err
	case <-w.done:
		return nil, ErrWorldClosed
	}
}

func (r *Region) entitiesInChunk(query entityQuery) {
	ch := r.chunks[chunkKey{query.cx, query.cz}]
	if ch == nil {
		query.result <- entityResult{err: ErrChunkNotLoaded}
		return
	}
	result := make([]WorldEntity, len(ch.entities))
	for i, e := range ch.entities {
		data, err := cloneEntityData(e.Data)
		if err != nil {
			query.result <- entityResult{err: err}
			return
		}
		e.Data = data
		result[i] = e
	}
	query.result <- entityResult{entities: result}
}

func cloneEntityData(data nbt.Object) (nbt.Object, error) {
	encoded, err := nbt.MarshalNetwork(nbt.Tag{Type: nbt.Compound, Value: data})
	if err != nil {
		return nil, err
	}
	tag, _, err := nbt.ParseNetwork(encoded)
	if err != nil {
		return nil, err
	}
	object, _ := tag.Object()
	return object, nil
}

// Keep encoded defaults immutable and indexed. Parsing the selected snapshot
// gives each actor its own nested maps and slices without scanning/decoding all
// embedded constructor snapshots every time a generation request is realized.
var generatedEntityDefaultIndex = sync.OnceValues(func() (map[string][]byte, error) {
	index := make(map[string][]byte)
	for _, source := range []string{officialEntityDefaults, officialAnimalDefaults, officialOtherAnimalDefaults, officialBastionActorDefaults, officialDrownedActorDefaults, officialOutpostActorDefaults, officialVehicleDefaults, officialEndCityActorDefaults, officialVillageActorDefaults} {
		for _, line := range strings.Split(source, "\n") {
			fields := strings.Fields(line)
			if len(fields) != 2 {
				continue
			}
			if _, exists := index[fields[0]]; exists {
				continue
			}
			encoded, err := hex.DecodeString(fields[1])
			if err != nil {
				return nil, fmt.Errorf("decode official entity default %s: %w", fields[0], err)
			}
			index[fields[0]] = encoded
		}
	}
	return index, nil
})

func defaultGeneratedEntity(id string) (nbt.Object, bool, error) {
	index, err := generatedEntityDefaultIndex()
	if err != nil {
		return nil, false, err
	}
	encoded, supported := index[id]
	if !supported {
		return nil, false, nil
	}
	tag, used, err := nbt.ParseNetwork(encoded)
	if err != nil || used != len(encoded) {
		return nil, true, fmt.Errorf("decode official entity default %s: %v", id, err)
	}
	object, ok := tag.Object()
	if !ok {
		return nil, true, fmt.Errorf("entity default is not compound")
	}
	return object, true, nil
}

func entityUUIDTag(uuid [16]byte) nbt.Tag {
	parts := make([]int32, 4)
	for i := range parts {
		parts[i] = int32(binary.BigEndian.Uint32(uuid[i*4:]))
	}
	return nbt.Tag{Type: nbt.IntArray, Value: parts}
}

func generatedEntityUUID(seed int64, dimension string, pending nbt.Object) [16]byte {
	id, _ := pending.Text("id")
	x, _ := pending.Number("x")
	y, _ := pending.Number("y")
	z, _ := pending.Number("z")
	flag, _ := pending.Text("PieceFlag")
	// Identity survives a crash between the entities and block-chunk writes.
	// It does not consume or perturb the terrain/feature random stream.
	hash := sha256.Sum256([]byte(fmt.Sprintf("atlas-generated-entity\x00%d\x00%s\x00%s\x00%d,%d,%d\x00%s", seed, dimension, id, x, y, z, flag)))
	var uuid [16]byte
	copy(uuid[:], hash[:16])
	uuid[6] = uuid[6]&15 | 0x50
	uuid[8] = uuid[8]&63 | 0x80
	return uuid
}

func entityPositionTag(position [3]float64) nbt.Tag {
	values := make([]nbt.Tag, 3)
	for i, v := range position {
		values[i] = nbt.Tag{Type: nbt.Double, Value: v}
	}
	return nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: values}
}

func realizeGeneratedEntity(seed int64, dimension string, pending nbt.Object) (WorldEntity, bool, error) {
	id, ok := pending.Text("id")
	if !ok {
		return WorldEntity{}, false, fmt.Errorf("pending entity id missing")
	}
	data, supported, err := defaultGeneratedEntity(id)
	if err != nil || !supported {
		return WorldEntity{}, supported, err
	}
	var position [3]float64
	for i, key := range []string{"x", "y", "z"} {
		value, ok := pending.Number(key)
		if !ok {
			return WorldEntity{}, true, fmt.Errorf("pending entity %s missing", key)
		}
		position[i] = float64(value)
		if i != 1 {
			position[i] += .5
		}
	}
	uuid := generatedEntityUUID(seed, dimension, pending)
	data["Pos"] = entityPositionTag(position)
	data["UUID"] = entityUUIDTag(uuid)
	if persistent, _ := pending.Number("Persistent"); persistent != 0 && id != "minecraft:end_crystal" && id != "minecraft:chest_minecart" {
		data["PersistenceRequired"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	}
	if extra, ok := pending.Object("Data"); ok {
		copy, err := cloneEntityData(extra)
		if err != nil {
			return WorldEntity{}, true, err
		}
		for name, value := range copy {
			if name != "UUID" && name != "id" {
				if name == "attributes" {
					data[name] = mergeEntityAttributes(data[name], value)
					continue
				}
				data[name] = value
			}
		}
	}
	if id == "minecraft:chest_minecart" {
		if table, _ := data.Text("LootTable"); table != "" {
			delete(data, "Items")
		}
	}
	assignGeneratedPassengerUUIDs(data, uuid)
	entity, err := worldEntityFromNBT(data)
	return entity, true, err
}

func assignGeneratedPassengerUUIDs(data nbt.Object, parent [16]byte) {
	passengers, _ := data["Passengers"].Elements()
	for i, tag := range passengers {
		child, ok := tag.Object()
		if !ok {
			continue
		}
		id, _ := child.Text("id")
		hash := sha256.Sum256([]byte(fmt.Sprintf("atlas-generated-passenger\x00%x\x00%d\x00%s", parent, i, id)))
		var uuid [16]byte
		copy(uuid[:], hash[:16])
		uuid[6] = uuid[6]&15 | 0x50
		uuid[8] = uuid[8]&63 | 0x80
		child["UUID"] = entityUUIDTag(uuid)
		assignGeneratedPassengerUUIDs(child, uuid)
	}
}

// EntityPassengers returns independent saved snapshots from the vanilla
// nested Passengers list. Anvil keeps the vehicle as the root entity.
func EntityPassengers(entity WorldEntity) ([]WorldEntity, error) {
	values, _ := entity.Data["Passengers"].Elements()
	result := make([]WorldEntity, 0, len(values))
	for _, tag := range values {
		data, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("saved passenger is not compound")
		}
		copy, err := cloneEntityData(data)
		if err != nil {
			return nil, err
		}
		passenger, err := worldEntityFromNBT(copy)
		if err != nil {
			return nil, err
		}
		passenger.LoadedFromDisk = entity.LoadedFromDisk
		result = append(result, passenger)
	}
	return result, nil
}

func mergeEntityAttributes(base, extra nbt.Tag) nbt.Tag {
	values, ok := base.Elements()
	if !ok {
		return extra
	}
	added, ok := extra.Elements()
	if !ok {
		return extra
	}
	for _, attribute := range added {
		object, _ := attribute.Object()
		id, _ := object.Text("id")
		replaced := false
		for i, existing := range values {
			data, _ := existing.Object()
			name, _ := data.Text("id")
			if name == id {
				values[i] = attribute
				replaced = true
				break
			}
		}
		if !replaced {
			values = append(values, attribute)
		}
	}
	return nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: values}
}

func worldEntityFromNBT(data nbt.Object) (WorldEntity, error) {
	var entity WorldEntity
	entity.Type, _ = data.Text("id")
	entity.Data = data
	parts, ok := data["UUID"].Value.([]int32)
	if entity.Type == "" || !ok || data["UUID"].Type != nbt.IntArray || len(parts) != 4 {
		return entity, fmt.Errorf("saved entity missing id/UUID")
	}
	for i, value := range parts {
		binary.BigEndian.PutUint32(entity.UUID[i*4:], uint32(value))
	}
	position, ok := data["Pos"].Elements()
	if !ok || len(position) != 3 || data["Pos"].ListType != nbt.Double {
		return entity, fmt.Errorf("saved entity %s invalid Pos", entity.Type)
	}
	for i, tag := range position {
		value, ok := tag.Value.(float64)
		if !ok || math.IsNaN(value) || math.IsInf(value, 0) {
			return entity, fmt.Errorf("invalid entity position")
		}
		entity.Position[i] = value
	}
	rotation, ok := data["Rotation"].Elements()
	if !ok || len(rotation) != 2 || data["Rotation"].ListType != nbt.Float {
		return entity, fmt.Errorf("saved entity %s invalid Rotation", entity.Type)
	}
	for i, tag := range rotation {
		value, ok := tag.Value.(float32)
		if !ok || math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return entity, fmt.Errorf("invalid entity rotation")
		}
		entity.Rotation[i] = value
	}
	return entity, nil
}

func (v *VanillaSave) entityDirectory() string {
	return filepath.Join(v.dir, v.dimension.regionSubdir, "entities")
}

func (v *VanillaSave) readEntityChunk(cx, cz int32) (string, nbt.Object, []WorldEntity, error) {
	name, root, exists, err := v.regions.ReadChunk(v.entityDirectory(), cx, cz)
	if err != nil {
		return "", nil, nil, err
	}
	if !exists {
		return "", nbt.Object{"DataVersion": {Type: nbt.Int, Value: int32(3955)}, "Position": {Type: nbt.IntArray, Value: []int32{cx, cz}}, "Entities": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{}}}, nil, nil
	}
	object, ok := root.Object()
	version, _ := object.Number("DataVersion")
	position, pOK := object["Position"].Value.([]int32)
	if !ok || version != 3955 || !pOK || len(position) != 2 || position[0] != cx || position[1] != cz {
		return "", nil, nil, fmt.Errorf("invalid entity chunk %d,%d position/version", cx, cz)
	}
	list, ok := object["Entities"].Elements()
	if !ok || object["Entities"].ListType != nbt.Compound {
		return "", nil, nil, fmt.Errorf("entity chunk missing Entities list")
	}
	var result []WorldEntity
	seen := map[[16]byte]bool{}
	for _, tag := range list {
		data, ok := tag.Object()
		if !ok {
			return "", nil, nil, fmt.Errorf("entity is not compound")
		}
		entity, err := worldEntityFromNBT(data)
		if err != nil {
			return "", nil, nil, err
		}
		entity.LoadedFromDisk = true
		if seen[entity.UUID] {
			return "", nil, nil, fmt.Errorf("duplicate entity UUID")
		}
		seen[entity.UUID] = true
		result = append(result, entity)
	}
	return name, object, result, nil
}

// Entities are committed before pending requests or piece flags are changed.
// Replaying a surviving request merges by its stable UUID after interruption.
func (v *VanillaSave) loadAndRealizeEntities(chunk *Chunk) error {
	name, root, entities, err := v.readEntityChunk(chunk.cx, chunk.cz)
	if err != nil {
		return err
	}
	blockRoot, _ := chunk.anvilRoot.Object()
	pending, _ := blockRoot["AtlasPendingStructureEntities"].Elements()
	if len(pending) == 0 {
		chunk.entities = entities
		return nil
	}
	seen := map[[16]byte]WorldEntity{}
	for _, entity := range entities {
		seen[entity.UUID] = entity
	}
	var remaining, fulfilled []nbt.Tag
	changed := false
	for _, tag := range pending {
		request, ok := tag.Object()
		if !ok {
			return fmt.Errorf("pending entity is not compound")
		}
		entity, supported, err := realizeGeneratedEntity(v.seed, v.dimension.name, request)
		if err != nil {
			return err
		}
		if !supported {
			remaining = append(remaining, tag)
			continue
		}
		cx, cz := int32(math.Floor(entity.Position[0]/16)), int32(math.Floor(entity.Position[2]/16))
		if cx != chunk.cx || cz != chunk.cz {
			return fmt.Errorf("pending entity belongs to %d,%d rather than %d,%d", cx, cz, chunk.cx, chunk.cz)
		}
		if previous, exists := seen[entity.UUID]; exists {
			if previous.Type != entity.Type || previous.Position != entity.Position {
				return fmt.Errorf("generated entity UUID collision")
			}
		} else {
			entities = append(entities, entity)
			seen[entity.UUID] = entity
			changed = true
		}
		fulfilled = append(fulfilled, tag)
	}
	if changed {
		list := make([]nbt.Tag, len(entities))
		for i, e := range entities {
			list[i] = nbt.Tag{Type: nbt.Compound, Value: e.Data}
		}
		root["Entities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: list}
		if err := v.regions.WriteChunk(v.entityDirectory(), chunk.cx, chunk.cz, name, nbt.Tag{Type: nbt.Compound, Value: root}); err != nil {
			return err
		}
	}
	chunk.entities = entities
	if len(fulfilled) > 0 {
		if len(remaining) == 0 {
			delete(blockRoot, "AtlasPendingStructureEntities")
		} else {
			blockRoot["AtlasPendingStructureEntities"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: remaining}
		}
		for _, tag := range fulfilled {
			request, _ := tag.Object()
			markOwnedEntityPieceFlag(blockRoot, request)
		}
		chunk.dirty = true
		if err := v.persistChunk(chunk); err != nil {
			return err
		}
	}
	return nil
}

func markOwnedEntityPieceFlag(root, request nbt.Object) {
	flag, _ := request.Text("PieceFlag")
	if flag != "Witch" && flag != "Cat" {
		return
	}
	structures, _ := root.Object("structures")
	starts, _ := structures.Object("starts")
	hut, _ := starts.Object("minecraft:swamp_hut")
	children, _ := hut["Children"].Elements()
	x, _ := request.Number("x")
	y, _ := request.Number("y")
	z, _ := request.Number("z")
	for _, tag := range children {
		piece, _ := tag.Object()
		id, _ := piece.Text("id")
		box, ok := piece["BB"].Value.([]int32)
		if id == "minecraft:tesh" && ok && len(box) == 6 && x >= int64(box[0]) && x <= int64(box[3]) && y >= int64(box[1]) && y <= int64(box[4]) && z >= int64(box[2]) && z <= int64(box[5]) {
			piece[flag] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
		}
	}
}
