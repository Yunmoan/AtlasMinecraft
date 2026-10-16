package world

// Vanilla 1.21.1 spawn column scan (PlayerRespawnLogic.getSpawnPosInChunk and
// getOverworldRespawnPos), transcribed from the reference JAR. A column is
// safe when its surface is not fluid and, scanning down from above the top
// motion-blocking block, contains a block whose collision fills its top face
// without fluid above it. Atlas additionally checks the standing player's
// clearance and avoids damaging blocks because players join directly at the
// stored position rather than running vanilla's later player placement logic.

import "atlas/internal/blockrules"

// findSafeSpawnColumn scans the chunk's columns in vanilla order (x outer,
// z inner) and returns the first safe spawn position in world coordinates.
func findSafeSpawnColumn(ch *Chunk) (int32, int32, int32, bool) {
	for lx := 0; lx < 16; lx++ {
		for lz := 0; lz < 16; lz++ {
			if y, ok := safeSpawnY(ch, lx, lz); ok {
				return int32(ch.cx*16) + int32(lx), y, int32(ch.cz*16) + int32(lz), true
			}
		}
	}
	return 0, 0, 0, false
}

// safeSpawnY reports the spawn feet Y for one chunk-local column, mirroring
// vanilla's per-column checks: reject fluid surfaces, then scan down for the
// first top-face-full block.
func safeSpawnY(ch *Chunk, lx, lz int) (int32, bool) {
	// Top-down classification with "not found" sentinels just below ch.minY,
	// mirroring vanilla heightmap top-block values.
	motion, surface, ocean := int32(ch.minY-1), int32(ch.minY-1), int32(ch.minY-1)
	for y := int32(ch.maxY - 1); y >= ch.minY; y-- {
		state := ch.GetBlock(lx, lz, y)
		if motion == ch.minY-1 && stateFlag(motionFlags, state) {
			motion = y
		}
		if surface == ch.minY-1 && !isAir(state) {
			surface = y
		}
		if ocean == ch.minY-1 && stateFlag(oceanFloorFlags, state) {
			ocean = y
		}
		if motion != ch.minY-1 && surface != ch.minY-1 && ocean != ch.minY-1 {
			break
		}
	}
	if motion < ch.minY {
		// An empty column: vanilla rejects heights below the build floor.
		return 0, false
	}
	// ws <= mb && ws > ocean identifies fluid-surfaced columns (oceans, lava
	// lakes): the world surface is at or below the motion-blocking top while
	// still above the solid floor.
	if surface <= motion && surface > ocean {
		return 0, false
	}
	for y := motion + 1; y >= ch.minY; y-- {
		state := ch.GetBlock(lx, lz, y)
		if stateFlag(fluidFlags, state) {
			break
		}
		if !stateFlag(faceFullFlags, state) {
			continue
		}
		feetY := y + 1
		if !spawnBodyFits(ch, lx, lz, feetY) {
			return 0, false
		}
		return feetY, true
	}
	return 0, false
}

func spawnBodyFits(ch *Chunk, lx, lz int, feetY int32) bool {
	if feetY < ch.minY || feetY+1 >= ch.maxY {
		return false
	}
	// A standing player centered in this column is 0.6 blocks wide and 1.8
	// blocks tall. Include the support cell because some shapes extend above it.
	body := blockrules.Box{.2, 0, .2, .8, 1.8, .8}
	for y := feetY - 1; y <= feetY+1; y++ {
		state := ch.GetBlock(lx, lz, y)
		if stateFlag(fluidFlags, state) || spawnHazard(state) {
			return false
		}
		rules, err := blockrules.Lookup(state)
		if err != nil {
			return false
		}
		for _, shape := range rules.Boxes {
			shape[1] += float64(y - feetY)
			shape[4] += float64(y - feetY)
			if overlaps(shape, body) {
				return false
			}
		}
	}
	return true
}

func spawnHazard(state uint32) bool {
	block, err := BlockStateByID(state)
	if err != nil {
		return true
	}
	switch block.Name {
	case "minecraft:magma_block", "minecraft:cactus", "minecraft:fire", "minecraft:soul_fire",
		"minecraft:sweet_berry_bush", "minecraft:wither_rose", "minecraft:powder_snow":
		return true
	case "minecraft:campfire", "minecraft:soul_campfire":
		return block.Properties["lit"] == "true"
	}
	return false
}
