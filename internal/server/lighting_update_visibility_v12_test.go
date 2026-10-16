package server

import (
	"atlas/internal/anvil"
	"atlas/internal/hub"
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"bytes"
	"math/bits"
	"path/filepath"
	"strconv"
	"testing"
)

type receivedLightV12 struct{ sky, block [][]byte }

func decodeLightV12(t *testing.T, r *proto.Reader) receivedLightV12 {
	t.Helper()
	var masks [4]uint64
	for i := range masks {
		n, err := r.VarInt()
		if err != nil || n != 1 {
			t.Fatal("light bitset length", n, err)
		}
		mask, err := r.I64()
		if err != nil {
			t.Fatal(err)
		}
		masks[i] = uint64(mask)
	}
	all := uint64(1)<<(world.NumSections+2) - 1
	if masks[0]&masks[2] != 0 || masks[0]|masks[2] != all || masks[1]&masks[3] != 0 || masks[1]|masks[3] != all {
		t.Fatal("lighting masks do not partition dimension sections")
	}
	data := receivedLightV12{make([][]byte, world.NumSections+2), make([][]byte, world.NumSections+2)}
	for kind, layers := range [][][]byte{data.sky, data.block} {
		n, err := r.VarInt()
		if err != nil || int(n) != bits.OnesCount64(masks[kind]) {
			t.Fatal("light array count", n, err)
		}
		for i := range layers {
			if masks[kind]&(1<<i) == 0 {
				continue
			}
			length, err := r.VarInt()
			if err != nil || length != 2048 {
				t.Fatal("light array length", length, err)
			}
			layers[i], err = r.Bytes(int(length))
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if r.Remaining() != 0 {
		t.Fatal("trailing light packet data")
	}
	return data
}

func sampleLightV12(layers [][]byte, x, y, z int) byte {
	layer := layers[(y-world.MinY+16)>>4]
	if layer == nil {
		return 0
	}
	i := (y&15)*256 + z*16 + x
	return layer[i/2] >> uint((i&1)*4) & 15
}

func equalLightV12(a, b receivedLightV12) bool {
	for _, pair := range [][2][][]byte{{a.sky, b.sky}, {a.block, b.block}} {
		for i := range pair[0] {
			if !bytes.Equal(pair[0][i], pair[1][i]) {
				return false
			}
		}
	}
	return true
}

func TestPlayerLightEditsTwoSinksPersistAndReload(t *testing.T) {
	dir, seed := t.TempDir(), int64(0)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	h := hub.New()
	w, err := world.Open(hub.DimensionView{Hub: h, Dimension: "minecraft:overworld"}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { w.Close() }()
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	s := New(Config{}, h, w)
	var viewers []*player.Player
	for id := byte(1); id <= 2; id++ {
		p := player.New("LightViewer"+strconv.Itoa(int(id)), [16]byte{id}, int32(id), nil)
		p.SetInitialDimension("minecraft:overworld")
		p.SetPos(float64(id)+.5, -60, 8.5, 0, 0, true)
		p.SetPlayReady(true)
		if !h.Join(p) {
			t.Fatal("viewer join failed")
		}
		viewers = append(viewers, p)
	}
	barrier := func() {
		t.Helper()
		if _, err := w.ReadLoadedBlocks([]world.BlockPosition{{X: -1, Y: -60, Z: 8}, {X: 0, Y: -60, Z: 8}}); err != nil {
			t.Fatal(err)
		}
	}
	deliver := func() []map[int32]receivedLightV12 {
		t.Helper()
		result := make([]map[int32]receivedLightV12, 2)
		for i, p := range viewers {
			result[i] = make(map[int32]receivedLightV12)
			for _, cx := range []int32{-1, 0} {
				if !p.ReserveChunk(cx, 0) {
					t.Fatal("chunk reservation failed")
				}
				if err := w.SpawnChunk(p.WorldPacketSink(), cx, 0, func(int32, int32) {}); err != nil {
					t.Fatal(err)
				}
			}
			barrier()
			for len(p.Out) > 0 {
				packet := <-p.Out
				if packet.ID != proto.SChunkData {
					continue
				}
				r := proto.NewReader(packet.Payload)
				cx, err := r.I32()
				if err != nil {
					t.Fatal(err)
				}
				cz, err := r.I32()
				if err != nil || cz != 0 {
					t.Fatal("chunk coordinates", cz, err)
				}
				_, size, err := nbt.ParseNetwork(r.Rest())
				if err != nil {
					t.Fatal(err)
				}
				if _, err := r.Bytes(size); err != nil {
					t.Fatal(err)
				}
				size32, err := r.VarInt()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := r.Bytes(int(size32)); err != nil {
					t.Fatal(err)
				}
				entities, err := r.VarInt()
				if err != nil || entities != 0 {
					t.Fatal("unexpected block entities", entities, err)
				}
				result[i][cx] = decodeLightV12(t, r)
			}
			if len(result[i]) != 2 {
				t.Fatal("missing actual chunk sink delivery")
			}
		}
		return result
	}
	deliver()
	glowstone, err := world.BlockStateID("glowstone", nil)
	if err != nil {
		t.Fatal(err)
	}
	viewers[0].SetHotbarSlot(36, glowstone)
	consumeEdits := func(state uint32, emitted byte) []map[int32]receivedLightV12 {
		t.Helper()
		barrier()
		result := make([]map[int32]receivedLightV12, 2)
		for i, p := range viewers {
			result[i] = make(map[int32]receivedLightV12)
			seenBlock := false
			for len(p.Out) > 0 {
				packet := <-p.Out
				r := proto.NewReader(packet.Payload)
				switch packet.ID {
				case proto.SBlockChange:
					position, err := r.I64()
					if err != nil {
						t.Fatal("wrong source block packet", position, err)
					}
					got, err := r.VarInt()
					// Placement also resends the clicked support and the space
					// above the target to reconcile client block prediction.
					if position != proto.PackPosition(-1, -60, 8) {
						if err != nil || position != proto.PackPosition(-1, -61, 8) && position != proto.PackPosition(-1, -59, 8) {
							t.Fatal("unexpected prediction correction", position, err)
						}
						continue
					}
					if err != nil || uint32(got) != state {
						t.Fatal("wrong player edit packet", got, err)
					}
					seenBlock = true
				case proto.SUpdateLight:
					if !seenBlock {
						t.Fatal("lighting reached player sink before source block modification")
					}
					cx, err := r.VarInt()
					if err != nil {
						t.Fatal(err)
					}
					cz, err := r.VarInt()
					if err != nil || cz != 0 || cx != -1 && cx != 0 {
						t.Fatal("light update for terrain not delivered to player", cx, cz, err)
					}
					result[i][cx] = decodeLightV12(t, r)
				}
			}
			if !seenBlock || len(result[i]) != 2 {
				t.Fatal("player sink missed source or neighbor edit")
			}
			for _, cx := range []int32{-1, 0} {
				want, x := emitted, 15
				if cx == 0 {
					x = 0
					if want > 0 {
						want--
					}
				}
				if got := sampleLightV12(result[i][cx].block, x, -60, 8); got != want {
					t.Fatal("actual player block-light nibble", cx, got, want)
				}
			}
		}
		for _, cx := range []int32{-1, 0} {
			if !equalLightV12(result[0][cx], result[1][cx]) {
				t.Fatal("two players received different light arrays")
			}
		}
		return result
	}
	checkSaved := func(want map[int32]receivedLightV12) {
		t.Helper()
		if err := w.Save(); err != nil {
			t.Fatal(err)
		}
		for cx, expected := range want {
			_, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), cx, 0)
			if err != nil || !exists {
				t.Fatal("missing saved edited lighting", err)
			}
			root, _ := tag.Object()
			marker, _ := root.Object("AtlasLight")
			if stale, _ := marker.Number("Stale"); stale != 0 {
				t.Fatal("saved player edit left stale lighting marker")
			}
			data := receivedLightV12{make([][]byte, world.NumSections+2), make([][]byte, world.NumSections+2)}
			sections, _ := root["sections"].Elements()
			for _, section := range sections {
				o, _ := section.Object()
				y, _ := o.Number("Y")
				index := int(y) - (world.MinY/16 - 1)
				if index >= 0 && index < len(data.sky) {
					data.sky[index], _ = o["SkyLight"].Value.([]byte)
					data.block[index], _ = o["BlockLight"].Value.([]byte)
				}
			}
			if !equalLightV12(data, expected) {
				t.Fatal("saved Anvil light differs from actual two-player update", cx)
			}
		}
	}
	reopen := func(expected map[int32]receivedLightV12) {
		t.Helper()
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		w, err = world.Open(hub.DimensionView{Hub: h, Dimension: "minecraft:overworld"}, 2, dir)
		if err != nil {
			t.Fatal(err)
		}
		s.World = w
		for _, p := range viewers {
			p.UnmarkChunkSent(-1, 0)
			p.UnmarkChunkSent(0, 0)
		}
		actual := deliver()
		for i := range viewers {
			for cx, data := range expected {
				if !equalLightV12(actual[i][cx], data) {
					t.Fatal("reopened terrain packet lost saved lighting", i, cx)
				}
			}
		}
	}
	s.handlePlace(viewers[0], placePayload(-1, -61, 8, 1, 1200, .5))
	assertInteractionState(t, w, -1, -60, 8, glowstone)
	lit := consumeEdits(glowstone, 15)
	if sky := sampleLightV12(lit[0][-1].sky, 15, -60, 8); sky != 0 {
		t.Fatal("opaque light source did not update sky channel", sky)
	}
	checkSaved(lit[0])
	reopen(lit[0])
	s.handleDig(viewers[0], proto.NewWriter().VarInt(0).I64(proto.PackPosition(-1, -60, 8)).I8(1).VarInt(1201).Payload())
	assertInteractionState(t, w, -1, -60, 8, world.StateAir)
	dark := consumeEdits(world.StateAir, 0)
	if sky := sampleLightV12(dark[0][-1].sky, 15, -60, 8); sky != 15 {
		t.Fatal("removed source retained sky occlusion", sky)
	}
	checkSaved(dark[0])
	reopen(dark[0])
}
