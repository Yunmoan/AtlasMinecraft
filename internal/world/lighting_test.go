package world

import (
	"atlas/internal/anvil"
	"atlas/internal/light"
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"bytes"
	"context"
	"math/bits"
	"path/filepath"
	"testing"
)

func sampleLight(data *light.Data, sky bool, x, y, z int) byte {
	layers := data.Block
	if sky {
		layers = data.Sky
	}
	layer := layers[(y-MinY+16)>>4]
	if layer == nil {
		return 0
	}
	i := (y&15)*256 + z*16 + x
	return layer[i/2] >> uint((i&1)*4) & 15
}

func TestLightMasksAndNibbleArrays(t *testing.T) {
	ch := NewChunk(0, 0)
	torch, err := BlockStateID("torch", nil)
	if err != nil {
		t.Fatal(err)
	}
	ch.SetBlock(8, 8, -60, torch)
	data := detachedLight(ch)
	if v := sampleLight(data, false, 8, -60, 8); v != 14 {
		t.Fatal(v)
	}
	w := proto.NewWriter()
	writeLightData(w, data)
	r := proto.NewReader(w.Payload())
	var masks [4]uint64
	for i := range masks {
		count, err := r.VarInt()
		if err != nil || count != 1 {
			t.Fatal("invalid bitset")
		}
		mask, err := r.I64()
		if err != nil {
			t.Fatal(err)
		}
		masks[i] = uint64(mask)
	}
	all := uint64(1)<<(NumSections+2) - 1
	if masks[0]&masks[2] != 0 || masks[0]|masks[2] != all || masks[1]&masks[3] != 0 || masks[1]|masks[3] != all {
		t.Fatal("data and empty masks must partition light sections")
	}
	for kind, layers := range [][][]byte{data.Sky, data.Block} {
		count, err := r.VarInt()
		if err != nil || int(count) != bits.OnesCount64(masks[kind]) {
			t.Fatal("array count disagrees with mask")
		}
		for i, layer := range layers {
			if masks[kind]&(1<<i) == 0 {
				continue
			}
			n, err := r.VarInt()
			if err != nil || n != 2048 {
				t.Fatal("invalid light array size")
			}
			for _, want := range layer {
				got, err := r.U8()
				if err != nil || got != want {
					t.Fatal("nibble array altered")
				}
			}
		}
	}
	if r.Remaining() != 0 {
		t.Fatal("trailing packet data")
	}
}

func TestLightingRestartAndExternalBlockChange(t *testing.T) {
	w, dir := mechanicsWorld(t)
	torch, err := BlockStateID("torch", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = w.SetBlock(15, -60, 8, torch); err != nil {
		t.Fatal(err)
	}
	if got := sampleLight(w.lighting.Snapshot(light.Position{X: 1, Z: 0}), false, 0, -60, 8); got != 13 {
		t.Fatalf("border torch %d", got)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(noViewers{}, 4, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if err = reopened.PrepareSpawn(context.Background(), 0, 0, 1, nil); err != nil {
		t.Fatal(err)
	}
	if got := sampleLight(reopened.lighting.Snapshot(light.Position{X: 1, Z: 0}), false, 0, -60, 8); got != 13 {
		t.Fatalf("restart border torch %d", got)
	}
	if err = reopened.SetBlock(15, -60, 8, StateAir); err != nil {
		t.Fatal(err)
	}
	if got := sampleLight(reopened.lighting.Snapshot(light.Position{X: 1, Z: 0}), false, 0, -60, 8); got != 0 {
		t.Fatalf("removed torch leaves %d", got)
	}
	if err = reopened.Save(); err != nil {
		t.Fatal(err)
	}
	name, root, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), 1, 0)
	if err != nil || !exists {
		t.Fatal(err)
	}
	saved, err := chunkFromAnvil(name, root, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !saved.lights.Preserved {
		t.Fatal("incomplete edge lighting not retained")
	}
	saved.SetBlock(1, 8, -60, StateStone)
	_, modified, err := saved.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	object, _ := modified.Object()
	marker, _ := object.Object("AtlasLight")
	// Simulate another tool changing blocks while retaining the old marker.
	marker["BlocksSHA256"] = nbt.Tag{Type: nbt.ByteArray, Value: lightBlockHash(NewChunk(1, 0))}
	changed, err := chunkFromAnvil(name, modified, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if changed.lights != nil {
		t.Fatal("stale incomplete imported light trusted after block change")
	}
}

func TestIncompleteLightingIsPreservedOnRoundTrip(t *testing.T) {
	ch := NewChunk(-1, 0)
	ch.lights = detachedLight(ch)
	if err := ch.initializeAnvilRoot(); err != nil {
		t.Fatal(err)
	}
	name, root, err := ch.updateAnvilBlocks()
	if err != nil {
		t.Fatal(err)
	}
	reload, err := chunkFromAnvil(name, root, -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if reload.lights == nil || !reload.lights.Preserved {
		t.Fatal("missing preserved lighting")
	}
	a, b := proto.NewWriter(), proto.NewWriter()
	writeLightData(a, ch.lights)
	writeLightData(b, reload.lights)
	if !bytes.Equal(a.Payload(), b.Payload()) {
		t.Fatal("light changes on reopening")
	}
}
