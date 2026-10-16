package worldgen

// SaplingBlock inherits BushBlock's canSurvive unchanged. MangrovePropagule
// overrides it and is handled separately in featurePlantSurvives.
func isFeatureSapling(name string) bool {
	switch name {
	case "minecraft:oak_sapling", "minecraft:birch_sapling", "minecraft:spruce_sapling", "minecraft:acacia_sapling", "minecraft:jungle_sapling", "minecraft:dark_oak_sapling", "minecraft:cherry_sapling":
		return true
	default:
		return false
	}
}
