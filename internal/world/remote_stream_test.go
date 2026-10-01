package world

import (
	"atlas/internal/proto"
	"context"
	"os"
	"testing"
	"time"
)

type remoteChunkSink struct {
	received chan [2]int32
	cx, cz   int32
}

func (s *remoteChunkSink) SendPacket(int32, []byte) {}
func (s *remoteChunkSink) SendChunkData(payload []byte) {
	r := proto.NewReader(payload)
	x, e := r.I32()
	if e != nil {
		return
	}
	z, e := r.I32()
	if e != nil {
		return
	}
	s.received <- [2]int32{x, z}
}
func (s *remoteChunkSink) ChunkPos() (int32, int32) { return s.cx, s.cz }

// Opt-in acceptance timing at the coordinates in the reported client image.
func TestRemoteChunkStreamingAtReportedCoordinates(t *testing.T) {
	if os.Getenv("ATLAS_REMOTE_STREAM_BENCH") != "1" {
		t.Skip("set ATLAS_REMOTE_STREAM_BENCH=1 to measure cold streaming")
	}
	seed := int64(1234)
	dir := t.TempDir()
	if err := CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		w, err := Open(noViewers{}, 8, dir)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		if pass == 0 {
			if err := w.PrepareSpawn(ctx, 0, 0, 0, nil); err != nil {
				cancel()
				w.Close()
				t.Fatal(err)
			}
		}
		sink := &remoteChunkSink{received: make(chan [2]int32, 49), cx: 625045, cz: 624993}
		started := time.Now()
		for _, k := range spawnPreparationChunks(sink.cx, sink.cz, 3) {
			w.RequestChunk(sink, k.x, k.z)
		}
		seen := map[[2]int32]bool{}
		near := 0
		for len(seen) < 49 {
			select {
			case k := <-sink.received:
				if seen[k] {
					t.Fatalf("duplicate delivery %v", k)
				}
				seen[k] = true
				if len(seen) == 1 {
					t.Logf("pass %d first chunk %v after %s", pass, k, time.Since(started))
				}
				dx, dz := k[0]-sink.cx, k[1]-sink.cz
				if dx >= -1 && dx <= 1 && dz >= -1 && dz <= 1 {
					near++
					if near == 9 {
						t.Logf("pass %d immediate 3x3 delivered after %s", pass, time.Since(started))
					}
				}
			case <-ctx.Done():
				cancel()
				w.Close()
				t.Fatal(ctx.Err())
			}
		}
		t.Logf("pass %d all 49 chunks delivered after %s", pass, time.Since(started))
		cancel()
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
