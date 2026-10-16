package server

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func originalMenuStacks(t *testing.T, encoded string) [65]item.Stack {
	t.Helper()
	var slots [65]item.Stack
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	tag, used, err := nbt.ParseNetwork(raw)
	if err != nil || used != len(raw) {
		t.Fatal("invalid original menu NBT", err)
	}
	entries, _ := tag.Elements()
	for _, entry := range entries {
		o, _ := entry.Object()
		slot, _ := o.Number("Slot")
		if slot < 0 || slot >= 65 {
			t.Fatal("invalid original menu slot")
		}
		slots[slot], err = item.FromNBT(entry)
		if err != nil {
			t.Fatal(err)
		}
	}
	return slots
}
func TestChestMenuActionsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/chest_menu_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	counts := map[int]int{}
	vectors := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 6 || fields[0] != "chest_menu" {
			t.Fatal("invalid menu vector")
		}
		before, want := originalMenuStacks(t, fields[4]), originalMenuStacks(t, fields[5])
		var slots [63]item.Stack
		copy(slots[:], before[:63])
		carried, offhand := before[63], before[64]
		drag := chestDrag{}
		for _, action := range strings.Split(fields[3], ",") {
			parts := strings.Split(action, ":")
			var numbers [3]int
			for i, part := range parts {
				numbers[i], err = strconv.Atoi(part)
				if err != nil {
					t.Fatal(err)
				}
			}
			applyChestAction(&slots, &carried, &offhand, &drag, numbers[0], numbers[1], numbers[2], fields[2] == "true")
			counts[numbers[2]]++
		}
		var got [65]item.Stack
		copy(got[:], slots[:])
		got[63], got[64] = carried, offhand
		for i, stack := range got {
			if stack.Count != want[i].Count || !stack.Empty() && !stack.SameItem(want[i]) {
				t.Fatalf("case%s creative%s actions%s slot%d: got %x want%x", fields[1], fields[2], fields[3], i, stack.Bytes(), want[i].Bytes())
			}
		}
		vectors++
	}
	if vectors != 1024 || len(counts) != 6 {
		t.Fatal("missing original generic-menu action coverage", vectors, counts)
	}
	t.Logf("%d actual ChestMenu.clicked vectors, pickup, shift, hotbar/offhand, creative clone, three drag kinds and two-pass pickup-all", vectors)
}
func TestChestOpenScreenAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/chest_loot_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	first := strings.Fields(strings.SplitN(string(data), "\n", 2)[0])
	if len(first) != 2 || first[0] != "open_screen" {
		t.Fatal("missing original open-screen fixture")
	}
	want, err := hex.DecodeString(first[1])
	if err != nil {
		t.Fatal(err)
	}
	if got := chestOpenPayload(37, ""); !bytes.Equal(got, want) {
		t.Fatalf("open screen %x want original %x", got, want)
	}
}
func TestChestDropRequestsPreserveItems(t *testing.T) {
	id, _ := item.ID("stone")
	var slots [63]item.Stack
	slots[0] = item.New(id, 31)
	carried := item.New(id, 12)
	offhand := item.Stack{}
	drag := chestDrag{}
	applyChestAction(&slots, &carried, &offhand, &drag, -999, 0, 0, false)
	applyChestAction(&slots, &carried, &offhand, &drag, 0, 1, 4, false)
	if slots[0].Count != 31 || carried.Count != 12 {
		t.Fatal("unimplemented drop deleted authoritative items")
	}
}

func TestChestOffhandSlotAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/chest_loot_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[0] != "offhand_slot" {
			continue
		}
		want, err := hex.DecodeString(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		stone, _ := item.ID("stone")
		got := chestOffhandPayload(17, item.New(stone, 6))
		if !bytes.Equal(got, want) {
			t.Fatalf("offhand %x want original %x", got, want)
		}
		return
	}
	t.Fatal("missing original -2 raw-inventory40 slot codec fixture")
}
