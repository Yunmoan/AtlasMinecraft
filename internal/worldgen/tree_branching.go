package worldgen

func (tree *treePlacement) forkingTrunk(origin FeaturePosition, height int32) []treeFoliageAttachment {
	dir := featureHorizontalDirections[tree.random.NextIntBound(4)]
	start := height - tree.random.NextIntBound(4) - 1
	steps := 3 - tree.random.NextIntBound(3)
	x, z := origin.X, origin.Z
	var top *int32
	for y := int32(0); y < height; y++ {
		if y >= start && steps > 0 {
			x += dir.X
			z += dir.Z
			steps--
		}
		if tree.log(FeaturePosition{x, origin.Y + y, z}, "y") {
			next := origin.Y + y + 1
			top = &next
		}
	}
	var attachments []treeFoliageAttachment
	if top != nil {
		attachments = append(attachments, treeFoliageAttachment{position: FeaturePosition{x, *top, z}, radiusOffset: 1})
	}
	x, z = origin.X, origin.Z
	second := featureHorizontalDirections[tree.random.NextIntBound(4)]
	if second != dir {
		branchStart := start - tree.random.NextIntBound(2) - 1
		branchSteps := 1 + tree.random.NextIntBound(3)
		top = nil
		for y := branchStart; y < height && branchSteps > 0; y, branchSteps = y+1, branchSteps-1 {
			if y < 1 {
				continue
			}
			x += second.X
			z += second.Z
			if tree.log(FeaturePosition{x, origin.Y + y, z}, "y") {
				next := origin.Y + y + 1
				top = &next
			}
		}
		if top != nil {
			attachments = append(attachments, treeFoliageAttachment{position: FeaturePosition{x, *top, z}})
		}
	}
	return attachments
}

func (tree *treePlacement) darkOakTrunk(origin FeaturePosition, height int32) []treeFoliageAttachment {
	below := treeOffset(origin, 0, -1, 0)
	for _, delta := range []FeaturePosition{{1, 0, 0}, {0, 0, 1}, {1, 0, 1}} {
		tree.setDirt(treeOffset(below, delta.X, 0, delta.Z))
	}
	dir := featureHorizontalDirections[tree.random.NextIntBound(4)]
	start := height - tree.random.NextIntBound(4)
	steps := 2 - tree.random.NextIntBound(3)
	x, z := origin.X, origin.Z
	top := origin.Y + height - 1
	for y := int32(0); y < height; y++ {
		if y >= start && steps > 0 {
			x += dir.X
			z += dir.Z
			steps--
		}
		p := FeaturePosition{x, origin.Y + y, z}
		s := tree.block(p)
		if !tree.air(p) && !treeTags["leaves"][s.Name] {
			continue
		}
		for _, delta := range []FeaturePosition{{0, 0, 0}, {1, 0, 0}, {0, 0, 1}, {1, 0, 1}} {
			tree.log(treeOffset(p, delta.X, 0, delta.Z), "y")
		}
	}
	attachments := []treeFoliageAttachment{{position: FeaturePosition{x, top, z}, double: true}}
	for dx := int32(-1); dx <= 2; dx++ {
		for dz := int32(-1); dz <= 2; dz++ {
			if dx >= 0 && dx <= 1 && dz >= 0 && dz <= 1 || tree.random.NextIntBound(3) > 0 {
				continue
			}
			length := tree.random.NextIntBound(3) + 2
			for y := int32(0); y < length; y++ {
				tree.log(FeaturePosition{origin.X + dx, top - y - 1, origin.Z + dz}, "y")
			}
			attachments = append(attachments, treeFoliageAttachment{position: FeaturePosition{x + dx, top, z + dz}})
		}
	}
	return attachments
}
