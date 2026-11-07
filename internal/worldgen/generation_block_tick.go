package worldgen

import (
	"atlas/internal/blockstate"
	"fmt"
)

// PendingBlockTick is a ProtoChunkTicks entry. The original generator discards
// requested delays when packing; activation establishes a runtime time origin.
type PendingBlockTick struct {
	Position FeaturePosition
	Block    string
	Priority int32
}

type generationBlockTickKey struct {
	position FeaturePosition
	block    string
}
type generationBlockTickQueue struct {
	ticks []PendingBlockTick
	seen  map[generationBlockTickKey]struct{}
}

func (q *generationBlockTickQueue) add(tick PendingBlockTick) {
	key := generationBlockTickKey{tick.Position, tick.Block}
	if q.seen == nil {
		q.seen = make(map[generationBlockTickKey]struct{})
	}
	if _, ok := q.seen[key]; ok {
		return
	}
	q.seen[key] = struct{}{}
	q.ticks = append(q.ticks, tick)
}

func (region *OverworldFeatureRegion) ScheduleBlockTick(p FeaturePosition, name string, _ int32, priority int32) error {
	return region.ScheduleGenerationBlockTick(PendingBlockTick{Position: p, Block: name, Priority: priority})
}

func (region *OverworldFeatureRegion) ScheduleGenerationBlockTick(tick PendingBlockTick) error {
	// WorldGenTickAccess routes to the X/Z chunk without a write-bounds test.
	// It retains out-of-height requests, but a missing owning chunk is an error.
	if tick.Priority < -3 || tick.Priority > 3 {
		return fmt.Errorf("invalid block tick priority %d", tick.Priority)
	}
	if _, err := blockstate.BlockStateID(tick.Block, nil); err != nil {
		return err
	}
	chunk, err := region.chunkAt(tick.Position.X, tick.Position.Z)
	if err != nil {
		return err
	}
	key := featureChunkPosition{chunk.X, chunk.Z}
	if region.blockTicks == nil {
		region.blockTicks = make(map[featureChunkPosition]*generationBlockTickQueue)
	}
	q := region.blockTicks[key]
	if q == nil {
		q = &generationBlockTickQueue{}
		for _, previous := range chunk.BlockTicks {
			q.add(previous)
		}
		region.blockTicks[key] = q
	}
	q.add(tick)
	if record := region.structureRecording; record != nil {
		record.blockTicks.add(tick)
	}
	return nil
}

func (region *OverworldFeatureRegion) blockTicksInChunk(cx, cz int32) []PendingBlockTick {
	key := featureChunkPosition{cx, cz}
	if q := region.blockTicks[key]; q != nil {
		return append([]PendingBlockTick(nil), q.ticks...)
	}
	return append([]PendingBlockTick(nil), region.chunks[key].BlockTicks...)
}
