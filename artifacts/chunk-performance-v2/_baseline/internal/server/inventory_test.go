package server

import (
	"atlas/internal/hub"
	"atlas/internal/item"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"bytes"
	"context"
	"testing"
	"time"
)

func readInventoryPacket(t *testing.T, packet player.OutPacket) item.Inventory {
	t.Helper()
	r := proto.NewReader(packet.Payload)
	window, e1 := r.U8()
	_, e2 := r.VarInt()
	count, e3 := r.VarInt()
	if e1 != nil || e2 != nil || e3 != nil || window != 0 || count != 46 {
		t.Fatal("wrong inventory content header")
	}
	var v item.Inventory
	for i := range v.Slots {
		stack, err := item.Read(r)
		if err != nil {
			t.Fatal(err)
		}
		v.Slots[i] = stack
	}
	var err error
	v.Carried, err = item.Read(r)
	if err != nil || r.Remaining() != 0 {
		t.Fatal("invalid cursor stack")
	}
	return v
}

func TestTCPInventoryPersistsAcrossLogoutAndRestart(t *testing.T) {
	dir, seed := t.TempDir(), int64(1234)
	if err := world.CreateFlatWorld(dir, &seed); err != nil {
		t.Fatal(err)
	}
	stone, _ := item.ID("stone")
	sword, _ := item.ID("diamond_sword")
	helmet, _ := item.ID("diamond_helmet")
	shield, _ := item.ID("shield")
	expected := map[int]item.Stack{9: item.New(sword, 1), 38: item.New(stone, 31), 5: item.New(helmet, 1), 45: item.New(shield, 1)}
	expected[9] = item.Stack{ID: sword, Count: 1, Patch: proto.NewWriter().VarInt(1).VarInt(0).VarInt(3).VarInt(13).Payload()}
	for pass := 0; pass < 2; pass++ {
		t.Run([]string{"write-and-rejoin", "restart"}[pass], func(t *testing.T) {
			h := hub.New()
			w, err := world.Open(h, 4, dir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { w.Close() })
			if err := w.PrepareSpawn(context.Background(), 0, 0, 1, nil); err != nil {
				t.Fatal(err)
			}
			count := 1
			if pass == 0 {
				count = 2
			}
			s := New(Config{MaxPlayers: 10, ViewDistance: 2, SimulationDistance: 2}, h, w)
			address := serveTrackingConnections(t, s, count)
			client := connectTrackingClient(t, address, "InventoryProbe")
			selected := byte(0)
			if pass != 0 {
				selected = 2
			}
			client.await(t, func(packet player.OutPacket) bool {
				if packet.ID != proto.SHeldItemSlot {
					return false
				}
				if !bytes.Equal(packet.Payload, []byte{selected}) {
					t.Fatal("selected hotbar slot was not restored")
				}
				return true
			})
			initial := readInventoryPacket(t, client.await(t, func(p player.OutPacket) bool { return p.ID == proto.SContainerContent }))
			client.await(t, func(p player.OutPacket) bool { return p.ID == proto.SPositionSync })
			if pass == 0 {
				for _, slot := range []int{9, 38, 5, 45} {
					payload := proto.NewWriter().I16(int16(slot))
					expected[slot].Write(payload)
					if err := client.send(proto.CSetCreativeSlot, payload.Payload()); err != nil {
						t.Fatal(err)
					}
					client.await(t, func(packet player.OutPacket) bool {
						if packet.ID != proto.SContainerSlot {
							return false
						}
						r := proto.NewReader(packet.Payload)
						r.U8()
						r.VarInt()
						got, _ := r.I16()
						return got == int16(slot)
					})
				}
				if err := client.send(proto.CHeldItemSlot, proto.NewWriter().I16(2).Payload()); err != nil {
					t.Fatal(err)
				}
				// A command reply fences the preceding inventory/selection packets.
				if err := client.send(proto.CChatCommand, proto.NewWriter().String("where").Payload()); err != nil {
					t.Fatal(err)
				}
				client.await(t, func(packet player.OutPacket) bool {
					return packet.ID == proto.SSystemChat && bytes.Contains(packet.Payload, []byte("Position:"))
				})
				s.SavePlayers()
				client.conn.Conn.Close()
				deadline := time.Now().Add(5 * time.Second)
				for h.Count() != 0 && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
				if h.Count() != 0 {
					t.Fatal("logout did not finish saving")
				}
				client = connectTrackingClient(t, address, "InventoryProbe")
				client.await(t, func(packet player.OutPacket) bool {
					if packet.ID != proto.SHeldItemSlot {
						return false
					}
					if !bytes.Equal(packet.Payload, []byte{2}) {
						t.Fatal("selected hotbar slot was not restored")
					}
					return true
				})
				initial = readInventoryPacket(t, client.await(t, func(p player.OutPacket) bool { return p.ID == proto.SContainerContent }))
			}
			for slot, stack := range expected {
				if !bytes.Equal(initial.Slots[slot].Bytes(), stack.Bytes()) {
					t.Fatalf("lost saved slot %d: %x != %x", slot, initial.Slots[slot].Bytes(), stack.Bytes())
				}
			}
		})
	}
}

func TestInventoryClicksPreserveCountsAndRejectForgedDeltas(t *testing.T) {
	stone, _ := item.ID("stone")
	sword, _ := item.ID("diamond_sword")
	v := item.Inventory{}
	v.Slots[9] = item.New(stone, 17)
	applyInventoryClick(&v, 9, 1, 0)
	if v.Slots[9].Count != 8 || v.Carried.Count != 9 {
		t.Fatal("right-click split")
	}
	applyInventoryClick(&v, 10, 1, 0)
	if v.Slots[10].Count != 1 || v.Carried.Count != 8 {
		t.Fatal("right-click one item placement")
	}
	applyInventoryClick(&v, 10, 0, 0)
	if v.Slots[10].Count != 9 || !v.Carried.Empty() {
		t.Fatal("left-click merge")
	}
	applyInventoryClick(&v, 10, 0, 1)
	if !v.Slots[10].Empty() || v.Slots[36].Count != 9 {
		t.Fatal("shift-click to hotbar")
	}
	applyInventoryClick(&v, 9, 40, 2)
	if !v.Slots[9].Empty() || v.Slots[45].Count != 8 {
		t.Fatal("offhand swap")
	}
	s, p := interactionServer(t)
	s.setInventory(p, v)
	forged := item.New(sword, 1)
	payload := proto.NewWriter().U8(0).VarInt(0).I16(11).I8(0).VarInt(0).VarInt(1).I16(11)
	forged.Write(payload)
	item.Stack{}.Write(payload)
	s.handleContainerClick(p, payload.Payload())
	got, _ := p.Inventory()
	if !got.Slots[11].Empty() {
		t.Fatal("client delta injected an item")
	}
}
