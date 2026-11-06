package worldgen

import "fmt"

func GenerateMansionStructureStart(seed int64, cx, cz int32, biomeAt StructureBiomeAtQuart, heightAt StructureBaseOccupiedHeightAt) (*StructureStart, error) {
	base := NewLegacyRandom(seed)
	xm, zm := base.NextLong(), base.NextLong()
	random := NewLegacyRandom(int64(uint64(int64(cx))*xm ^ uint64(int64(cz))*zm ^ uint64(seed)))
	return generateMansionStructureStart(cx, cz, biomeAt, heightAt, random)
}

func generateMansionStructureStart(cx, cz int32, biomeAt StructureBiomeAtQuart, heightAt StructureBaseOccupiedHeightAt, random *LegacyRandom) (*StructureStart, error) {
	if biomeAt == nil || heightAt == nil {
		return nil, fmt.Errorf("mansion requires biomes and base-noise heights")
	}
	rotation := TemplateRotation(random.NextIntBound(4))
	x, z := cx*16+7, cz*16+7
	dx, dz := int32(5), int32(5)
	switch rotation {
	case TemplateRotationClockwise90:
		dx = -5
	case TemplateRotationClockwise180:
		dx, dz = -5, -5
	case TemplateRotationCounterclockwise90:
		dz = -5
	}
	y := int32(2147483647)
	for _, p := range []FeaturePosition{{X: x, Z: z}, {X: x, Z: z + dz}, {X: x + dx, Z: z}, {X: x + dx, Z: z + dz}} {
		height, err := heightAt(p.X, p.Z)
		if err != nil {
			return nil, err
		}
		y = min(y, height)
	}
	if y < 60 {
		return nil, nil
	}
	biome, err := biomeAt(x>>2, y>>2, z>>2)
	if err != nil {
		return nil, err
	}
	names, err := StructureBiomeNames("minecraft:mansion")
	if err != nil {
		return nil, err
	}
	valid := false
	for _, name := range names {
		valid = valid || name == biome
	}
	if !valid {
		return nil, nil
	}
	pieces, err := assembleMansion(random, FeaturePosition{x, y, z}, rotation)
	if err != nil {
		return nil, err
	}
	return &StructureStart{Structure: "minecraft:mansion", Chunk: StructureChunkPosition{cx, cz}, TerrainAdjustment: "none", Pieces: pieces}, nil
}
