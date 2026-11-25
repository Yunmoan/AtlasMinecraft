package light

import (
	"atlas/internal/blockstate"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"testing"
)

func recordLightEngine(h hash.Hash, e *Engine) {
	var number [8]byte
	put := func(v uint64) { binary.LittleEndian.PutUint64(number[:], v); h.Write(number[:]) }
	put(e.revision)
	for x := int32(-2); x <= 2; x++ {
		for z := int32(-2); z <= 2; z++ {
			c := e.columns[Position{x, z}]
			if c == nil {
				put(0)
				continue
			}
			put(1)
			data := c.light
			if data == nil {
				put(0)
				continue
			}
			put(1)
			put(data.Revision)
			flags := uint64(0)
			if data.Complete {
				flags |= 1
			}
			if data.Preserved {
				flags |= 2
			}
			put(flags)
			for _, channel := range [][][]byte{data.Sky, data.Block} {
				for _, layer := range channel {
					put(uint64(len(layer)))
					h.Write(layer)
				}
			}
		}
	}
}

// The whole replay digest is captured with 6f40b0e's original dense solver.
// Include every intermediate snapshot and revision, not only the final center.
func TestLightSolverRetainsDenseReplay(t *testing.T) {
	wanted := map[string]string{
		"height=32/sky=false":  "d2d9c06bd5378d892e4e22b3be93eb0c70ff6867acb6e0cf10b4416e856d37c5",
		"height=32/sky=true":   "b54817376d917b2dd0f0aa83a6ee5e8ca681d0784aaba18cfbcbf97856403846",
		"height=384/sky=false": "5a13be24cf5ee2a1a131a846b34aa2f845c7ca2cdb0be80f5126114acff438a8",
		"height=384/sky=true":  "2f00bd03e1733a3439d338eb0bad423ca5bf99f1f95216856bdd6bc78ecb0bf0",
	}
	for _, height := range []int{32, 384} {
		for _, sky := range []bool{false, true} {
			name := fmt.Sprintf("height=%d/sky=%t", height, sky)
			t.Run(name, func(t *testing.T) {
				e, err := New(-16, height, sky)
				if err != nil {
					t.Fatal(err)
				}
				stone, slab := lightState(t, "stone", nil), lightState(t, "oak_slab", nil)
				torch, glow, water := lightState(t, "torch", nil), lightState(t, "glowstone", nil), lightState(t, "water", nil)
				h := sha256.New()
				for x := int32(-2); x <= 2; x++ {
					for z := int32(-2); z <= 2; z++ {
						states := make([]uint16, height*256)
						floor := min(height/2, 128)
						for i := 0; i < floor*256; i++ {
							states[i] = stone
						}
						for i := 0; i < 256; i++ {
							if (i+int(x*3+z*7))&7 == 0 {
								states[(floor-1)*256+i] = 0
							}
							if i%5 == 0 {
								states[(floor+3)*256+i] = slab
							}
							if i%13 == 0 {
								states[floor*256+i] = water
							}
						}
						states[(floor+1)*256+15] = torch
						states[(floor-3)*256+128] = glow
						var imported *Data
						if (x+z)%3 == 0 {
							imported = &Data{Sky: make([][]byte, height/16+2), Block: make([][]byte, height/16+2), Preserved: true, Complete: x < 0}
						}
						if _, err := e.Install(Position{x, z}, states, imported); err != nil {
							t.Fatal(err)
						}
						recordLightEngine(h, e)
					}
				}
				for _, change := range []Change{{15, int(min(height/2, 128)) - 15, 0, 0}, {0, 0, 0, glow}, {0, 0, 0, stone}} {
					if _, err := e.Edit(Position{}, []Change{change}); err != nil {
						t.Fatal(err)
					}
					recordLightEngine(h, e)
				}
				if err := e.Uninstall(Position{1, 1}, func(*Data) error { return nil }); err != nil {
					t.Fatal(err)
				}
				if _, err := e.Install(Position{1, 1}, make([]uint16, height*256), nil); err != nil {
					t.Fatal(err)
				}
				recordLightEngine(h, e)
				digest := fmt.Sprintf("%x", h.Sum(nil))
				if digest != wanted[name] {
					t.Fatalf("replay changed: got %s, want %s", digest, wanted[name])
				}
				t.Logf("%s %s", name, digest)
			})
		}
	}
}

func BenchmarkLightInstallArea(b *testing.B) {
	for _, imported := range []bool{false, true} {
		b.Run(fmt.Sprintf("imported=%t", imported), func(b *testing.B) {
			id, err := blockstate.BlockStateID("stone", nil)
			if err != nil {
				b.Fatal(err)
			}
			stone := uint16(id)
			states := make([]uint16, 384*256)
			for i := 0; i < 128*256; i++ {
				states[i] = stone
			}
			var data *Data
			if imported {
				data = &Data{Sky: make([][]byte, 26), Block: make([][]byte, 26), Preserved: true}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				e, err := New(-64, 384, true)
				if err != nil {
					b.Fatal(err)
				}
				for x := int32(-4); x <= 4; x++ {
					for z := int32(-4); z <= 4; z++ {
						if _, err := e.Install(Position{x, z}, states, data); err != nil {
							b.Fatal(err)
						}
					}
				}
			}
		})
	}
}
