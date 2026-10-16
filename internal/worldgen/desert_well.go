package worldgen

import "atlas/internal/nbt"

func PlaceDesertWellFeature(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
	e := featureEditor{world: region}
	p := treeOffset(origin, 0, 1, 0)
	for surfaceAir(e.block(p)) && p.Y > -62 {
		p.Y--
	}
	if e.block(p).Name != "minecraft:sand" {
		return false, e.err
	}
	for x := int32(-2); x <= 2; x++ {
		for z := int32(-2); z <= 2; z++ {
			if surfaceAir(e.block(treeOffset(p, x, -1, z))) && surfaceAir(e.block(treeOffset(p, x, -2, z))) {
				return false, e.err
			}
		}
	}
	sandstone := SurfaceBlockState{Name: "minecraft:sandstone"}
	slab := SurfaceBlockState{Name: "minecraft:sandstone_slab"}
	for y := int32(-2); y <= 0; y++ {
		for x := int32(-2); x <= 2; x++ {
			for z := int32(-2); z <= 2; z++ {
				e.set(treeOffset(p, x, y, z), sandstone)
			}
		}
	}
	e.set(p, SurfaceBlockState{Name: "minecraft:water"})
	for _, d := range []int{2, 5, 3, 4} {
		e.set(directionPosition(p, d), SurfaceBlockState{Name: "minecraft:water"})
	}
	below := treeOffset(p, 0, -1, 0)
	e.set(below, SurfaceBlockState{Name: "minecraft:sand"})
	for _, d := range []int{2, 5, 3, 4} {
		e.set(directionPosition(below, d), SurfaceBlockState{Name: "minecraft:sand"})
	}
	for x := int32(-2); x <= 2; x++ {
		for z := int32(-2); z <= 2; z++ {
			if x == -2 || x == 2 || z == -2 || z == 2 {
				e.set(treeOffset(p, x, 1, z), sandstone)
			}
		}
	}
	for _, d := range featureHorizontalDirections {
		e.set(treeOffset(p, d.X*2, 1, d.Z*2), slab)
	}
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			state := slab
			if x == 0 && z == 0 {
				state = sandstone
			}
			e.set(treeOffset(p, x, 4, z), state)
		}
	}
	for y := int32(1); y <= 3; y++ {
		for _, c := range [][2]int32{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}} {
			e.set(treeOffset(p, c[0], y, c[1]), sandstone)
		}
	}
	choices := []FeaturePosition{p, directionPosition(p, 5), directionPosition(p, 3), directionPosition(p, 4), directionPosition(p, 2)}
	for _, depth := range []int32{1, 2} {
		q := choices[r.NextIntBound(5)]
		q.Y -= depth
		e.set(q, SurfaceBlockState{Name: "minecraft:suspicious_sand"})
		if e.err != nil {
			return false, e.err
		}
		if err := region.SetBlockEntity(q, "minecraft:suspicious_sand", nbt.Object{
			"id":            {Type: nbt.String, Value: "minecraft:brushable_block"},
			"LootTable":     {Type: nbt.String, Value: "minecraft:archaeology/desert_well"},
			"LootTableSeed": {Type: nbt.Long, Value: featurePackedPosition(q)},
		}); err != nil {
			return false, err
		}
	}
	return true, e.err
}
