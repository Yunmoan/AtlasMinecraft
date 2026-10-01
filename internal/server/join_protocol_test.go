package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"bytes"
	"context"
	"math"
	"os"
	"testing"
	"time"
)

type joinTimings struct {
	configuration, loadStart, position, firstChunk, nearChunks, allChunks time.Duration
}

// Observe the compressed TCP stream, including the loading-screen event and
// completed batches. Queueing a chunk on the server is not delivery evidence.
func awaitJoinChunks(t *testing.T, client *trackingClient, center [2]int32, count int, started time.Time) joinTimings {
	t.Helper()
	timings := joinTimings{configuration: time.Since(started)}
	login, position, view, batch := false, false, false, false
	loadStarts, batchChunks, completed, near := 0, 0, 0, 0
	seen := make(map[[2]int32]bool)
	client.await(t, func(packet player.OutPacket) bool {
		r := proto.NewReader(packet.Payload)
		switch packet.ID {
		case proto.SLogin:
			if login {
				t.Fatal("duplicate play login")
			}
			login = true
		case proto.SGameStateChange:
			reason, err := r.U8()
			if err != nil {
				t.Fatal(err)
			}
			if reason == 13 { // level_chunks_load_start in protocol 767
				value, err := r.F32()
				if err != nil || len(packet.Payload) != 5 || value != 0 || !login || position || len(seen) != 0 {
					t.Fatal("malformed or out-of-order level chunks load start")
				}
				loadStarts++
				timings.loadStart = time.Since(started)
			}
		case proto.SPositionSync:
			if !login || loadStarts != 1 || position {
				t.Fatal("initial position arrived without exactly one level chunks load start")
			}
			position = true
			x, err := r.F64()
			if err != nil {
				t.Fatal(err)
			}
			r.F64()
			z, err := r.F64()
			if err != nil || int32(math.Floor(x/16)) != center[0] || int32(math.Floor(z/16)) != center[1] {
				t.Fatal("initial position does not match the expected current chunk")
			}
			timings.position = time.Since(started)
		case proto.SSetViewPosition:
			x, ex := r.VarInt()
			z, ez := r.VarInt()
			if !position || ex != nil || ez != nil || [2]int32{x, z} != center {
				t.Fatal("wrong or out-of-order initial view center")
			}
			view = true
		case proto.SChunkBatchStart:
			if !view || batch {
				t.Fatal("chunk batch started before view center, or nested")
			}
			batch, batchChunks = true, 0
		case proto.SChunkData:
			x, ex := r.I32()
			z, ez := r.I32()
			key := [2]int32{x, z}
			if !batch || ex != nil || ez != nil || seen[key] {
				t.Fatal("invalid, duplicate or unbatched initial chunk")
			}
			if len(seen) == 0 {
				if key != center {
					t.Fatalf("first chunk %v is not the player's current chunk %v", key, center)
				}
				timings.firstChunk = time.Since(started)
			}
			seen[key] = true
			batchChunks++
			dx, dz := x-center[0], z-center[1]
			if dx >= -1 && dx <= 1 && dz >= -1 && dz <= 1 {
				near++
				if near == 9 {
					timings.nearChunks = time.Since(started)
				}
			}
		case proto.SChunkBatchFinished:
			n, err := r.VarInt()
			if !batch || err != nil || n != int32(batchChunks) || n == 0 {
				t.Fatal("incorrect or incomplete initial chunk batch")
			}
			batch = false
			completed += batchChunks
			if completed == count {
				if loadStarts != 1 || near != 9 {
					t.Fatal("initial stream missing the load event or immediate neighbors")
				}
				timings.allChunks = time.Since(started)
				return true
			}
		}
		return false
	})
	t.Logf("TCP join from connection: config=%s load-start=%s position=%s first-chunk=%s 3x3=%s all-%d=%s",
		timings.configuration, timings.loadStart, timings.position, timings.firstChunk, timings.nearChunks, count, timings.allChunks)
	return timings
}

func TestTCPJoinStartsLevelLoadingBeforePositionAndChunks(t *testing.T) {
	for _, saved := range []bool{false, true} {
		name := "new-player"
		if saved {
			name = "saved-away-from-spawn"
		}
		t.Run(name, func(t *testing.T) {
			dir, seed := t.TempDir(), int64(1234)
			if err := world.CreateFlatWorld(dir, &seed); err != nil {
				t.Fatal(err)
			}
			h := hub.New()
			w, err := world.Open(h, 4, dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { w.Close() })
			if err := w.PrepareSpawn(context.Background(), 0, 0, 3, nil); err != nil {
				t.Fatal(err)
			}
			sx, sz := w.SpawnXZ()
			center := [2]int32{int32(math.Floor(float64(sx) / 16)), int32(math.Floor(float64(sz) / 16))}
			if saved {
				// Negative fractional coordinates must select the chunk containing
				// the saved player, rather than the boot-prepared spawn chunk.
				center = [2]int32{-11, 13}
				if err := w.SavePlayer(offlineUUID("JoinProbe"), world.PlayerPosition{X: -160.25, Y: -60, Z: 208.5}); err != nil {
					t.Fatal(err)
				}
			}
			s := New(Config{MaxPlayers: 10, ViewDistance: 3, SimulationDistance: 3}, h, w)
			address := serveTrackingConnections(t, s, 1)
			started := time.Now()
			client := connectTrackingClient(t, address, "JoinProbe")
			awaitJoinChunks(t, client, center, 45, started)
			// Periodic weather/time sync must not restart the loading screen.
			p := h.FindName("JoinProbe")
			if p == nil {
				t.Fatal("joined player missing")
			}
			s.sendWorldState(p)
			marker := systemChat("world-state-synced")
			p.SendPacket(proto.SSystemChat, marker)
			client.await(t, func(packet player.OutPacket) bool {
				if packet.ID == proto.SGameStateChange && len(packet.Payload) > 0 && packet.Payload[0] == 13 {
					t.Fatal("world state sync repeated the join-only level loading event")
				}
				return packet.ID == proto.SSystemChat && bytes.Equal(packet.Payload, marker)
			})
		})
	}
}

// Explicit timing run: keep generation and disk loading distinct from packet
// delivery. No hard speed assertions: host load is not a protocol regression.
func TestNormalWorldTCPJoinLoadingTimings(t *testing.T) {
	if os.Getenv("ATLAS_JOIN_STREAM_BENCH") != "1" {
		t.Skip("set ATLAS_JOIN_STREAM_BENCH=1 to measure initial TCP world delivery")
	}
	dir, seed := t.TempDir(), int64(1234)
	if err := world.CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	for _, pass := range []string{"prepared", "restart-disk"} {
		t.Run(pass, func(t *testing.T) {
			h := hub.New()
			w, err := world.Open(h, 8, dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { w.Close() })
			if pass == "prepared" {
				if err := w.PrepareSpawn(context.Background(), 0, 0, 3, nil); err != nil {
					t.Fatal(err)
				}
			}
			sx, sz := w.SpawnXZ()
			center := [2]int32{int32(math.Floor(float64(sx) / 16)), int32(math.Floor(float64(sz) / 16))}
			s := New(Config{MaxPlayers: 10, ViewDistance: 3, SimulationDistance: 3}, h, w)
			address := serveTrackingConnections(t, s, 1)
			started := time.Now()
			client := connectTrackingClient(t, address, "JoinProbe")
			awaitJoinChunks(t, client, center, 45, started)
		})
	}
}
