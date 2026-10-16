package light

import (
	"atlas/internal/blockstate"
	"testing"
)

func BenchmarkLightEdit(b *testing.B) {
	state := func(name string) uint16 {
		id, err := blockstate.BlockStateID(name, nil)
		if err != nil {
			b.Fatal(err)
		}
		return uint16(id)
	}
	stone, dirt, glass, glow := state("stone"), state("dirt"), state("glass"), state("glowstone")
	for _, test := range []struct {
		name        string
		y           int
		first, next uint16
	}{
		{"OpaqueEquivalent", 127, stone, dirt},
		{"TransparentEquivalent", 128, 0, glass},
		{"EmitterChange", 128, 0, glow},
	} {
		b.Run(test.name, func(b *testing.B) {
			e, err := New(0, 384, true)
			if err != nil {
				b.Fatal(err)
			}
			area := make(map[Position][]uint16)
			for x := int32(-1); x <= 1; x++ {
				for z := int32(-1); z <= 1; z++ {
					values := make([]uint16, 384*256)
					for i := 0; i < 128*256; i++ {
						values[i] = stone
					}
					area[Position{x, z}] = values
				}
			}
			if _, err := e.InitializeArea(Position{}, area); err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				id := test.next
				if n&1 != 0 {
					id = test.first
				}
				if _, err := e.Edit(Position{}, []Change{{X: 8, Y: test.y, Z: 8, State: id}}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
