package worldgen

import (
	"atlas/internal/nbt"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestMonumentStartsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/structure_monument_starts_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	var currentSeed int64
	var noise *OverworldNoiseGenerator
	var biome *OverworldBiomeSampler
	var heights map[[2]int32]int32
	count, accepted, naturalAccepted := 0, 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		if len(f) != 9 || f[0] != "monument_start" {
			t.Fatalf("invalid start vector %q", line)
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		seed, cx, cz := parse(1), int32(parse(2)), int32(parse(3))
		if noise == nil || seed != currentSeed {
			currentSeed = seed
			noise, err = NewOverworldNoiseGenerator(seed)
			if err != nil {
				t.Fatal(err)
			}
			biome, err = NewOverworldBiomeSampler(seed)
			if err != nil {
				t.Fatal(err)
			}
			heights = make(map[[2]int32]int32)
		}
		position := [2]int32{cx*16 + 8, cz*16 + 8}
		height, found := heights[position]
		if !found {
			column, err := noise.BaseStructureColumnAt(position[0], position[1])
			if err != nil {
				t.Fatal(err)
			}
			height = column.OccupiedHeight(true)
			heights[position] = height
		}
		if height != int32(parse(5)) {
			t.Fatalf("%s: ocean-floor height %d", line, height)
		}
		at := func(x, y, z int32) (string, error) {
			if f[4] == "natural" {
				return biome.BiomeAtQuart(x, y, z)
			}
			return "minecraft:" + f[4], nil
		}
		start, err := GenerateOceanMonumentStructureStart(seed, cx, cz, OceanMonumentStartSources{63, at, func(x, z int32) (int32, error) {
			if x != position[0] || z != position[1] {
				t.Fatalf("wrong floor column %d,%d", x, z)
			}
			return height, nil
		}})
		if err != nil {
			t.Fatal(err)
		}
		pieces, hash := 0, "-"
		random := strongholdLargeFeatureRandom(seed, cx, cz)
		if start != nil {
			accepted++
			if f[4] == "natural" {
				naturalAccepted++
			}
			pieces = len(start.Pieces)
			tag, err := start.NBT()
			if err != nil {
				t.Fatal(err)
			}
			hash = stringHexHash(iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: tag}))
			direction := [...]string{"north", "east", "south", "west"}[random.NextIntBound(4)]
			if _, err := assembleMonument(random, cx*16-29, cz*16-29, direction); err != nil {
				t.Fatal(err)
			}
		}
		next := int64(random.NextLong())
		if pieces != int(parse(6)) || hash != f[7] || next != parse(8) {
			t.Fatalf("%s: got %d pieces hash %s next %d", line, pieces, hash, next)
		}
		count++
	}
	if count != 504 || accepted < 224 || naturalAccepted == 0 {
		t.Fatalf("coverage: %d starts, %d accepted, %d natural accepted", count, accepted, naturalAccepted)
	}
	t.Logf("matched %d original starts (%d accepted, %d natural), base heights, complete NBT and subsequent RNG", count, accepted, naturalAccepted)
}

func TestMonumentBiomeScanBoundsAndFailures(t *testing.T) {
	// The 1.21.1 bytecode scans inclusive quart bounds around block +9,
	// Y=sea level, then tests the stub at block +8 and occupied ocean floor.
	for _, chunk := range [][2]int32{{0, 0}, {-2, 3}} {
		for _, bad := range []FeaturePosition{{-5, 8, -5}, {9, 23, 9}, {-6, 8, -5}, {9, 24, 9}} {
			calls, heightCalls := 0, 0
			inside := bad.X >= -5 && bad.X <= 9 && bad.Y >= 8 && bad.Y <= 23 && bad.Z >= -5 && bad.Z <= 9
			at := func(x, y, z int32) (string, error) {
				if calls < 3600 {
					want := FeaturePosition{chunk[0]*4 - 5 + int32(calls/16%15), 8 + int32(calls%16), chunk[1]*4 - 5 + int32(calls/(16*15))}
					if (FeaturePosition{x, y, z}) != want {
						t.Fatalf("scan %d: %v want %v", calls, FeaturePosition{x, y, z}, want)
					}
				} else if (FeaturePosition{x, y, z}) != (FeaturePosition{chunk[0]*4 + 2, 10, chunk[1]*4 + 2}) {
					t.Fatal("wrong stub biome position")
				}
				calls++
				if x == chunk[0]*4+bad.X && y == bad.Y && z == chunk[1]*4+bad.Z {
					return "minecraft:plains", nil
				}
				return "minecraft:deep_ocean", nil
			}
			start, err := GenerateOceanMonumentStructureStart(1234, chunk[0], chunk[1], OceanMonumentStartSources{63, at, func(x, z int32) (int32, error) { heightCalls++; return 43, nil }})
			if err != nil {
				t.Fatal(err)
			}
			if inside && (start != nil || calls != 3600 || heightCalls != 0) {
				t.Fatalf("rejected scan: calls=%d heights=%d", calls, heightCalls)
			}
			if !inside && (start == nil || calls != 3601 || heightCalls != 1) {
				t.Fatalf("accepted scan: calls=%d heights=%d", calls, heightCalls)
			}
		}
	}
	sentinel := errors.New("unavailable biome")
	_, err := GenerateOceanMonumentStructureStart(0, 0, 0, OceanMonumentStartSources{63, func(int32, int32, int32) (string, error) { return "", sentinel }, func(int32, int32) (int32, error) { t.Fatal("height read after biome failure"); return 0, nil }})
	if !errors.Is(err, sentinel) {
		t.Fatalf("lost biome error: %v", err)
	}
}

func monumentLayoutDescription(t *testing.T, layout *monumentLayout) string {
	t.Helper()
	var text strings.Builder
	rooms := append([]*monumentRoom(nil), layout.rooms...)
	sort.Slice(rooms, func(i, j int) bool { return rooms[i].id < rooms[j].id })
	bit := func(value bool) int {
		if value {
			return 1
		}
		return 0
	}
	for _, room := range rooms {
		fmt.Fprintf(&text, "room %d", room.id)
		for d, other := range room.connections {
			id := -1
			if other != nil {
				id = other.id
			}
			fmt.Fprintf(&text, " %d:%d", id, bit(room.open[d]))
		}
		fmt.Fprintf(&text, " %d %d\n", bit(room.claimed), bit(room.source))
	}
	pieces := append([]monumentChild{{piece: layout.building, variant: -1}}, layout.children...)
	for _, child := range pieces {
		id := -1
		if child.room != nil {
			id = child.room.id
		}
		tag, err := child.piece.NBT()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&text, "piece %d %d %s\n", id, child.variant, iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: tag}))
	}
	return text.String()
}

func TestMonumentLayoutsAgainstVanillaJAR(t *testing.T) {
	data, err := os.ReadFile("../../testdata/worldgen/structure_monument_layout_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		f := strings.Fields(line)
		if len(f) != 8 || f[0] != "monument_layout" {
			t.Fatalf("invalid monument vector %q", line)
		}
		parse := func(i int) int64 {
			v, err := strconv.ParseInt(f[i], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return v
		}
		random := NewLegacyRandom(parse(1))
		layout, err := assembleMonument(random, int32(parse(2)), int32(parse(3)), f[4])
		if err != nil {
			t.Fatal(err)
		}
		description := monumentLayoutDescription(t, layout)
		hash, next := stringHexHash(description), int64(random.NextLong())
		if len(layout.children) != int(parse(5)) || hash != f[6] || next != parse(7) {
			t.Fatalf("%s: got %d children hash %s next %d\n%s", line, len(layout.children), hash, next, description)
		}
		count++
	}
	if count != 112 {
		t.Fatalf("got %d original layouts, want 112", count)
	}
	t.Logf("matched %d full room graphs, fitted pieces, variants, saved NBT and subsequent RNG", count)
}
