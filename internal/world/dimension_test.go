package world

import (
	"atlas/internal/anvil"
	"atlas/internal/light"
	"atlas/internal/proto"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDimensionWorldRestartAndMetadataIsolation(t *testing.T) {
	for _, dimension := range []string{"nether", "end"} {
		t.Run(dimension, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "world")
			seed := int64(1234)
			if err := CreateFlatWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(filepath.Join(dir, "level.dat"))
			if err != nil {
				t.Fatal(err)
			}
			w, err := OpenDimension(noViewers{}, 4, dir, dimension)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Close()
			if w.MinBuildHeight() != 0 || w.MaxBuildHeight() != 256 || w.HasSkyLight() {
				t.Fatal("incorrect official dimension bounds/skylight")
			}
			if _, err := w.LoadChunk(-1, 0); err != nil {
				t.Fatal(err)
			}
			torch, err := BlockStateID("minecraft:torch", nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := w.SetBlock(-8, 199, 8, StateStone); err != nil {
				t.Fatal(err)
			}
			if err := w.SetBlock(-8, 200, 8, torch); err != nil {
				t.Fatal(err)
			}
			if err := w.SetBlock(-7, 255, 8, StateStone); err != nil {
				t.Fatal(err)
			}
			if err := w.SetBlock(-7, -1, 8, StateStone); err == nil {
				t.Fatal("block edit below dimension floor accepted")
			}
			if err := w.SetBlock(-7, 256, 8, StateStone); err == nil {
				t.Fatal("block edit above dimension ceiling accepted")
			}
			lighting := w.lighting.Snapshot(light.Position{X: -1, Z: 0})
			if lighting == nil || len(lighting.Sky) != 18 || len(lighting.Block) != 18 {
				t.Fatal("incorrect light layer count")
			}
			for _, layer := range lighting.Sky {
				for _, v := range layer {
					if v != 0 {
						t.Fatal("sky light generated in a dimension without skylight")
					}
				}
			}
			layer := lighting.Block[(200+16)>>4]
			if len(layer) != 2048 || layer[((200&15)*256+8*16+8)/2]&15 != 14 {
				t.Fatal("torch lighting incorrect in upper build sections")
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(filepath.Join(dir, "level.dat"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("dimension rewrote shared level.dat", err)
			}
			settings, _ := builtinDimension(dimension)
			name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, settings.regionSubdir, "region"), -1, 0)
			if err != nil || !exists {
				t.Fatal("dimension Anvil chunk missing", err)
			}
			saved, err := chunkFromAnvilBounds(name, root, -1, 0, 0, 256, false)
			if err != nil {
				t.Fatal(err)
			}
			if len(saved.sect) != 16 || saved.Height(9, 8) != 256 || saved.GetBlock(8, 8, 200) != torch {
				t.Fatal("dimension bounds or upper edits did not persist")
			}
			r := proto.NewReader(saved.EncodeSections())
			container := func(paletteLimit uint8) {
				bits, err := r.U8()
				if err != nil {
					t.Fatal(err)
				}
				if bits == 0 {
					if _, err := r.VarInt(); err != nil {
						t.Fatal(err)
					}
				} else if bits < paletteLimit {
					n, err := r.VarInt()
					if err != nil {
						t.Fatal(err)
					}
					for i := int32(0); i < n; i++ {
						if _, err := r.VarInt(); err != nil {
							t.Fatal(err)
						}
					}
				}
				n, err := r.VarInt()
				if err != nil {
					t.Fatal(err)
				}
				for i := int32(0); i < n; i++ {
					if _, err := r.I64(); err != nil {
						t.Fatal(err)
					}
				}
			}
			for i := 0; i < 16; i++ {
				if _, err := r.I16(); err != nil {
					t.Fatal(err)
				}
				container(9)
				container(4)
			}
			if r.Remaining() != 0 {
				t.Fatal("dimension packet contains extra sections")
			}
			reopened, err := OpenDimension(noViewers{}, 4, dir, dimension)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			if _, err := reopened.LoadChunk(-1, 0); err != nil {
				t.Fatal(err)
			}
			id, err := reopened.BlockAt(-8, 200, 8)
			if err != nil || id != torch {
				t.Fatal("restart lost dimension edit", err)
			}
			if !reflect.DeepEqual(lighting.Block, reopened.lighting.Snapshot(light.Position{X: -1, Z: 0}).Block) {
				t.Fatal("restart changed block lighting")
			}
			if _, err := os.Stat(filepath.Join(dir, "region", "r.-1.0.mca")); !os.IsNotExist(err) {
				t.Fatal("dimension wrote the Overworld region directory", err)
			}
		})
	}
}
