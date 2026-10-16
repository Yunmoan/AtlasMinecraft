package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
)

// ProcessOceanRuinTemplateBlocks preserves the original positional BlockRot
// draws and the capped archaeology rule's world-seed shuffle. Clipping happens
// after the full processor chain, as in StructureTemplate.placeInWorld.
func ProcessOceanRuinTemplateBlocks(template *StructureTemplate, origin FeaturePosition, rotation TemplateRotation, integrity float32, cold bool, worldSeed int64) ([]StructureTemplateBlock, error) {
	if template == nil || len(template.Palettes) == 0 || rotation > TemplateRotationCounterclockwise90 || integrity < 0 || integrity > 1 {
		return nil, fmt.Errorf("invalid ocean ruin processor settings")
	}
	paletteRandom := NewLegacyRandom(positionSeed(origin.X, origin.Y, origin.Z))
	palette := template.Palettes[paletteRandom.NextIntBound(int32(len(template.Palettes)))]
	blocks := make([]StructureTemplateBlock, 0, len(palette))
	for _, block := range palette {
		p, _ := TransformTemplatePosition(block.Position, FeaturePosition{}, TemplateMirrorNone, rotation)
		p = treeOffset(origin, p.X, p.Y, p.Z)
		if NewLegacyRandom(positionSeed(p.X, p.Y, p.Z)).NextFloat() > integrity {
			continue
		}
		if surfaceAir(block.State) || block.State.Name == "minecraft:structure_block" {
			continue
		}
		block.Position = p
		block.State = cloneSurfaceState(block.State)
		if block.NBT != nil {
			block.NBT = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: block.NBT}).Value.(nbt.Object)
		}
		blocks = append(blocks, block)
	}
	random := NewLegacyRandom(worldSeed).ForkPositional().AtPosition(origin)
	indices := make([]int, len(blocks))
	for index := range indices {
		indices[index] = index
	}
	for remaining := len(indices); remaining > 1; remaining-- {
		other := int(random.NextIntBound(int32(remaining)))
		indices[remaining-1], indices[other] = indices[other], indices[remaining-1]
	}
	from, to, loot := "minecraft:sand", "minecraft:suspicious_sand", "minecraft:archaeology/ocean_ruin_warm"
	if cold {
		from, to, loot = "minecraft:gravel", "minecraft:suspicious_gravel", "minecraft:archaeology/ocean_ruin_cold"
	}
	changed := 0
	for _, index := range indices {
		if changed >= 5 {
			break
		}
		block := &blocks[index]
		if block.State.Name != from {
			continue
		}
		block.State = SurfaceBlockState{Name: to}
		if block.NBT == nil {
			block.NBT = make(nbt.Object)
		}
		block.NBT["LootTable"] = nbt.Tag{Type: nbt.String, Value: loot}
		block.NBT["LootTableSeed"] = nbt.Tag{Type: nbt.Long, Value: int64(NewLegacyRandom(positionSeed(block.Position.X, block.Position.Y, block.Position.Z)).NextLong())}
		changed++
	}
	return blocks, nil
}
