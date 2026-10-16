package worldgen

import (
	"atlas/internal/blockrules"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
)

func PlaceMonsterRoomFeature(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
	e := featureEditor{world: region}
	rx, rz := r.NextIntBound(2)+2, r.NextIntBound(2)+2
	minX, maxX, minZ, maxZ := -rx-1, rx+1, -rz-1, rz+1
	solid := func(p FeaturePosition) bool {
		value, err := featureSolid(e.block(p))
		if err != nil && e.err == nil {
			e.err = err
		}
		return value
	}
	openings := 0
	for x := minX; x <= maxX; x++ {
		for y := int32(-1); y <= 4; y++ {
			for z := minZ; z <= maxZ; z++ {
				p := treeOffset(origin, x, y, z)
				if (y == -1 || y == 4) && !solid(p) {
					return false, e.err
				}
				if (x == minX || x == maxX || z == minZ || z == maxZ) && y == 0 && surfaceAir(e.block(p)) && surfaceAir(e.block(treeOffset(p, 0, 1, 0))) {
					openings++
				}
			}
		}
	}
	if openings < 1 || openings > 5 {
		return false, e.err
	}
	set := func(p FeaturePosition, state SurfaceBlockState) {
		if !vegetationTags["features_cannot_replace"][e.block(p).Name] {
			e.set(p, state)
		}
	}
	air := SurfaceBlockState{Name: "minecraft:cave_air"}
	for x := minX; x <= maxX; x++ {
		for y := int32(3); y >= -1; y-- {
			for z := minZ; z <= maxZ; z++ {
				p := treeOffset(origin, x, y, z)
				old := e.block(p)
				if x == minX || x == maxX || y == -1 || z == minZ || z == maxZ {
					if p.Y >= -64 && !solid(treeOffset(p, 0, -1, 0)) {
						e.set(p, air)
					} else if solid(p) && old.Name != "minecraft:chest" {
						state := SurfaceBlockState{Name: "minecraft:cobblestone"}
						if y == -1 && r.NextIntBound(4) != 0 {
							state.Name = "minecraft:mossy_cobblestone"
						}
						set(p, state)
					}
				} else if old.Name != "minecraft:chest" && old.Name != "minecraft:spawner" {
					set(p, air)
				}
			}
		}
	}
	for chest := 0; chest < 2; chest++ {
		for attempt := 0; attempt < 3; attempt++ {
			p := treeOffset(origin, r.NextIntBound(rx*2+1)-rx, 0, r.NextIntBound(rz*2+1)-rz)
			if !surfaceAir(e.block(p)) {
				continue
			}
			walls := 0
			for _, d := range featureHorizontalDirections {
				if solid(treeOffset(p, d.X, 0, d.Z)) {
					walls++
				}
			}
			if walls != 1 {
				continue
			}
			state := monsterRoomChestState(&e, p)
			set(p, state)
			if e.block(p).Name == "minecraft:chest" {
				if err := region.SetBlockEntity(p, "minecraft:chest", nbt.Object{
					"id": {Type: nbt.String, Value: "minecraft:chest"}, "LootTable": {Type: nbt.String, Value: "minecraft:chests/simple_dungeon"}, "LootTableSeed": {Type: nbt.Long, Value: int64(r.NextLong())},
				}); err != nil {
					return false, err
				}
			}
			break
		}
	}
	set(origin, SurfaceBlockState{Name: "minecraft:spawner"})
	if e.block(origin).Name == "minecraft:spawner" {
		entity := []string{"minecraft:skeleton", "minecraft:zombie", "minecraft:zombie", "minecraft:spider"}[r.NextIntBound(4)]
		data := nbt.Object{"id": {Type: nbt.String, Value: "minecraft:mob_spawner"},
			"SpawnData":       {Type: nbt.Compound, Value: nbt.Object{"entity": {Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: entity}}}}},
			"SpawnPotentials": {Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{}}}
		for name, value := range map[string]int16{"Delay": 20, "MinSpawnDelay": 200, "MaxSpawnDelay": 800, "SpawnCount": 4, "MaxNearbyEntities": 6, "RequiredPlayerRange": 16, "SpawnRange": 4} {
			data[name] = nbt.Tag{Type: nbt.Short, Value: value}
		}
		if err := region.SetBlockEntity(origin, "minecraft:spawner", data); err != nil {
			return false, err
		}
	}
	return true, e.err
}

func monsterRoomChestState(e *featureEditor, p FeaturePosition) SurfaceBlockState {
	solidRender := func(q FeaturePosition) bool {
		state := e.block(q)
		id, err := blockstate.BlockStateID(state.Name, state.Properties)
		if err != nil {
			e.err = err
			return false
		}
		rules, err := blockrules.Lookup(id)
		if err != nil {
			e.err = err
		}
		return rules.SolidRender
	}
	facing, wall := 2, -1
	for _, d := range []int{2, 5, 3, 4} {
		q := directionPosition(p, d)
		if e.block(q).Name == "minecraft:chest" {
			return SurfaceBlockState{Name: "minecraft:chest"}
		}
		if solidRender(q) {
			if wall != -1 {
				wall = -1
				break
			}
			wall = d
		}
	}
	if wall != -1 {
		facing = wall ^ 1
	} else {
		if solidRender(directionPosition(p, facing)) {
			facing ^= 1
		}
		if solidRender(directionPosition(p, facing)) {
			facing = map[int]int{2: 5, 5: 3, 3: 4, 4: 2}[facing]
		}
		if solidRender(directionPosition(p, facing)) {
			facing ^= 1
		}
	}
	return SurfaceBlockState{Name: "minecraft:chest", Properties: map[string]string{"facing": directionNames[facing]}}
}
