package worldgen

import (
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Only the original built-in payloads are pre-normalized here. These are
// original block entity load/save codec results, not block placement records;
// rotation, clipping, processors and loot seed draws still run in pure Go.
//
//go:embed jigsaw_block_entities_1.21.1.txt
var jigsawBlockEntityData string

type jigsawBlockEntityCodec struct {
	randomized bool
	data       nbt.Object
}
type jigsawBlockEntityCodecs struct {
	blocks  map[string]bool
	entries map[string]jigsawBlockEntityCodec
}

var loadJigsawBlockEntityCodecs = sync.OnceValues(func() (*jigsawBlockEntityCodecs, error) {
	result := &jigsawBlockEntityCodecs{blocks: make(map[string]bool), entries: make(map[string]jigsawBlockEntityCodec)}
	scanner := bufio.NewScanner(strings.NewReader(jigsawBlockEntityData))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	if !scanner.Scan() || scanner.Text() != "version 3955" {
		return nil, fmt.Errorf("Jigsaw block entity codecs require original version 3955")
	}
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "entity_block" {
			result.blocks[fields[1]] = true
			continue
		}
		if len(fields) != 4 || len(fields[1]) != 64 || (fields[2] != "0" && fields[2] != "1") {
			return nil, fmt.Errorf("invalid original Jigsaw block entity codec row")
		}
		encoded, err := base64.StdEncoding.DecodeString(fields[3])
		if err != nil {
			return nil, err
		}
		tag, size, err := nbt.ParseNetwork(encoded)
		if err != nil || size != len(encoded) {
			return nil, fmt.Errorf("invalid Jigsaw block entity codec NBT: %v", err)
		}
		data, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("Jigsaw block entity codec is not a compound")
		}
		result.entries[fields[0]+" "+fields[1]] = jigsawBlockEntityCodec{randomized: fields[2] == "1", data: data}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(result.entries) != 83 || len(result.blocks) < 80 {
		return nil, fmt.Errorf("incomplete original Jigsaw block entity codec inventory: %d/%d", len(result.entries), len(result.blocks))
	}
	return result, nil
})

func jigsawCanonicalNBT(tag nbt.Tag) (string, error) {
	if object, ok := tag.Object(); ok {
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var result strings.Builder
		result.WriteString("C{")
		for _, key := range keys {
			value, err := jigsawCanonicalNBT(object[key])
			if err != nil {
				return "", err
			}
			result.WriteString(base64.StdEncoding.EncodeToString([]byte(key)))
			result.WriteByte(':')
			result.WriteString(value)
			result.WriteByte(';')
		}
		result.WriteByte('}')
		return result.String(), nil
	}
	if list, ok := tag.Elements(); ok {
		var result strings.Builder
		fmt.Fprintf(&result, "L%d[", tag.ListType)
		for _, item := range list {
			value, err := jigsawCanonicalNBT(item)
			if err != nil {
				return "", err
			}
			result.WriteString(value)
			result.WriteByte(';')
		}
		result.WriteByte(']')
		return result.String(), nil
	}
	encoded, err := nbt.MarshalNetwork(tag)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encoded), nil
}

func jigsawNormalizeBlockEntity(state SurfaceBlockState, input nbt.Object, random StructurePieceRandom) (nbt.Object, error) {
	codecs, err := loadJigsawBlockEntityCodecs()
	if err != nil {
		return nil, err
	}
	if !codecs.blocks[state.Name] {
		return nil, nil
	}
	if state.Name == "minecraft:suspicious_gravel" || state.Name == "minecraft:suspicious_sand" {
		// AppendLoot produces this payload dynamically in CappedProcessor.
		// BrushableBlockEntity is not a randomizable container: retain the
		// rule's position-random loot seed without a decoration-random draw.
		data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:brushable_block"}}
		if table, ok := input["LootTable"]; ok {
			data["LootTable"] = table
			if seed, ok := input["LootTableSeed"]; ok {
				if value, _ := seed.Number(); value != 0 {
					data["LootTableSeed"] = seed
				}
			}
		}
		return data, nil
	}
	canonical, err := jigsawCanonicalNBT(nbt.Tag{Type: nbt.Compound, Value: input})
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s %x", state.Name, sha256.Sum256([]byte(canonical)))
	codec, ok := codecs.entries[key]
	if !ok {
		return nil, fmt.Errorf("Jigsaw block entity %q has an unsupported custom payload", state.Name)
	}
	data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: codec.data}).Value.(nbt.Object)
	if codec.randomized {
		seed := int64(random.NextLong())
		if table, _ := data.Text("LootTable"); table != "" {
			if seed != 0 {
				data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: seed}
			}
		}
	}
	return data, nil
}
