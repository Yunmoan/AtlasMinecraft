package worldgen

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

//go:embed structure_order_1.21.1.txt
var structureOrderReference string

type StructureDecorationOrder struct {
	Step, Index int32
}

var structureDecorationOrders = sync.OnceValue(func() map[string]StructureDecorationOrder {
	result := make(map[string]StructureDecorationOrder)
	for _, line := range strings.Split(structureOrderReference, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 4 || fields[0] != "order" {
			continue
		}
		step, stepErr := strconv.ParseInt(fields[2], 10, 32)
		index, indexErr := strconv.ParseInt(fields[3], 10, 32)
		if stepErr != nil || indexErr != nil {
			panic("invalid embedded original structure decoration order")
		}
		result[fields[1]] = StructureDecorationOrder{int32(step), int32(index)}
	}
	if len(result) != 34 {
		panic("embedded original structure decoration order must contain 34 structures")
	}
	return result
})

// OriginalStructureDecorationOrder returns the exact built-in registry
// encounter index within the structure's GenerationStep.Decoration. The index
// counts every structure in that step, including those the caller cannot place.
func OriginalStructureDecorationOrder(name string) (StructureDecorationOrder, error) {
	order, found := structureDecorationOrders()[densityName(name)]
	if !found {
		return StructureDecorationOrder{}, fmt.Errorf("unknown original structure %q", name)
	}
	return order, nil
}

// NewStructureRegionRandom reproduces the Overworld WorldGenRegion constructor:
// RandomState.getOrCreateRandomFactory(minecraft:worldgen_region_random)
// .at(centerChunk.getWorldPosition()). Retain this caller-owned stream across
// structures and configured features in this region's decoration order.
func NewStructureRegionRandom(seed int64, chunkX, chunkZ int32) *Xoroshiro {
	factory := NewXoroshiro(seed).ForkPositional().FromHashOf("minecraft:worldgen_region_random").ForkPositional()
	return factory.At(chunkX*16, 0, chunkZ*16)
}
