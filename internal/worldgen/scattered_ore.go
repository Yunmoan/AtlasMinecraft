package worldgen

import "math"

func compileScatteredOrePlan(name string, def ConfiguredFeatureDefinition) (configuredFeaturePlacer, error) {
	c, err := decodeOreConfiguration(name, def)
	if err != nil {
		return nil, err
	}
	return func(region *OverworldFeatureRegion, r *WorldgenRandom, origin FeaturePosition) (bool, error) {
		attempts := r.NextIntBound(c.Size + 1)
		for i := int32(0); i < attempts; i++ {
			distance := float32(min(i, int32(7)))
			offset := func() int32 { return int32(math.Floor(float64((r.NextFloat()-r.NextFloat())*distance) + 0.5)) }
			p := treeOffset(origin, offset(), offset(), offset())
			s, err := region.BlockAt(p)
			if err != nil {
				return false, err
			}
			// Each matching target independently checks air exposure; rejecting one
			// target does not skip later targets with the same matching block.
			for _, target := range c.Targets {
				one := c
				one.Targets = []OreTarget{target}
				replacement, allowed, err := oreMayReplace(region, r, p, s, one)
				if err != nil {
					return false, err
				}
				if allowed {
					if err := setFeatureBlock(region, p, replacement); err != nil {
						return false, err
					}
					break
				}
			}
		}
		return true, nil
	}, nil
}
