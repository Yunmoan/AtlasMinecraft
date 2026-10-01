package worldgen

// javaPositionSet follows HashSet<BlockPos>'s bucket traversal, insertion and
// resize order. TreeFeature can queue a leaf at multiple distances before it is
// visited, and later entries overwrite its distance; using Go map iteration or
// a shortest-path BFS changes vanilla's final block states.
type javaPositionSet struct {
	buckets [][]FeaturePosition
	size    int
}

func positionHash(p FeaturePosition) uint32 {
	h := uint32((p.Y+p.Z*31)*31 + p.X)
	return h ^ (h >> 16)
}

func (set *javaPositionSet) resize(size int) {
	old := set.buckets
	set.buckets = make([][]FeaturePosition, size)
	for _, bucket := range old {
		for _, p := range bucket {
			index := int(positionHash(p)) & (size - 1)
			set.buckets[index] = append(set.buckets[index], p)
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
	set.buckets[index] = append(set.buckets[index], p)
	set.size++
	if (len(set.buckets[index]) >= 9 && len(set.buckets) < 64) || set.size > len(set.buckets)*3/4 {
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
		set.size--
		return position
	}
	panic("worldgen: empty tree position set")
}
