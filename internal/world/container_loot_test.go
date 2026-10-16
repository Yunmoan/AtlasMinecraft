package world

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"encoding/base64"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestChestLootAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../docs/worldgen/chest_loot_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	tables := make(map[string]int)
	horns, books, vectors := 0, 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && (fields[0] == "open_screen" || fields[0] == "offhand_slot") {
			continue
		}
		if len(fields) != 5 || fields[0] != "chest_loot" {
			t.Fatal("invalid original loot fixture")
		}
		seed, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		luck, err := strconv.ParseFloat(fields[3], 32)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := base64.StdEncoding.DecodeString(fields[4])
		if err != nil {
			t.Fatal(err)
		}
		tag, used, err := nbt.ParseNetwork(raw)
		if err != nil || used != len(raw) {
			t.Fatal("invalid original loot NBT", err)
		}
		var want [27]item.Stack
		entries, _ := tag.Elements()
		for _, entry := range entries {
			o, _ := entry.Object()
			slot, _ := o.Number("Slot")
			if slot < 0 || slot >= 27 {
				t.Fatal("original loot slot outside container")
			}
			want[slot], err = item.FromNBT(entry)
			if err != nil {
				t.Fatal("original loot component import", fields[1], err)
			}
		}
		got, err := FillChestLoot(fields[1], seed, float32(luck))
		if err != nil {
			t.Fatal(fields[1], seed, err)
		}
		for i, stack := range got {
			if stack.Count != want[i].Count || (!stack.Empty() && !stack.SameItem(want[i])) {
				t.Fatalf("%s seed%d luck%s slot%d: got %x want %x", fields[1], seed, fields[3], i, stack.Bytes(), want[i].Bytes())
			}
			name, _ := item.Name(stack.ID)
			if name == "minecraft:goat_horn" {
				horns++
			}
			if name == "minecraft:enchanted_book" {
				books++
			}
			if stack.Empty() {
				continue
			}
			saved, err := stack.ToNBT()
			if err != nil {
				t.Fatal(err)
			}
			again, err := item.FromNBT(saved)
			if err != nil || again.Count != stack.Count || !again.SameItem(stack) {
				t.Fatal("loot component failed persisted round trip", err)
			}
		}
		tables[fields[1]]++
		vectors++
	}
	if len(tables) != 17 || vectors != 1088 || horns == 0 || books == 0 {
		t.Fatal("incomplete original loot coverage", len(tables), vectors, horns, books)
	}
	for table, count := range tables {
		if count != 64 {
			t.Fatal("missing original table seed/luck vectors", table, count)
		}
	}
	t.Logf("%d actual LootTable.fill vectors,17 tables,all27 exact slots/counts/components, %d enchanted books and %d named-instrument goat horns", vectors, books, horns)
}

func TestChestLootUnsupportedRetainsAuthority(t *testing.T) {
	if _, err := FillChestLoot("minecraft:chests/pillager_outpost", 0, 0); err == nil {
		t.Fatal("zero seed silently invented a saved world sequence")
	}
	if _, err := FillChestLoot("minecraft:not_implemented", 1234, 0); err == nil {
		t.Fatal("unknown table silently invented loot")
	}
}
