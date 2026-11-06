package worldgen

import (
	"atlas/internal/nbt"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
)

// Original load/save results for the nine distinct authored payloads, including
// the spider spawner and the efficiency axe chest. Placement/loot RNG stays Go.
//
//go:embed mansion_block_entities_1.21.1.txt
var mansionBlockEntityData string

var loadMansionBlockEntityCodecs = sync.OnceValues(func() (map[string]jigsawBlockEntityCodec, error) {
	result := make(map[string]jigsawBlockEntityCodec)
	for _, line := range strings.Split(strings.TrimSpace(mansionBlockEntityData), "\n") {
		f := strings.Fields(line)
		if len(f) != 5 || f[0] != "mansion_block_entity" || len(f[2]) != 64 || f[3] != "0" && f[3] != "1" {
			return nil, fmt.Errorf("invalid mansion block-entity codec row")
		}
		encoded, err := base64.StdEncoding.DecodeString(f[4])
		if err != nil {
			return nil, err
		}
		tag, used, err := nbt.ParseNetwork(encoded)
		if err != nil || used != len(encoded) {
			return nil, fmt.Errorf("invalid mansion block-entity NBT: %v", err)
		}
		data, ok := tag.Object()
		if !ok {
			return nil, fmt.Errorf("mansion block-entity codec requires a compound")
		}
		result[f[1]+" "+f[2]] = jigsawBlockEntityCodec{randomized: f[3] == "1", data: data}
	}
	if len(result) != 9 {
		return nil, fmt.Errorf("expected nine mansion block-entity payloads")
	}
	return result, nil
})

func mansionBlockEntity(state SurfaceBlockState, input nbt.Object, random StructurePieceRandom) (nbt.Object, error) {
	codecs, err := loadMansionBlockEntityCodecs()
	if err != nil {
		return nil, err
	}
	canonical, err := jigsawCanonicalNBT(nbt.Tag{Type: nbt.Compound, Value: input})
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s %x", state.Name, sha256.Sum256([]byte(canonical)))
	codec, ok := codecs[key]
	if !ok {
		return nil, fmt.Errorf("unsupported mansion block-entity payload for %s", state.Name)
	}
	data := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: codec.data}).Value.(nbt.Object)
	if codec.randomized {
		seed := int64(random.NextLong())
		if table, _ := data.Text("LootTable"); table != "" && seed != 0 {
			data["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: seed}
		}
	}
	return data, nil
}
