package worldgen_test

import "testing"

// Pin the reported world's coast, ocean, river and river-edge terrain against
// the original 1.21.1 JAR. See docs/worldgen/river_ocean_regression.md for exact
// seed/coordinates, provenance and the distinction from final decoration.
func TestRiverOceanNoiseAgainstVanillaJAR(t *testing.T) {
	testNoiseChunksAgainstVanillaJAR(t, "river_ocean_noise_1.21.1.txt", 4)
}

func TestRiverOceanSurfaceAgainstVanillaJAR(t *testing.T) {
	testSurfaceChunksAgainstVanillaJAR(t, "river_ocean_surface_1.21.1.txt", 4)
}

func TestRiverOceanCarversAgainstVanillaJAR(t *testing.T) {
	testCarvedChunksAgainstVanillaJAR(t, "river_ocean_carver_1.21.1.txt", 4)
}
