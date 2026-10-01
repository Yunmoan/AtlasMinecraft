package worldgen

// javaPositionSet follows HashSet<BlockPos>'s bucket traversal, insertion and
// resize order. TreeFeature can queue a leaf at multiple distances before it is
// visited, and later entries overwrite its distance; using Go map iteration or
// a shortest-path BFS changes vanilla's final block states.
type javaPositionSet struct {
	buckets [][]FeaturePosition
	trees   []*javaPositionTree
	size    int
}

func positionHash(p FeaturePosition) uint32 {
	h := uint32((p.Y+p.Z*31)*31 + p.X)
	return h ^ (h >> 16)
}

func (set *javaPositionSet) resize(size int) {
	old := set.buckets
	oldTrees := set.trees
	set.buckets = make([][]FeaturePosition, size)
	set.trees = make([]*javaPositionTree, size)
	for i, bucket := range old {
		for _, p := range bucket {
			index := int(positionHash(p)) & (size - 1)
			set.buckets[index] = append(set.buckets[index], p)
		}
		if i < len(oldTrees) && oldTrees[i] != nil {
			low, high := i, i+len(old)
			for _, index := range []int{low, high} {
				if len(set.buckets[index]) <= 6 {
					continue
				}
				other := low
				if index == low {
					other = high
				}
				if len(set.buckets[other]) == 0 {
					set.trees[index] = oldTrees[i]
				} else {
					set.treeify(index)
				}
			}
		}
	}
}

func (set *javaPositionSet) add(p FeaturePosition) {
	if len(set.buckets) == 0 {
		set.resize(16)
	}
	index := int(positionHash(p)) & (len(set.buckets) - 1)
	for _, existing := range set.buckets[index] {
		if existing == p {
			return
		}
	}
	if set.trees[index] != nil {
		set.addTree(index, p)
	} else {
		set.buckets[index] = append(set.buckets[index], p)
	}
	set.size++
	if len(set.buckets[index]) >= 9 && set.trees[index] == nil {
		if len(set.buckets) < 64 {
			set.resize(len(set.buckets) * 2)
		} else {
			set.treeify(index)
		}
	}
	if set.size > len(set.buckets)*3/4 {
		set.resize(len(set.buckets) * 2)
	}
}

func (set *javaPositionSet) positions() []FeaturePosition {
	result := make([]FeaturePosition, 0, set.size)
	for _, bucket := range set.buckets {
		result = append(result, bucket...)
	}
	return result
}

func (set *javaPositionSet) pop() FeaturePosition {
	for i, bucket := range set.buckets {
		if len(bucket) == 0 {
			continue
		}
		position := bucket[0]
		set.buckets[i] = bucket[1:]
		// TreeFeature's pending sets are small. For a tree bin, rebuild its
		// remaining iteration order after removal before another insertion.
		if set.trees[i] != nil {
			set.trees[i] = nil
			if len(set.buckets[i]) > 6 {
				set.treeify(i)
			}
		}
		set.size--
		return position
	}
	panic("worldgen: empty tree position set")
}
