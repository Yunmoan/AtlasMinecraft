package server

import (
	"atlas/internal/item"
	"atlas/internal/world"
	"testing"
)

func serverTestPortal(t *testing.T, w *world.World, base world.BlockPosition, axis string, fill bool) {
	t.Helper()
	obsidian, _ := world.BlockStateID("minecraft:obsidian", nil)
	portal, _ := world.BlockStateID("minecraft:nether_portal", map[string]string{"axis": axis})
	for x := int32(-1); x <= 2; x++ {
		for y := int32(-1); y <= 3; y++ {
			p := base
			p.Y += y
			if axis == "x" {
				p.X += x
			} else {
				p.Z += x
			}
			state := obsidian
			if x >= 0 && x < 2 && y >= 0 && y < 3 {
				state = world.StateAir
				if fill {
					state = portal
				}
			}
			if err := w.SetBlock(p.X, p.Y, p.Z, state); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestPortalIgnitesThroughCreativeItemInteraction(t *testing.T) {
	for _, name := range []string{"minecraft:flint_and_steel", "minecraft:fire_charge"} {
		t.Run(name, func(t *testing.T) {
			s, p := interactionServer(t)
			base := world.BlockPosition{X: 0, Y: -60, Z: 0}
			serverTestPortal(t, s.World, base, "x", false)
			p.SetPos(.5, -60, 1.5, 0, 0, false)
			id, _ := item.ID(name)
			inventory, _ := p.Inventory()
			inventory.Slots[36] = item.New(id, 1)
			s.setInventory(p, inventory)
			s.handlePlace(p, placePayload(0, -61, 0, 1, 700, .5))
			assertAck(t, p, 700)
			state, err := s.World.BlockAt(0, -60, 0)
			if err != nil {
				t.Fatal(err)
			}
			block, err := world.BlockStateByID(state)
			if err != nil || block.Name != "minecraft:nether_portal" || block.Properties["axis"] != "x" {
				t.Fatal("item did not ignite valid frame", block, err)
			}
		})
	}
}

func TestPlayerPortalTravelRotationCooldownAndSave(t *testing.T) {
	s, p, _ := dimensionTestServer(t)
	nether, _ := s.WorldInDimension("minecraft:the_nether")
	serverTestPortal(t, s.World, world.BlockPosition{X: 0, Y: -60, Z: 0}, "x", true)
	serverTestPortal(t, nether, world.BlockPosition{X: 0, Y: 80, Z: 0}, "z", true)
	p.SetPos(.8, -60, .5, 30, 15, false)
	for tick := 0; tick < 2; tick++ {
		if !p.BeginPortalCheck() {
			t.Fatal("portal check overlap")
		}
		s.checkPlayerPortal(p, 1)
		if tick == 0 && p.Dimension() != "minecraft:overworld" {
			t.Fatal("creative portal activated one tick early")
		}
	}
	if p.Dimension() != "minecraft:the_nether" {
		t.Fatal("portal contact did not transfer")
	}
	_, _, _, yaw, pitch, _ := p.Pos()
	if yaw != 120 || pitch != 15 || p.PortalCooldown() != 10 {
		t.Fatal("arrival rotation/cooldown incorrect", yaw, pitch, p.PortalCooldown())
	}
	p.ConfirmTeleport(s.nextTeleportID.Load())
	for i := 0; i < 3; i++ {
		p.BeginPortalCheck()
		s.checkPlayerPortal(p, 1)
	}
	if p.Dimension() != "minecraft:the_nether" || p.PortalCooldown() != 9 {
		t.Fatal("standing in destination portal caused return or wrong cooldown", p.PortalCooldown())
	}
	s.savePlayer(p)
	saved, exists, err := s.World.LoadPlayer(p.UUID)
	if err != nil || !exists || saved.Dimension != "minecraft:the_nether" || saved.PortalCooldown != 9 {
		t.Fatal("portal save lost dimension/cooldown", saved, err)
	}
}

func TestPlayerEndPortalArrivalUsesOriginalPlatform(t *testing.T) {
	s, p, _ := dimensionTestServer(t)
	state, _ := world.BlockStateID("minecraft:end_portal", nil)
	if err := s.World.SetBlock(0, -60, 0, state); err != nil {
		t.Fatal(err)
	}
	p.SetPos(.5, -59.5, .5, 15, 20, false)
	p.BeginPortalCheck()
	s.checkPlayerPortal(p, 1)
	x, y, z, yaw, pitch, _ := p.Pos()
	if p.Dimension() != "minecraft:the_end" || x != 100.5 || y != 49 || z != .5 || yaw != 90 || pitch != 0 {
		t.Fatal("End portal arrival incorrect", p.Dimension(), x, y, z, yaw, pitch)
	}
	end, _ := s.WorldInDimension("minecraft:the_end")
	block, err := end.BlockAt(100, 48, 0)
	obsidian, _ := world.BlockStateID("minecraft:obsidian", nil)
	if err != nil || block != obsidian {
		t.Fatal("End portal omitted arrival platform", err)
	}
}
