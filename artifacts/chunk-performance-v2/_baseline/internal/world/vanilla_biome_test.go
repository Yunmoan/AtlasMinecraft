package world

import (
	"atlas/internal/anvil"
	"atlas/internal/registry"
	"atlas/internal/worldgen"
	"os"
	"path/filepath"
	"testing"
)

// A stable 1.21.1 default-overworld save from the supplied JAR is an
// independent reference for the complete climate-to-biome sampling path.
func TestOverworldBiomeSamplesAgainstVanillaChunk(t *testing.T) {
	dir := os.Getenv("ATLAS_VANILLA_BIOME_REFERENCE_WORLD")
	if dir == "" {
		t.Skip("no independent vanilla noise-world reference")
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	if v.flat != nil {
		t.Fatal("biome reference is a flat world")
	}
	name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), 0, 0)
	if err != nil || !exists {
		t.Fatalf("reference chunk 0,0: exists=%v error=%v", exists, err)
	}
	chunk, err := chunkFromAnvil(name, root, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	sampler, err := worldgen.NewOverworldBiomeSampler(v.seed)
	if err != nil {
		t.Fatal(err)
	}
	for _, qy := range []int32{-15, 0, 16, 40} {
		section := int((qy*4 - MinY) / 16)
		localY := int((qy*4 - MinY) % 16 / 4)
		for qz := int32(0); qz < 4; qz++ {
			for qx := int32(0); qx < 4; qx++ {
				got, err := sampler.BiomeAtQuart(qx, qy, qz)
				if err != nil {
					t.Fatal(err)
				}
				index := localY*16 + int(qz)*4 + int(qx)
				want, err := registry.BiomeName(chunk.biomes[section][index])
				if err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("quart %d,%d,%d: Go=%s vanilla=%s", qx, qy, qz, got, want)
				}
			}
		}
	}
}
