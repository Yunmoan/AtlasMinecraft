package item

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type vanillaWireSample struct{ Name, Hex string }

func vanillaWireSamples(t testing.TB) []vanillaWireSample {
	t.Helper()
	data, err := os.ReadFile("testdata/vanilla_stacks_767.json")
	if err != nil {
		t.Fatal(err)
	}
	var samples []vanillaWireSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	return samples
}

func TestOfficialVanillaWireStacks(t *testing.T) {
	for _, sample := range vanillaWireSamples(t) {
		t.Run(sample.Name, func(t *testing.T) {
			data, err := hex.DecodeString(sample.Hex)
			if err != nil {
				t.Fatal(err)
			}
			// An adjacent empty slot must remain untouched by component framing.
			r := proto.NewReader(append(bytes.Clone(data), 0))
			stack, err := Read(r)
			if err != nil || r.Remaining() != 1 {
				t.Fatalf("official stack framing: remaining=%d err=%v", r.Remaining(), err)
			}
			if !bytes.Equal(stack.Bytes(), data) {
				t.Fatal("official wire changed")
			}
			saved, err := stack.ToNBT()
			if err != nil {
				t.Fatal(err)
			}
			restored, err := FromNBT(saved)
			if err != nil || !bytes.Equal(restored.Bytes(), data) {
				t.Fatalf("official stack persistence: %v", err)
			}
		})
	}
}

func TestVanillaConsumablesImportWithoutAtlasBytes(t *testing.T) {
	for _, sample := range vanillaWireSamples(t) {
		switch sample.Name {
		case "custom_0", "custom_1", "custom_2", "custom_3", "default_minecraft:suspicious_stew_effects":
		default:
			continue
		}
		t.Run(sample.Name, func(t *testing.T) {
			data, _ := hex.DecodeString(sample.Hex)
			stack, err := Read(proto.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			tag, err := stack.ToNBT()
			if err != nil {
				t.Fatal(err)
			}
			o, _ := tag.Object()
			if components, _ := o.Object("components"); len(components) == 0 {
				t.Fatal("missing standard consumable components")
			}
			delete(o, "AtlasNetworkItem")
			delete(o, "AtlasItemFingerprint")
			imported, err := FromNBT(tag)
			if err != nil {
				t.Fatal(err)
			}
			want, _, _ := stack.components()
			got, _, _ := imported.components()
			// Imported nested items additionally retain their original NBT.
			if sample.Name == "custom_3" {
				s := got["food"].(map[string]any)["usingConvertsTo"].(Stack)
				s.Saved = nil
				got["food"].(map[string]any)["usingConvertsTo"] = s
			}
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("changed consumable values:\nwant=%#v\ngot=%#v", want, got)
			}
		})
	}
	// The potion codec also accepts a bare registry name in a Java save.
	id, _ := ID("potion")
	stack, err := FromNBT(compound(nbt.Object{"id": text("minecraft:potion"), "count": integer(1), "components": compound(nbt.Object{"minecraft:potion_contents": text("minecraft:healing")})}))
	if err != nil || stack.ID != id {
		t.Fatal("potion shorthand import", err)
	}
}

func TestComponentStacksSaveWithoutLosingPatch(t *testing.T) {
	name, _ := nbt.MarshalNetwork(text("测试工具"))
	custom, _ := nbt.MarshalNetwork(compound(nbt.Object{"owner": text("Builder"), "value": integer(17)}))
	sword, _ := ID("diamond_sword")
	swordPatch := proto.NewWriter().VarInt(3).VarInt(1).VarInt(3).VarInt(13).VarInt(5).Raw(name).VarInt(0).Raw(custom).VarInt(15).Payload()
	s := Stack{ID: sword, Count: 1, Patch: swordPatch}
	shulker, _ := ID("blue_shulker_box")
	patch := proto.NewWriter().VarInt(1).VarInt(0).VarInt(52).VarInt(3)
	Stack{}.Write(patch)
	s.Write(patch)
	Stack{}.Write(patch)
	stacks := []Stack{s, {ID: shulker, Count: 1, Patch: patch.Payload()}}
	// The captured configuration registry supplies the wire enchantment IDs.
	if enchantment, ok := registryName("enchantment", 0); ok {
		levels := nbt.Object{enchantment: integer(2)}
		tag := compound(nbt.Object{"id": text("minecraft:diamond_sword"), "count": integer(1), "components": compound(nbt.Object{"minecraft:enchantments": compound(nbt.Object{"levels": compound(levels), "show_in_tooltip": boolean(true)})})})
		imported, err := FromNBT(tag)
		if err != nil {
			t.Fatal(err)
		}
		stacks = append(stacks, imported)
	} else {
		t.Fatal("missing vanilla enchantment registry")
	}
	for _, stack := range stacks {
		r := proto.NewReader(stack.Bytes())
		parsed, err := Read(r)
		if err != nil || r.Remaining() != 0 {
			t.Fatalf("parse component stack: %v", err)
		}
		tag, err := parsed.ToNBT()
		if err != nil {
			t.Fatal(err)
		}
		data, err := nbt.Marshal("", tag)
		if err != nil {
			t.Fatal(err)
		}
		_, tag, err = nbt.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		got, err := FromNBT(tag)
		if err != nil || !bytes.Equal(got.Bytes(), stack.Bytes()) {
			t.Fatalf("saved stack lost components: %v\n%x\n%x", err, got.Bytes(), stack.Bytes())
		}
	}
}

func TestInventoryJavaSlotMappingAndExternalEdits(t *testing.T) {
	stone, _ := ID("stone")
	shield, _ := ID("shield")
	helmet, _ := ID("diamond_helmet")
	v := Inventory{Selected: 3}
	v.Slots[9] = New(stone, 37)
	v.Slots[39] = New(stone, 12)
	v.Slots[45] = New(shield, 1)
	v.Slots[5] = New(helmet, 1)
	v.Carried = New(stone, 3)
	o := nbt.Object{"EnderItems": nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{}}, "Health": {Type: nbt.Float, Value: float32(20)}}
	if err := v.SaveNBT(o); err != nil {
		t.Fatal(err)
	}
	data, err := nbt.Marshal("", compound(o))
	if err != nil {
		t.Fatal(err)
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := root.Object()
	got, err := InventoryFromNBT(parsed)
	if err != nil {
		t.Fatal(err)
	}
	for i := range v.Slots {
		if !bytes.Equal(v.Slots[i].Bytes(), got.Slots[i].Bytes()) {
			t.Fatalf("lost menu slot %d", i)
		}
	}
	if got.Selected != 3 || got.Carried.Count != 3 {
		t.Fatal("selected/cursor item lost")
	}
	entries, _ := parsed["Inventory"].Elements()
	for _, tag := range entries {
		entry, _ := tag.Object()
		slot, _ := entry.Number("Slot")
		if slot == 9 {
			entry["count"] = integer(4)
		}
	}
	got, err = InventoryFromNBT(parsed)
	if err != nil || got.Slots[9].Count != 4 {
		t.Fatalf("stale Atlas patch overrode an external vanilla edit: %v", err)
	}
}

func TestItemDecoderRejectsMalformedData(t *testing.T) {
	stone, _ := ID("stone")
	for _, data := range [][]byte{
		proto.NewWriter().VarInt(-1).Payload(),
		proto.NewWriter().VarInt(1).VarInt(9999).VarInt(0).VarInt(0).Payload(),
		proto.NewWriter().VarInt(1).VarInt(stone).VarInt(58).VarInt(0).Payload(),
		proto.NewWriter().VarInt(1).VarInt(stone).VarInt(1).VarInt(0).VarInt(5).U8(nbt.Compound).Payload(),
		proto.NewWriter().VarInt(1).VarInt(stone).VarInt(1).VarInt(0).VarInt(7).VarInt(1000000).Payload(),
		proto.NewWriter().VarInt(1).VarInt(stone).VarInt(1).VarInt(1).VarInt(3).VarInt(1).VarInt(3).Payload(),
	} {
		if _, err := Read(proto.NewReader(data)); err == nil {
			t.Fatalf("accepted malformed stack %x", data)
		}
	}
}

func TestStackIdentityIgnoresComponentOrderAndSavedSource(t *testing.T) {
	id, _ := ID("stone")
	name, _ := nbt.MarshalNetwork(text("same item"))
	a := Stack{ID: id, Count: 2, Patch: proto.NewWriter().VarInt(2).VarInt(2).VarInt(13).VarInt(100).VarInt(5).Raw(name).VarInt(14).VarInt(15).Payload()}
	b := Stack{ID: id, Count: 7, Patch: proto.NewWriter().VarInt(2).VarInt(2).VarInt(5).Raw(name).VarInt(13).VarInt(100).VarInt(15).VarInt(14).Payload()}
	if !a.SameItem(b) {
		t.Fatal("component/removal order prevented stack merge")
	}
	shulker, _ := ID("blue_shulker_box")
	wrap := func(s Stack) Stack {
		w := proto.NewWriter().VarInt(1).VarInt(0).VarInt(52).VarInt(1)
		s.Write(w)
		return Stack{ID: shulker, Count: 1, Patch: w.Payload()}
	}
	if !wrap(a).SameItem(wrap(b.WithCount(a.Count))) {
		t.Fatal("nested component order prevented equality")
	}
	if wrap(a).SameItem(wrap(b)) {
		t.Fatal("different contained quantities compared equal")
	}
	b.Patch = proto.NewWriter().VarInt(2).VarInt(2).VarInt(5).Raw(name).VarInt(13).VarInt(101).VarInt(15).VarInt(14).Payload()
	if a.SameItem(b) {
		t.Fatal("different component values compared equal")
	}
	// Original save bytes of a nested item do not affect its component identity.
	if !componentValueEqual(a, Stack{ID: a.ID, Count: a.Count, Patch: a.Patch, Saved: []byte{1, 2, 3}}) {
		t.Fatal("original save bytes changed identity")
	}
}

func TestVanillaColorAndMapEnums(t *testing.T) {
	for _, entry := range []struct {
		name, value string
		wire        int32
	}{{"base_color", "blue", 11}, {"map_post_processing", "lock", 0}, {"map_post_processing", "scale", 1}} {
		encoded, err := componentBytes(entry.name, text(entry.value))
		if err != nil {
			t.Fatal(err)
		}
		got, err := proto.NewReader(encoded).VarInt()
		if err != nil || got != entry.wire {
			t.Fatal("wrong enum wire value")
		}
		tag, ok := componentNBT(entry.name, entry.wire)
		value, _ := tag.Text()
		if !ok || value != entry.value {
			t.Fatal("wrong vanilla enum NBT")
		}
	}
}

func TestChatComponentSaveUsesJSON(t *testing.T) {
	tag, ok := componentNBT("custom_name", text("Atlas inventory probe"))
	if !ok {
		t.Fatal("export name")
	}
	jsonText, _ := tag.Text()
	if jsonText != `"Atlas inventory probe"` {
		t.Fatalf("wrong saved name: %q", jsonText)
	}
	data, err := componentBytes("custom_name", tag)
	if err != nil {
		t.Fatal(err)
	}
	decoded, _, err := nbt.ParseNetwork(data)
	name, _ := decoded.Text()
	if err != nil || name != "Atlas inventory probe" {
		t.Fatal("wrong restored name", err)
	}
}

func FuzzReadStack(f *testing.F) {
	stone, _ := ID("stone")
	f.Add(New(stone, 64).Bytes())
	f.Add([]byte{0})
	f.Add([]byte{1, 1, 1, 0, 52, 0})
	for _, sample := range vanillaWireSamples(f) {
		data, err := hex.DecodeString(sample.Hex)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		r := proto.NewReader(data)
		stack, err := Read(r)
		if err != nil {
			return
		}
		decoded, err := Read(proto.NewReader(stack.Bytes()))
		if err != nil || !bytes.Equal(decoded.Bytes(), stack.Bytes()) {
			t.Fatal("decoded stack did not round-trip")
		}
	})
}
