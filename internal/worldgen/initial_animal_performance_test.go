package worldgen

import "testing"

func BenchmarkInitialAnimalBrightness(b *testing.B) {
	benchmarkInitialAnimalBrightness(b, false)
}

func BenchmarkInitialAnimalShadedBrightness(b *testing.B) {
	benchmarkInitialAnimalBrightness(b, true)
}

func benchmarkInitialAnimalBrightness(b *testing.B, shaded bool) {
	var chunks []*NoiseChunkResult
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			ch := &NoiseChunkResult{X: x, Z: z, MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256), SurfaceBlocks: map[int]SurfaceBlockState{}}
			for y := int32(-64); y < 64; y++ {
				for at := int32(0); at < 256; at++ {
					i := (y+64)*256 + at
					ch.Blocks[i] = NoiseStone
					if y == 63 {
						ch.SurfaceBlocks[int(i)] = SurfaceBlockState{Name: "minecraft:grass_block", Properties: map[string]string{"snowy": "false"}}
					}
				}
			}
			chunks = append(chunks, ch)
		}
	}
	region, err := NewOverworldFeatureRegion(chunks...)
	if err != nil {
		b.Fatal(err)
	}
	point := FeaturePosition{0, 64, 0}
	want := int32(15)
	if shaded {
		if err := region.SetBlock(FeaturePosition{0, 80, 0}, SurfaceBlockState{Name: "minecraft:oak_leaves", Properties: map[string]string{"distance": "1", "persistent": "false", "waterlogged": "false"}}); err != nil {
			b.Fatal(err)
		}
		point.Y = 79
		want = 14
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := initialAnimalRegion{region: region}
		if value, err := w.rawBrightness(point); err != nil || value != want {
			b.Fatal(value, err)
		}
	}
}
