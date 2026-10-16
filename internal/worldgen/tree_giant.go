package worldgen

import "math"

type treeFoliageAttachment struct {
	position     FeaturePosition
	radiusOffset int32
	double       bool
}

func (tree *treePlacement) setDirt(p FeaturePosition) {
	ground := tree.block(p)
	if tree.config.ForceDirt || !treeTags["dirt"][ground.Name] || ground.Name == "minecraft:grass_block" || ground.Name == "minecraft:mycelium" {
		state := tree.config.Dirt
		if tree.config.dirtProvider != nil {
			state = tree.config.dirtProvider(tree.random, p)
		}
		tree.set(p, state, tree.trunks)
		tree.trunkOrder.add(p)
	}
}

func (tree *treePlacement) giantTrunk(origin FeaturePosition, height int32) []treeFoliageAttachment {
	// The first dirt consumer is shared with the other trunk placers.
	below := treeOffset(origin, 0, -1, 0)
	for _, delta := range []FeaturePosition{{1, 0, 0}, {0, 0, 1}, {1, 0, 1}} {
		tree.setDirt(treeOffset(below, delta.X, 0, delta.Z))
	}
	for y := int32(0); y < height; y++ {
		tree.log(treeOffset(origin, 0, y, 0), "y")
		if y < height-1 {
			for _, delta := range []FeaturePosition{{1, 0, 0}, {1, 0, 1}, {0, 0, 1}} {
				tree.log(treeOffset(origin, delta.X, y, delta.Z), "y")
			}
		}
	}
	attachments := []treeFoliageAttachment{{position: treeOffset(origin, 0, height, 0), double: true}}
	if tree.config.TrunkType == "minecraft:mega_jungle_trunk_placer" {
		for y := height - 2 - tree.random.NextIntBound(4); y > height/2; y -= 2 + tree.random.NextIntBound(4) {
			angle := tree.random.NextFloat() * float32(math.Pi*2)
			var x, z int32
			for i := int32(0); i < 5; i++ {
				x = int32(1.5 + carverCos(angle)*float32(i))
				z = int32(1.5 + carverSin(angle)*float32(i))
				tree.log(treeOffset(origin, x, y-3+i/2, z), "y")
			}
			attachments = append(attachments, treeFoliageAttachment{position: treeOffset(origin, x, y, z), radiusOffset: -2})
		}
	}
	return attachments
}
