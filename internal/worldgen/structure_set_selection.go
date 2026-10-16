package worldgen

import "fmt"

// SelectStructureSetStart follows ChunkGenerator.createStructures: draw from
// the remaining weights, try that structure's terrain/biome checks, remove a
// rejected entry and repeat. Each candidate's own start RNG remains separate.
func SelectStructureSetStart(seed int64, chunkX, chunkZ int32, set StructureSetDefinition, generate func(string) (*StructureStart, error)) (*StructureStart, error) {
	if len(set.Structures) == 0 || generate == nil {
		return nil, fmt.Errorf("structure-set selection requires entries and a start generator")
	}
	var total int32
	for _, entry := range set.Structures {
		if entry.Weight <= 0 || entry.Weight > 1<<30 || total > (1<<31-1)-entry.Weight {
			return nil, fmt.Errorf("invalid structure-set weight for %s", entry.Structure)
		}
		total += entry.Weight
	}
	if len(set.Structures) == 1 {
		return generate(set.Structures[0].Structure)
	}
	entries := append([]StructureSetEntry(nil), set.Structures...)
	base := NewLegacyRandom(seed)
	xm, zm := base.NextLong(), base.NextLong()
	random := NewLegacyRandom(int64(uint64(int64(chunkX))*xm ^ uint64(int64(chunkZ))*zm ^ uint64(seed)))
	for len(entries) > 0 {
		value := random.NextIntBound(total)
		index := 0
		for ; index < len(entries)-1; index++ {
			value -= entries[index].Weight
			if value < 0 {
				break
			}
		}
		entry := entries[index]
		start, err := generate(entry.Structure)
		if err != nil || start != nil {
			return start, err
		}
		total -= entry.Weight
		entries = append(entries[:index], entries[index+1:]...)
	}
	return nil, nil
}
