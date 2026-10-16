package worldgen

func (tree *treePlacement) cherryTrunk(origin FeaturePosition, height int32) []treeFoliageAttachment {
	config := tree.config
	first := max(int32(0), height-1+randomBetweenInclusive(tree.random, config.branchStartMin, config.branchStartMax))
	second := max(int32(0), height-1+randomBetweenInclusive(tree.random, config.branchStartMin, config.branchStartMax-1))
	if second >= first {
		second++
	}
	count := config.branchCount(tree.random)
	trunkHeight := first + 1
	if count == 3 {
		trunkHeight = height
	} else if count >= 2 {
		trunkHeight = max(first, second) + 1
	}
	for y := int32(0); y < trunkHeight; y++ {
		tree.log(treeOffset(origin, 0, y, 0), "y")
	}
	var attachments []treeFoliageAttachment
	if count == 3 {
		attachments = append(attachments, treeFoliageAttachment{position: treeOffset(origin, 0, trunkHeight, 0)})
	}
	direction := []int{2, 5, 3, 4}[tree.random.NextIntBound(4)]
	axis := "x"
	if direction == 2 || direction == 3 {
		axis = "z"
	}
	attachments = append(attachments, tree.cherryBranch(origin, height, direction, axis, first, first < trunkHeight-1))
	if count >= 2 {
		attachments = append(attachments, tree.cherryBranch(origin, height, direction^1, axis, second, second < trunkHeight-1))
	}
	return attachments
}

func (tree *treePlacement) cherryBranch(origin FeaturePosition, height int32, direction int, axis string, start int32, belowTop bool) treeFoliageAttachment {
	p := treeOffset(origin, 0, start, 0)
	endY := height - 1 + tree.config.branchEnd(tree.random)
	long := belowTop || endY < start
	length := tree.config.branchLength(tree.random)
	steps := 1
	if long {
		length++
		steps = 2
	}
	v := featureDirections[direction]
	end := treeOffset(origin, v.X*length, endY, v.Z*length)
	for i := 0; i < steps; i++ {
		p = directionPosition(p, direction)
		tree.log(p, axis)
	}
	vertical := int32(-1)
	if end.Y > p.Y {
		vertical = 1
	}
	for p != end {
		distance := abs32(end.X-p.X) + abs32(end.Y-p.Y) + abs32(end.Z-p.Z)
		up := tree.random.NextFloat() < float32(abs32(end.Y-p.Y))/float32(distance)
		if up {
			p.Y += vertical
			tree.log(p, "y")
		} else {
			p = directionPosition(p, direction)
			tree.log(p, axis)
		}
	}
	return treeFoliageAttachment{position: treeOffset(end, 0, 1, 0)}
}

func (tree *treePlacement) cherryHangingRow(center FeaturePosition, radius, y int32, double bool) {
	tree.leafRowLarge(center, radius, y, double)
	extra := int32(0)
	if double {
		extra = 1
	}
	below := treeOffset(center, 0, -1, 0)
	clockwise := [...]int{0, 1, 5, 4, 2, 3}
	for _, dir := range []int{2, 5, 3, 4} {
		perpendicular := clockwise[dir]
		distance := radius
		if perpendicular == 3 || perpendicular == 5 {
			distance += extra
		}
		v, w := featureDirections[perpendicular], featureDirections[dir]
		p := treeOffset(center, v.X*distance-w.X*radius, y-1, v.Z*distance-w.Z*radius)
		for i := -radius; i < radius+extra; i++ {
			above := treeOffset(p, 0, 1, 0)
			if tree.leaves[above] && tree.cherryHangingLeaf(below, p, tree.config.hangingChance) {
				tree.cherryHangingLeaf(below, treeOffset(p, 0, -1, 0), tree.config.hangingExtension)
			}
			p = directionPosition(p, dir)
		}
	}
}

func (tree *treePlacement) cherryHangingLeaf(center, p FeaturePosition, chance float32) bool {
	if abs32(p.X-center.X)+abs32(p.Y-center.Y)+abs32(p.Z-center.Z) >= 7 {
		return false
	}
	if tree.random.NextFloat() > chance {
		return false
	}
	return tree.leaf(p)
}
