package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"context"
	"math"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

type trackingClient struct {
	conn      *proto.FramedConn
	packets   chan player.OutPacket
	errors    chan error
	delivered map[[2]int32]bool
	writeMu   sync.Mutex
}

func (c *trackingClient) send(id int32, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WritePacket(id, payload)
}

func connectTrackingClient(t *testing.T, address, name string) *trackingClient {
	t.Helper()
	nc, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { nc.Close() })
	fc := proto.NewFramedConn(nc)
	write := func(id int32, payload []byte) {
		t.Helper()
		if err := fc.WritePacket(id, payload); err != nil {
			t.Fatal(err)
		}
	}
	write(proto.CHandshake, proto.NewWriter().VarInt(767).String("localhost").U16(25565).VarInt(2).Payload())
	write(proto.CLoginStart, proto.NewWriter().String(name).UUID(offlineUUID(name)).Payload())
	nc.SetDeadline(time.Now().Add(10 * time.Second))
	for {
		id, payload, err := fc.ReadPacket()
		if err != nil {
			t.Fatal(err)
		}
		if id == proto.SLoginSetCompress {
			threshold, _ := proto.NewReader(payload).VarInt()
			fc.EnableCompression(threshold)
		} else if id == proto.SLoginSuccess {
			break
		} else {
			t.Fatalf("unexpected login packet %#x", id)
		}
	}
	write(proto.CLoginAcknowledged, nil)
	for {
		id, _, err := fc.ReadPacket()
		if err != nil {
			t.Fatal(err)
		}
		if id == proto.SCfgFinish {
			break
		}
	}
	write(proto.CCfgClientSettings, proto.NewWriter().String("en_us").U8(8).VarInt(0).Bool(true).U8(127).Payload())
	write(proto.CCfgFinishAck, nil)
	nc.SetDeadline(time.Time{})
	c := &trackingClient{conn: fc, packets: make(chan player.OutPacket, 512), errors: make(chan error, 1), delivered: make(map[[2]int32]bool)}
	go func() {
		for {
			id, payload, err := fc.ReadPacket()
			if err != nil {
				c.errors <- err
				return
			}
			switch id {
			case proto.SPositionSync:
				r := proto.NewReader(payload)
				for range 3 {
					r.F64()
				}
				r.F32()
				r.F32()
				r.U8()
				tid, _ := r.VarInt()
				err = c.send(proto.CTeleportConfirm, proto.NewWriter().VarInt(tid).Payload())
			case proto.SKeepAlive:
				err = c.send(proto.CKeepAlive, payload)
			case proto.SChunkBatchFinished:
				err = c.send(proto.CChunkBatchReceived, proto.NewWriter().F32(16).Payload())
			}
			if err != nil {
				c.errors <- err
				return
			}
			c.packets <- player.OutPacket{ID: id, Payload: payload}
		}
	}()
	return c
}

func (c *trackingClient) await(t *testing.T, match func(player.OutPacket) bool) player.OutPacket {
	t.Helper()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case packet := <-c.packets:
			r := proto.NewReader(packet.Payload)
			if packet.ID == proto.SChunkData {
				x, _ := r.I32()
				z, _ := r.I32()
				c.delivered[[2]int32{x, z}] = true
			}
			if packet.ID == proto.SSpawnEntity {
				r.VarInt()
				r.UUID()
				r.VarInt()
				x, _ := r.F64()
				r.F64()
				z, _ := r.F64()
				key := [2]int32{int32(math.Floor(x / 16)), int32(math.Floor(z / 16))}
				if !c.delivered[key] {
					t.Fatalf("entity spawned before its chunk %v arrived over TCP", key)
				}
			}
			if match(packet) {
				return packet
			}
		case err := <-c.errors:
			t.Fatal(err)
		case <-timer.C:
			t.Fatal("timed out waiting for protocol event")
		}
	}
}

func serveTrackingConnections(t *testing.T, s *Server, count int) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var clients sync.WaitGroup
	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		for range count {
			nc, err := ln.Accept()
			if err != nil {
				return
			}
			clients.Add(1)
			go func() { defer clients.Done(); s.handle(nc) }()
		}
	}()
	t.Cleanup(func() { ln.Close(); <-accepted; clients.Wait() })
	return ln.Addr().String()
}

func TestTwoTCPClientsTrackMovementAndFarTeleport(t *testing.T) {
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
	if err := w.PrepareSpawn(context.Background(), 0, 0, 1, nil); err != nil {
		t.Fatal(err)
	}
	s := New(Config{MaxPlayers: 10, ViewDistance: 2, SimulationDistance: 2, Ops: map[string]bool{"probeb": true}}, h, w)
	address := serveTrackingConnections(t, s, 2)
	a := connectTrackingClient(t, address, "ProbeA")
	a.await(t, func(p player.OutPacket) bool { return p.ID == proto.SPositionSync })
	b := connectTrackingClient(t, address, "ProbeB")
	initial := b.await(t, func(p player.OutPacket) bool { return p.ID == proto.SPositionSync })
	r := proto.NewReader(initial.Payload)
	x, _ := r.F64()
	y, _ := r.F64()
	z, _ := r.F64()
	bp := h.FindName("ProbeB")
	if bp == nil {
		t.Fatal("second client missing from hub")
	}
	a.await(t, func(p player.OutPacket) bool {
		if p.ID != proto.SSpawnEntity {
			return false
		}
		id, _ := proto.NewReader(p.Payload).VarInt()
		return id == bp.EID
	})
	var sum int64
	start := x
	for step := 0; step < 100; step++ {
		previous := x
		x += .00037
		if err := b.send(proto.CPlayerPosition, proto.NewWriter().F64(x).F64(y).F64(z).Bool(true).Payload()); err != nil {
			t.Fatal(err)
		}
		packet := a.await(t, func(p player.OutPacket) bool {
			return p.ID == proto.SRelEntityMove || p.ID == proto.SEntityMoveLook || p.ID == proto.SEntityTeleport
		})
		if packet.ID != proto.SRelEntityMove {
			t.Fatalf("ordinary movement used packet %#x", packet.ID)
		}
		r := proto.NewReader(packet.Payload)
		eid, _ := r.VarInt()
		dx, _ := r.I16()
		dy, _ := r.I16()
		dz, _ := r.I16()
		want := int64(math.Floor(x*4096)) - int64(math.Floor(previous*4096))
		if eid != bp.EID || int64(dx) != want || dy != 0 || dz != 0 {
			t.Fatalf("step %d: wrong or duplicate observer movement", step)
		}
		sum += int64(dx)
	}
	if sum != int64(math.Floor(x*4096))-int64(math.Floor(start*4096)) {
		t.Fatal("observer fixed-point position drifted over the network")
	}
	if err := b.send(proto.CChatCommand, proto.NewWriter().String("tp 10000722 -60 9999902").Payload()); err != nil {
		t.Fatal(err)
	}
	b.await(t, func(p player.OutPacket) bool {
		if p.ID != proto.SPositionSync {
			return false
		}
		r := proto.NewReader(p.Payload)
		x, _ := r.F64()
		return x == 10000722
	})
	b.await(t, func(p player.OutPacket) bool {
		return p.ID == proto.SChunkData && b.delivered[[2]int32{625045, 624993}]
	})
	a.await(t, func(p player.OutPacket) bool {
		if p.ID != proto.SEntityDestroy {
			return false
		}
		r := proto.NewReader(p.Payload)
		n, _ := r.VarInt()
		for range n {
			id, _ := r.VarInt()
			if id == bp.EID {
				return true
			}
		}
		return false
	})
	if aPlayer := h.FindName("ProbeA"); aPlayer.IsVisible(bp.EID) {
		t.Fatal("remote entity remains tracked after a far teleport")
	}
	if bp.WantsChunk(0, 0) || bp.IsChunkSent(0, 0) {
		t.Fatal("far teleport retained the old client chunk window")
	}
}

// Includes the command's synchronous destination load, packet compression and
// actual TCP delivery. This complements the world's queue-only timing test.
func TestNormalWorldRemoteTCPStreamingAtReportedCoordinates(t *testing.T) {
	if os.Getenv("ATLAS_REMOTE_STREAM_BENCH") != "1" {
		t.Skip("set ATLAS_REMOTE_STREAM_BENCH=1 to measure remote teleport streaming")
	}
	dir, seed := t.TempDir(), int64(1234)
	if err := world.CreateNormalWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	for _, pass := range []string{"cold", "restart"} {
		t.Run(pass, func(t *testing.T) {
			h := hub.New()
			w, err := world.Open(h, 8, dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { w.Close() })
			if pass == "cold" {
				if err := w.PrepareSpawn(context.Background(), 0, 0, 0, nil); err != nil {
					t.Fatal(err)
				}
			}
			sx, sz := w.SpawnXZ()
			sy, err := w.SurfaceY(sx, sz)
			if err != nil {
				t.Fatal(err)
			}
			if err := w.SavePlayer(offlineUUID("ProbeB"), world.PlayerPosition{X: float64(sx) + .5, Y: float64(sy), Z: float64(sz) + .5}); err != nil {
				t.Fatal(err)
			}
			s := New(Config{MaxPlayers: 10, ViewDistance: 3, SimulationDistance: 3, Ops: map[string]bool{"probeb": true}}, h, w)
			address := serveTrackingConnections(t, s, 1)
			client := connectTrackingClient(t, address, "ProbeB")
			client.await(t, func(p player.OutPacket) bool { return p.ID == proto.SPositionSync })
			client.await(t, func(p player.OutPacket) bool { return p.ID == proto.SChunkData })
			started := time.Now()
			if err := client.send(proto.CChatCommand, proto.NewWriter().String("tp 10000722 89 9999902").Payload()); err != nil {
				t.Fatal(err)
			}
			seen := make(map[[2]int32]bool)
			near := 0
			for len(seen) < 45 {
				packet := client.await(t, func(p player.OutPacket) bool {
					if p.ID != proto.SChunkData {
						return false
					}
					r := proto.NewReader(p.Payload)
					x, _ := r.I32()
					z, _ := r.I32()
					dx, dz := x-625045, z-624993
					return dx >= -3 && dx <= 3 && dz >= -3 && dz <= 3
				})
				r := proto.NewReader(packet.Payload)
				x, _ := r.I32()
				z, _ := r.I32()
				key := [2]int32{x, z}
				if seen[key] {
					t.Fatalf("duplicate remote delivery %v", key)
				}
				seen[key] = true
				if len(seen) == 1 {
					t.Logf("first destination chunk received over TCP after %s", time.Since(started))
				}
				dx, dz := x-625045, z-624993
				if dx >= -1 && dx <= 1 && dz >= -1 && dz <= 1 {
					near++
					if near == 9 {
						t.Logf("immediate 3x3 received over TCP after %s", time.Since(started))
					}
				}
			}
			t.Logf("all 45 requested view chunks received over TCP after %s", time.Since(started))
		})
	}
}
