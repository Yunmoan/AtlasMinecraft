package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"sort"
)

type featureBlockEntity struct {
	block string
	data  nbt.Object
}

func cloneFeatureNBT(tag nbt.Tag) nbt.Tag {
	switch v := tag.Value.(type) {
	case nbt.Object:
		copy := make(nbt.Object, len(v))
		for k, t := range v {
			copy[k] = cloneFeatureNBT(t)
		}
		tag.Value = copy
	case []nbt.Tag:
		copy := make([]nbt.Tag, len(v))
		for i, t := range v {
			copy[i] = cloneFeatureNBT(t)
		}
		tag.Value = copy
	case []byte:
		tag.Value = append([]byte(nil), v...)
	case []int32:
		tag.Value = append([]int32(nil), v...)
	case []int64:
		tag.Value = append([]int64(nil), v...)
	}
	return tag
}

func (region *OverworldFeatureRegion) SetBlockEntity(p FeaturePosition, block string, data nbt.Object) error {
	if !region.InWorld(p) {
		return nil
	}
	s, err := region.BlockAt(p)
	if err != nil {
		return err
	}
	if s.Name != block {
		return fmt.Errorf("block entity for %s at %v has block %s", block, p, s.Name)
	}
	if region.blockEntities == nil {
		region.blockEntities = make(map[FeaturePosition]featureBlockEntity)
	}
	copy := cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: data}).Value.(nbt.Object)
	copy["x"] = nbt.Tag{Type: nbt.Int, Value: p.X}
	copy["y"] = nbt.Tag{Type: nbt.Int, Value: p.Y}
	copy["z"] = nbt.Tag{Type: nbt.Int, Value: p.Z}
	region.blockEntities[p] = featureBlockEntity{block, copy}
	if recording := region.structureRecording; recording != nil {
		recording.blockEntities[p] = featureBlockEntity{block, cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: copy}).Value.(nbt.Object)}
	}
	return nil
}

func (region *OverworldFeatureRegion) BlockEntitiesInChunk(cx, cz int32) []nbt.Object {
	var positions []FeaturePosition
	for p := range region.blockEntities {
		if p.X>>4 == cx && p.Z>>4 == cz {
			positions = append(positions, p)
		}
	}
	sort.Slice(positions, func(i, j int) bool {
		a, b := positions[i], positions[j]
		if a.X != b.X {
			return a.X < b.X
		}
		if a.Y != b.Y {
			return a.Y < b.Y
		}
		return a.Z < b.Z
	})
	result := make([]nbt.Object, 0, len(positions))
	for _, p := range positions {
		result = append(result, cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: region.blockEntities[p].data}).Value.(nbt.Object))
	}
	return result
}

func featurePackedPosition(p FeaturePosition) int64 {
	return int64((uint64(uint32(p.X))&0x3ffffff)<<38 | (uint64(uint32(p.Z))&0x3ffffff)<<12 | uint64(uint32(p.Y)&0xfff))
}
