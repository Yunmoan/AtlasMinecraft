package world

import (
	"atlas/internal/anvil"
	"atlas/internal/nbt"
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// These coordinates come from an actual official 1.21.1 full world save,
// independently of Atlas's generator and the synthetic block-state oracle.
func TestGeneratedBubbleColumnsPersistAndPreserveEdits(t *testing.T) {
	const seed = int64(4511451270791457643)
	file, err := os.Open("../../docs/worldgen/native_bubble_columns_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	parse := func(text string) int64 {
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	counts := make(map[[2]int32]int)
	want := make(map[BlockPosition]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) < 2 || parse(f[1]) != seed {
			t.Fatalf("invalid native bubble reference: %s", scanner.Text())
		}
		switch {
		case len(f) == 5 && f[0] == "native_bubble_chunk":
			counts[[2]int32{int32(parse(f[2])), int32(parse(f[3]))}] = int(parse(f[4]))
		case len(f) == 8 && f[0] == "native_bubble_column":
			p := BlockPosition{int32(parse(f[4])), int32(parse(f[5])), int32(parse(f[6]))}
			if p.X>>4 != int32(parse(f[2])) || p.Z>>4 != int32(parse(f[3])) || (f[7] != "true" && f[7] != "false") {
				t.Fatalf("invalid native bubble position: %s", scanner.Text())
			}
			want[p] = f[7]
		default:
			t.Fatalf("invalid native bubble reference: %s", scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(counts) != 2 || counts[[2]int32{0, 10}] != 15 || counts[[2]int32{-3, 44}] != 3 || len(want) != 18 {
		t.Fatalf("incomplete official bubble reference: chunks=%v bubbles=%d", counts, len(want))
	}
	dir := filepath.Join(t.TempDir(), "normal-world")
	requestedSeed := seed
	if err := CreateNormalWorld(dir, &requestedSeed); err != nil {
		t.Fatal(err)
	}
	// Match the independently generated reference world's structures option
	// and persist the decoration salt used by the preceding comparison.
	levelBytes, err := readGzip(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	levelName, root, err := nbt.Parse(levelBytes)
	if err != nil {
		t.Fatal(err)
	}
	levelRoot, _ := root.Object()
	data, _ := levelRoot.Object("Data")
	settings, _ := data.Object("WorldGenSettings")
	settings["generate_features"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	marker, _ := data.Object("AtlasWorldgen")
	marker["BeehiveShuffleSalt"] = nbt.Tag{Type: nbt.Long, Value: int64(5393695769601225306)}
	if err := writeGzipNBTAtomic(filepath.Join(dir, "level.dat"), levelName, root); err != nil {
		t.Fatal(err)
	}
	bubbleIDs := make(map[uint32]string)
	for _, drag := range []string{"true", "false"} {
		id, err := BlockStateID("minecraft:bubble_column", map[string]string{"drag": drag})
		if err != nil {
			t.Fatal(err)
		}
		bubbleIDs[id] = drag
	}
	readSaved := func() map[BlockPosition]string {
		got := make(map[BlockPosition]string)
		for c := range counts {
			name, tag, exists, err := anvil.ReadChunk(filepath.Join(dir, "region"), c[0], c[1])
			if err != nil || !exists {
				t.Fatalf("missing generated bubble chunk %v: %v", c, err)
			}
			chunk, err := chunkFromAnvil(name, tag, c[0], c[1])
			if err != nil {
				t.Fatal(err)
			}
			for y := int32(MinY); y < MaxY; y++ {
				for z := 0; z < 16; z++ {
					for x := 0; x < 16; x++ {
						if drag, ok := bubbleIDs[chunk.GetBlock(x, z, y)]; ok {
							got[BlockPosition{c[0]*16 + int32(x), y, c[1]*16 + int32(z)}] = drag
						}
					}
				}
			}
		}
		return got
	}
	positions := make([]BlockPosition, 0, len(want))
	for p := range want {
		positions = append(positions, p)
	}
	load := func(w *World) {
		for c := range counts {
			if _, err := w.LoadChunk(c[0], c[1]); err != nil {
				t.Fatalf("load generated bubble chunk %v: %v", c, err)
			}
		}
	}
	assertLoaded := func(w *World) {
		ids, err := w.ReadLoadedBlocks(positions)
		if err != nil {
			t.Fatal(err)
		}
		for i, p := range positions {
			if drag, ok := bubbleIDs[ids[i]]; !ok || drag != want[p] {
				t.Fatalf("generated/reloaded block at %v = %d, expected bubble_column[drag=%s]", p, ids[i], want[p])
			}
		}
	}
	w, err := Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.Close() })
	load(w)
	assertLoaded(w)
	if got := readSaved(); !reflect.DeepEqual(got, want) {
		t.Fatalf("newly generated Anvil bubbles differ from official save: got=%v want=%v", got, want)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w, err = Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	load(w)
	assertLoaded(w)
	if got := readSaved(); !reflect.DeepEqual(got, want) {
		t.Fatalf("reopened Anvil bubbles differ: got=%v want=%v", got, want)
	}
	// Reopening an existing chunk must preserve a player's replacement,
	// rather than running generation postprocessing over the saved world.
	edit := BlockPosition{13, 7, 175}
	if err := w.SetBlock(edit.X, edit.Y, edit.Z, StateStone); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w, err = Open(noViewers{}, 2, dir)
	if err != nil {
		t.Fatal(err)
	}
	load(w)
	ids, err := w.ReadLoadedBlocks([]BlockPosition{edit})
	if err != nil || ids[0] != StateStone {
		t.Fatalf("player edit was lost after restart: ids=%v error=%v", ids, err)
	}
	delete(want, edit)
	if got := readSaved(); !reflect.DeepEqual(got, want) {
		t.Fatalf("saved bubbles/player edit differ after restart: got=%v want=%v", got, want)
	}
	t.Log("matched all 18 official bubble states across 2 newly generated chunks; Anvil reopen and player edit survived restart")
}
