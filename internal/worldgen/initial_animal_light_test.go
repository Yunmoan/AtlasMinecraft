package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/light"
	"testing"
)

func TestInitialAnimalDirectSkyMatchesFullLighting(t *testing.T) {
	var chunks []*NoiseChunkResult
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			ch := &NoiseChunkResult{X: x, Z: z, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: map[int]SurfaceBlockState{}}
			for y := int32(-64); y < 64; y++ {
				for at := int32(0); at < 256; at++ {
					ch.Blocks[(y+64)*256+at] = NoiseStone
				}
			}
			chunks = append(chunks, ch)
		}
	}
	region, err := NewOverworldFeatureRegion(chunks...)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		p     FeaturePosition
		state SurfaceBlockState
	}{
		{FeaturePosition{1, 80, 1}, SurfaceBlockState{Name: "minecraft:stone"}},
		{FeaturePosition{2, 80, 2}, SurfaceBlockState{Name: "minecraft:oak_leaves", Properties: map[string]string{"distance": "1", "persistent": "false", "waterlogged": "false"}}},
		{FeaturePosition{3, 80, 3}, SurfaceBlockState{Name: "minecraft:stone_slab", Properties: map[string]string{"type": "bottom", "waterlogged": "false"}}},
		{FeaturePosition{4, 80, 4}, SurfaceBlockState{Name: "minecraft:glass"}},
		{FeaturePosition{5, 80, 5}, SurfaceBlockState{Name: "minecraft:water", Properties: map[string]string{"level": "0"}}},
		{FeaturePosition{6, 64, 6}, SurfaceBlockState{Name: "minecraft:torch"}},
	} {
		if err := region.SetBlock(fixture.p, fixture.state); err != nil {
			t.Fatal(err)
		}
	}
	states := map[light.Position][]uint16{}
	for _, chunk := range chunks {
		v := make([]uint16, len(chunk.Blocks))
		for y := int32(-64); y < 320; y++ {
			for z := int32(0); z < 16; z++ {
				for x := int32(0); x < 16; x++ {
					state, err := region.BlockAt(FeaturePosition{chunk.X*16 + x, y, chunk.Z*16 + z})
					if err != nil {
						t.Fatal(err)
					}
					id, err := blockstate.BlockStateID(state.Name, state.Properties)
					if err != nil {
						t.Fatal(err)
					}
					v[(y+64)*256+z*16+x] = uint16(id)
				}
			}
		}
		states[light.Position{X: chunk.X, Z: chunk.Z}] = v
	}
	engine, err := light.New(-64, 384, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.InitializeArea(light.Position{}, states); err != nil {
		t.Fatal(err)
	}
	data := engine.Snapshot(light.Position{})
	for _, point := range []FeaturePosition{{0, 64, 0}, {1, 64, 1}, {2, 79, 2}, {3, 79, 3}, {4, 79, 4}, {5, 79, 5}, {6, 64, 6}, {1, 81, 1}, {1, 63, 1}} {
		env := initialAnimalRegion{region: region}
		got, err := env.rawBrightness(point)
		if err != nil {
			t.Fatal(err)
		}
		section := int((point.Y+64)>>4) + 1
		at := int((point.Y&15)*256 + (point.Z&15)*16 + (point.X & 15))
		read := func(layers [][]byte) int32 {
			if layers[section] == nil {
				return 0
			}
			return int32(layers[section][at/2] >> uint((at&1)*4) & 15)
		}
		want := max(read(data.Sky), read(data.Block))
		if got != want {
			t.Fatalf("%v got%d want%d", point, got, want)
		}
		if point == (FeaturePosition{0, 64, 0}) || point == (FeaturePosition{4, 79, 4}) {
			if env.lights != nil {
				t.Fatalf("unobstructed %v unnecessarily solved 3x3 lighting", point)
			}
		}
	}
}

func TestInitialAnimalDirectSkyRequiresCompleteNeighborhood(t *testing.T) {
	chunk := &NoiseChunkResult{MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)}
	region, err := NewOverworldFeatureRegion(chunk)
	if err != nil {
		t.Fatal(err)
	}
	env := initialAnimalRegion{region: region}
	if value, err := env.rawBrightness(FeaturePosition{0, 64, 0}); err == nil {
		t.Fatalf("incomplete neighborhood returned %d", value)
	}
}
