package worldgen

func PlaceVinesFeature(region *OverworldFeatureRegion, p FeaturePosition) (bool, error) {
	e := featureEditor{world: region}
	if !surfaceAir(e.block(p)) {
		return false, e.err
	}
	for d := 1; d < 6; d++ {
		if !multifaceAttach(&e, p, d) {
			continue
		}
		state := SurfaceBlockState{Name: "minecraft:vine", Properties: map[string]string{"up": "false", "north": "false", "south": "false", "west": "false", "east": "false"}}
		e.set(p, stateWithProperty(state, directionNames[d], "true"))
		return e.err == nil, e.err
	}
	return false, e.err
}
