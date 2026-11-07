package world

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/worldgen"
	"bufio"
	"encoding/hex"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type fluidEffectPacket struct {
	id   int32
	data []byte
}
type fluidEffectSink struct {
	mu        sync.Mutex
	packets   []fluidEffectPacket
	soundArea [4]float64
}

func (s *fluidEffectSink) SendPacket(id int32, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.packets = append(s.packets, fluidEffectPacket{id, append([]byte(nil), data...)})
}
func (*fluidEffectSink) SendChunkData([]byte)                  {}
func (s *fluidEffectSink) ViewersOf(int32, int32) []PacketSink { return []PacketSink{s} }
func (s *fluidEffectSink) SoundListeners(x, y, z, radius float64) []PacketSink {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.soundArea = [4]float64{x, y, z, radius}
	return []PacketSink{s}
}

func TestFluidContainerSoundPacketsAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/fluid_container_sound_packets_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	count := 0
	for scan.Scan() {
		fields := strings.Fields(scan.Text())
		if len(fields) != 3 {
			t.Fatal("invalid native sound fixture")
		}
		body, err := hex.DecodeString(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		next, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		sink := &fluidEffectSink{}
		w := &World{hub: sink, tickRandom: worldgen.NewLegacyRandom(1234)}
		r := fluidRuntime{world: w}
		if err := r.FluidSound(worldgen.FeaturePosition{X: -39, Y: -53, Z: 712}, fields[0]); err != nil {
			t.Fatal(err)
		}
		if len(sink.packets) != 1 || sink.packets[0].id != proto.SSoundEffect || !reflect.DeepEqual(sink.packets[0].data, body) {
			t.Fatalf("native sound packet mismatch: %s %+v want %x", fields[0], sink.packets, body)
		}
		if sink.soundArea != [4]float64{-38.5, -52.5, 712.5, 16} || w.tickRandom.NextLong() != next {
			t.Fatal("wrong sound radius or runtime RNG advancement")
		}
		count++
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatal("incomplete original sound packets")
	}
}

func TestFluidCampfireDowsePreservesFoodAndCookingData(t *testing.T) {
	dir := t.TempDir()
	seed := int64(0)
	if err := CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	v, err := openVanilla(dir)
	if err != nil {
		t.Fatal(err)
	}
	c, err := v.loadChunk(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	id, err := BlockStateID("minecraft:campfire", map[string]string{"lit": "true", "waterlogged": "false", "facing": "north", "signal_fire": "false"})
	if err != nil {
		t.Fatal(err)
	}
	c.SetBlock(8, 8, 64, id)
	food := nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: nbt.Object{"Slot": {Type: nbt.Byte, Value: int8(0)}, "id": {Type: nbt.String, Value: "minecraft:potato"}, "count": {Type: nbt.Int, Value: int32(1)}}}}}
	entity := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{"id": {Type: nbt.String, Value: "minecraft:campfire"}, "x": {Type: nbt.Int, Value: int32(8)}, "y": {Type: nbt.Int, Value: int32(64)}, "z": {Type: nbt.Int, Value: int32(8)}, "Items": food, "CookingTimes": {Type: nbt.IntArray, Value: []int32{33, 0, 0, 0}}, "CookingTotalTimes": {Type: nbt.IntArray, Value: []int32{600, 0, 0, 0}}}}
	entities := nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{entity}}
	root, _ := c.anvilRoot.Object()
	root["block_entities"] = entities
	if err := v.persistChunk(c); err != nil {
		t.Fatal(err)
	}
	sink := &fluidEffectSink{}
	w, err := Open(sink, 1, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, err := w.LoadChunk(x, z); err != nil {
				t.Fatal(err)
			}
		}
	}
	w.tickRandom = worldgen.NewLegacyRandom(1234)
	r := fluidRuntime{world: w, settings: w.fluidSettings()}
	state, _ := BlockStateByID(id)
	w.blockUpdateMu.Lock()
	err = r.PlaceFluidInContainer(worldgen.FeaturePosition{X: 8, Y: 64, Z: 8}, worldgen.SurfaceBlockState{Name: state.Name, Properties: state.Properties}, worldgen.FluidState{Type: worldgen.FluidWater, Amount: 8})
	w.blockUpdateMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	saved, err := w.store.vanilla.loadChunk(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	root, _ = saved.anvilRoot.Object()
	if !reflect.DeepEqual(root["block_entities"], entities) {
		t.Fatal("dowse changed food or cooking progress")
	}
	after, _ := BlockStateByID(saved.GetBlock(8, 8, 64))
	if after.Properties["lit"] != "false" || after.Properties["waterlogged"] != "true" {
		t.Fatal("campfire did not extinguish", after)
	}
	sink.mu.Lock()
	defer sink.mu.Unlock()
	found := false
	for _, p := range sink.packets {
		if p.id == proto.SBlockEntityData {
			reader := proto.NewReader(p.data)
			position, e := reader.I64()
			if e != nil || position != proto.PackPosition(8, 64, 8) {
				t.Fatal("wrong block entity position")
			}
			kind, e := reader.VarInt()
			if e != nil || kind != 32 {
				t.Fatal("wrong block entity type")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("campfire update was not delivered")
	}
}
