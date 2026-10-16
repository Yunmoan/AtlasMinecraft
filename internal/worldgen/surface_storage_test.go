package worldgen

import (
	"atlas/internal/blockstate"
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestSurfaceStoragePreservesAuthoredBaseStateOverride(t *testing.T) {
	chunk := &NoiseChunkResult{MinY: -64, Height: 384, Blocks: make([]NoiseMaterial, 384*256)}
	at, _ := chunk.Index(3, 70, 4)
	chunk.Blocks[at] = NoiseStone
	chunk.setSurfaceState(3, 70, 4, SurfaceBlockState{Name: "minecraft:gold_block"})
	chunk.setSurfaceState(3, 70, 4, SurfaceBlockState{Name: "minecraft:stone"})
	if _, remains := chunk.SurfaceBlocks[at]; !remains {
		t.Fatal("reverting to the base material dropped its authored override")
	}
	got, err := chunk.BlockStateAt(3, 70, 4)
	if err != nil || !reflect.DeepEqual(got, SurfaceBlockState{Name: "minecraft:stone"}) {
		t.Fatal("reverting to the base material changed its state", err)
	}
	// Authored writes must retain their map entries even when their current
	// value equals the noise material, matching sparse borrowed overlays.
	chunk.Blocks[at] = NoiseAir
	got, err = chunk.BlockStateAt(3, 70, 4)
	if err != nil || !reflect.DeepEqual(got, SurfaceBlockState{Name: "minecraft:stone"}) {
		t.Fatal("authored override followed a later base-material change", err)
	}
	// Preserve explicit property maps, including an explicitly empty map;
	// getters must retain the exact state rather than broaden equivalence.
	state := SurfaceBlockState{Name: "minecraft:stone", Properties: map[string]string{}}
	chunk.setSurfaceState(3, 70, 4, state)
	if _, exists := chunk.SurfaceBlocks[at]; !exists {
		t.Fatal("explicit properties were dropped")
	}
}

func TestSurfacePreallocationExactChunkStatesAndEncoding(t *testing.T) {
	g, err := NewOverworldNoiseGenerator(1234)
	if err != nil {
		t.Fatal(err)
	}
	biomes, err := newOverworldBiomeSampler(1234, g.settings)
	if err != nil {
		t.Fatal(err)
	}
	for _, position := range [][2]int32{{-249, 327}, {-250, 328}} {
		raw, err := g.GenerateNoise(position[0], position[1])
		if err != nil {
			t.Fatal(err)
		}
		optimized, reference := cloneTerrain(raw), cloneTerrain(raw)
		optimized.SurfaceBlocks = nil
		// An existing empty map exercises the old zero-capacity growth path.
		reference.SurfaceBlocks = make(map[int]SurfaceBlockState)
		for _, chunk := range []*NoiseChunkResult{optimized, reference} {
			if err := g.ApplySurface(chunk, biomes.BiomeAtQuart); err != nil {
				t.Fatal(err)
			}
			if err := g.ApplyCarvers(chunk, biomes.BiomeAtQuart); err != nil {
				t.Fatal(err)
			}
		}
		if !reflect.DeepEqual(optimized, reference) {
			t.Fatal("preallocation changed complete carved chunk data", position)
		}
		encoded := [2][]byte{make([]byte, len(raw.Blocks)*2), make([]byte, len(raw.Blocks)*2)}
		for i, chunk := range []*NoiseChunkResult{optimized, reference} {
			for index := range chunk.Blocks {
				state, err := chunk.BlockStateAt(int32(index&15), int32(index/256)+chunk.MinY, int32(index/16&15))
				if err != nil {
					t.Fatal(err)
				}
				id, err := blockstate.BlockStateID(state.Name, state.Properties)
				if err != nil {
					t.Fatal(err)
				}
				binary.BigEndian.PutUint16(encoded[i][index*2:], uint16(id))
			}
		}
		if !bytes.Equal(encoded[0], encoded[1]) {
			t.Fatal("preallocation changed encoded original state IDs")
		}
	}
	t.Log("two complete cold carved chunks retain exact blocks, fluid/update flags, property maps and encoded original state IDs across old/new map capacity paths")
}
