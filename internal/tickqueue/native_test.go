package tickqueue

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

type nativeTick struct {
	Type     string   `json:"type"`
	Pos      [3]int32 `json:"pos"`
	Trigger  int64    `json:"trigger"`
	Delay    int32    `json:"delay"`
	Priority int32    `json:"priority"`
	Order    int64    `json:"order"`
}

func (t nativeTick) position() Position { return Position{t.Pos[0], t.Pos[1], t.Pos[2]} }
func (t nativeTick) scheduled() Scheduled[string] {
	return Scheduled[string]{t.Type, t.position(), t.Trigger, ClampPriority(t.Priority), t.Order}
}
func nativeSaved(ticks []nativeTick) []Saved[string] {
	result := make([]Saved[string], 0, len(ticks))
	for _, t := range ticks {
		result = append(result, Saved[string]{t.Type, t.position(), t.Delay, ClampPriority(t.Priority)})
	}
	return result
}

type nativeEvent struct {
	Tick      nativeTick   `json:"tick"`
	Has       bool         `json:"has"`
	Will      bool         `json:"will"`
	QueryHas  bool         `json:"queryHas"`
	QueryWill bool         `json:"queryWill"`
	HasAfter  bool         `json:"hasAfter"`
	Actions   []nativeTick `json:"actions"`
}
type nativeChunk struct {
	X, Z  int32
	Saved []nativeTick
	Has   []bool
}
type nativeOperation struct {
	Op                 string
	X, Z               int32
	Time               int64
	Limit, Mode, Count int
	Tick               nativeTick
	Saved              []nativeTick
	Chunks             []nativeChunk
	Events             []nativeEvent
}
type nativeCase struct {
	Seed, Count int
	Ties        bool
	Operations  []nativeOperation
}

func TestScheduledTicksAgainstVanilla(t *testing.T) {
	f, err := os.Open("../../testdata/worldgen/scheduled_ticks_1.21.1.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	scan := bufio.NewScanner(z)
	scan.Buffer(make([]byte, 64*1024), 32*1024*1024)
	cases := 0
	for scan.Scan() {
		var row nativeCase
		if err = json.Unmarshal(scan.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		cases++
		t.Run(fmt.Sprintf("seed%d/chunks%d/ties%t", row.Seed, row.Count, row.Ties), func(t *testing.T) {
			level := NewLevel[string]()
			chunks := make(map[uint64]*Chunk[string])
			for step, op := range row.Operations {
				k := ChunkKey(op.X, op.Z)
				switch op.Op {
				case "load":
					c := NewChunk(nativeSaved(op.Saved))
					chunks[k] = c
					level.Add(op.X, op.Z, c)
				case "unload":
					if level.Remove(op.X, op.Z) != chunks[k] {
						t.Fatalf("step %d removed wrong owner", step)
					}
					delete(chunks, k)
				case "unpack":
					chunks[k].Unpack(op.Time)
				case "schedule":
					level.Schedule(op.Tick.scheduled())
				case "snapshot":
					if level.Len() != op.Count {
						t.Fatalf("step %d count got %d want %d", step, level.Len(), op.Count)
					}
					if len(chunks) != len(op.Chunks) {
						t.Fatalf("step %d chunk count", step)
					}
					for _, want := range op.Chunks {
						c := chunks[ChunkKey(want.X, want.Z)]
						if c == nil {
							t.Fatalf("step %d missing chunk", step)
						}
						got := c.Save(op.Time)
						expected := nativeSaved(want.Saved)
						if !reflect.DeepEqual(got, expected) {
							t.Fatalf("step %d chunk %d,%d saved\ngot  %+v\nwant %+v", step, want.X, want.Z, got, expected)
						}
						for i, tick := range expected {
							if level.Has(tick.Position, tick.Type) != want.Has[i] {
								t.Fatalf("step %d pending identity differs %+v", step, tick)
							}
						}
					}
				case "tick":
					index := 0
					level.Tick(op.Time, op.Limit, func(x, z int32) bool { return op.Mode == 0 || x&1 == 0 }, func(got Scheduled[string]) {
						if index >= len(op.Events) {
							t.Fatalf("step %d unexpected callback %+v", step, got)
						}
						want := op.Events[index]
						index++
						if got != want.Tick.scheduled() {
							t.Fatalf("step %d callback %d got %+v want %+v", step, index, got, want.Tick.scheduled())
						}
						query := Position{0, 64, 0}
						actual := [4]bool{level.Has(got.Position, got.Type), level.WillTick(got.Position, got.Type), level.Has(query, "water"), level.WillTick(query, "water")}
						expected := [4]bool{want.Has, want.Will, want.QueryHas, want.QueryWill}
						if actual != expected {
							t.Fatalf("step %d callback %d membership got %v want %v", step, index, actual, expected)
						}
						for _, action := range want.Actions {
							level.Schedule(action.scheduled())
						}
						if level.Has(got.Position, got.Type) != want.HasAfter {
							t.Fatalf("step %d callback %d reschedule membership", step, index)
						}
					})
					if index != len(op.Events) {
						t.Fatalf("step %d callbacks got %d want %d", step, index, len(op.Events))
					}
					if level.WillTick(Position{0, 64, 0}, "water") {
						t.Fatalf("step %d stale collected tick", step)
					}
				default:
					t.Fatalf("unknown operation %q", op.Op)
				}
			}
		})
	}
	if err = scan.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 168 {
		t.Fatalf("got %d original scenarios", cases)
	}
}
