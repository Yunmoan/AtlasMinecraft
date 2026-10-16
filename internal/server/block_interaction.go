package server

import (
	"atlas/internal/player"
	"atlas/internal/world"
	"strings"
	"time"
)

func oppositeFacing(facing string) string {
	switch facing {
	case "north":
		return "south"
	case "south":
		return "north"
	case "west":
		return "east"
	case "east":
		return "west"
	}
	return facing
}
func doorHinge(facing string, x, z float32) string {
	right := false
	switch facing {
	case "north":
		right = x > .5
	case "south":
		right = x < .5
	case "east":
		right = z > .5
	case "west":
		right = z < .5
	}
	if right {
		return "right"
	}
	return "left"
}

func (s *Server) useBlock(p *player.Player, pos world.BlockPosition, id uint32, state world.BlockState) bool {
	properties := state.Properties
	interactionWorld := s.PlayerWorld(p)
	switch {
	case state.Name == "minecraft:chest" || state.Name == "minecraft:trapped_chest":
		s.openChest(p, pos)
		return true
	case world.IsDoor(state.Name):
		if state.Name == "minecraft:iron_door" {
			return true
		}
		if err := interactionWorld.UpdateDoor(pos, id, properties["open"] != "true"); err != nil {
			s.Dbg("door interaction rejected: %v", err)
		}
		return true
	case strings.HasSuffix(state.Name, "_trapdoor") || strings.HasSuffix(state.Name, "_fence_gate"):
		if state.Name == "minecraft:iron_trapdoor" {
			return true
		}
		if properties["open"] == "true" {
			properties["open"] = "false"
		} else {
			properties["open"] = "true"
		}
		updated, err := world.BlockStateID(state.Name, properties)
		if err == nil {
			err = interactionWorld.TryPlaceBlock(pos, id, updated, nil)
		}
		if err != nil {
			s.Dbg("openable block interaction rejected: %v", err)
		}
		return true
	case state.Name == "minecraft:lever" || strings.HasSuffix(state.Name, "_button"):
		previous := properties["powered"] == "true"
		if strings.HasSuffix(state.Name, "_button") && previous {
			return true
		}
		properties["powered"] = "true"
		if previous {
			properties["powered"] = "false"
		}
		updated, err := world.BlockStateID(state.Name, properties)
		if err == nil {
			err = interactionWorld.TryPlaceBlock(pos, id, updated, nil)
		}
		if err != nil {
			s.Dbg("switch interaction rejected: %v", err)
			return true
		}
		if strings.HasSuffix(state.Name, "_button") {
			delay := time.Second
			if state.Name != "minecraft:stone_button" && state.Name != "minecraft:polished_blackstone_button" {
				delay = 1500 * time.Millisecond
			}
			go func() {
				timer := time.NewTimer(delay)
				defer timer.Stop()
				<-timer.C
				properties["powered"] = "false"
				released, e := world.BlockStateID(state.Name, properties)
				if e == nil {
					_ = interactionWorld.TryPlaceBlock(pos, updated, released, nil)
				}
			}()
		}
		return true
	}
	return false
}
