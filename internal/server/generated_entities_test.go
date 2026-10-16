package server

import (
	"atlas/internal/nbt"
	"atlas/internal/player"
	"atlas/internal/world"
	"bufio"
	"bytes"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestGeneratedEntityPacketsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/generated_entity_packets_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	uuid := [16]byte{0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	scanner := bufio.NewScanner(file)
	spawnCount, metadataCount, typeCount := 0, 0, 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		switch fields[0] {
		case "generated_type":
			id, err := strconv.Atoi(fields[2])
			if err != nil {
				t.Fatal(err)
			}
			if generatedEntityTypes[fields[1]] != int32(id) {
				t.Fatal("entity type mismatch", fields[1])
			}
			typeCount++
		case "generated_spawn":
			yaw, err := strconv.ParseFloat(fields[2], 32)
			if err != nil {
				t.Fatal(err)
			}
			pitch, err := strconv.ParseFloat(fields[3], 32)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := hex.DecodeString(fields[4])
			if err != nil {
				t.Fatal(err)
			}
			entity := world.WorldEntity{UUID: uuid, Type: fields[1], Position: [3]float64{-17.5, 68, 32.5}, Rotation: [2]float32{float32(yaw), float32(pitch)}}
			if actual := generatedEntitySpawnPayload(77, entity); !bytes.Equal(actual, expected) {
				t.Fatalf("spawn %s yaw=%v pitch=%v: %x != %x", entity.Type, yaw, pitch, actual, expected)
			}
			spawnCount++
		case "generated_crystal_metadata":
			bottom, err := strconv.ParseBool(fields[1])
			if err != nil {
				t.Fatal(err)
			}
			beam, err := strconv.Atoi(fields[2])
			if err != nil {
				t.Fatal(err)
			}
			expected, err := hex.DecodeString(fields[3])
			if err != nil {
				t.Fatal(err)
			}
			data := nbt.Object{"ShowBottom": {Type: nbt.Byte, Value: int8(0)}}
			if bottom {
				data["ShowBottom"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
			}
			if beam == 1 {
				data["beam_target"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{-100, 50, 123}}
			}
			if beam == 2 {
				data["beam_target"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{30000000, -64, -30000000}}
			}
			entity := world.WorldEntity{Type: "minecraft:end_crystal", Data: data}
			if actual := generatedEntityMetadata(77, entity); !bytes.Equal(actual, expected) {
				t.Fatalf("crystal metadata: %x != %x", actual, expected)
			}
			metadataCount++
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if spawnCount != 200 || metadataCount != 6 || typeCount != 5 {
		t.Fatal("incomplete entity oracle", spawnCount, metadataCount, typeCount)
	}
}

func TestGeneratedEntityTrackingAfterChunksAndDimensionChange(t *testing.T) {
	s, p, _ := dimensionTestServer(t)
	s.nextEID.Store(100)
	end, _ := s.WorldInDimension("minecraft:the_end")
	entities, err := end.EntitiesInChunk(2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(entities) != 1 || entities[0].Type != "minecraft:end_crystal" {
		t.Fatalf("actual End spike crystal missing: %v", entities)
	}
	p.SetInitialDimension("minecraft:the_end")
	p.SetPos(40.5, 70, .5, 0, 0, true)
	s.updateGeneratedEntityVisibility(p)
	if len(s.generatedEntities) != 0 {
		t.Fatal("entity spawned before its chunk")
	}
	p.MarkChunkSent(2, 0)
	s.updateGeneratedEntityVisibility(p)
	key := generatedEntityKey{"minecraft:the_end", entities[0].UUID}
	entry := s.generatedEntities[key]
	if entry == nil || entry.viewers != 1 || !p.IsVisible(entry.id) {
		t.Fatal("generated entity not tracked")
	}
	q := player.New("EndViewer", [16]byte{4}, 4, nil)
	q.SetInitialDimension("minecraft:the_end")
	q.SetPos(40.5, 70, .5, 0, 0, true)
	q.SetPlayReady(true)
	q.MarkChunkSent(2, 0)
	s.Hub.Join(q)
	s.updateGeneratedEntityVisibility(q)
	if !q.IsVisible(entry.id) || entry.viewers != 2 {
		t.Fatal("viewers got different entity IDs")
	}
	s.updateGeneratedEntityVisibility(p)
	if entry.viewers != 2 {
		t.Fatal("repeat chunk event duplicated entity")
	}
	if err := s.TeleportToDimension(p, "minecraft:overworld", [3]float64{.5, -60, .5}); err != nil {
		t.Fatal(err)
	}
	if entry.viewers != 1 || p.IsVisible(entry.id) || !q.IsVisible(entry.id) {
		t.Fatal("dimension transfer left ghost or removed another viewer's entity")
	}
	q.PruneChunks(100, 100, 4)
	s.updateGeneratedEntityVisibility(q)
	if len(s.generatedEntities) != 0 || q.IsVisible(entry.id) {
		t.Fatal("last viewer leaked generated entity tracking")
	}
}
