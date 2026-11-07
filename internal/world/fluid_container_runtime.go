package world

import (
	"atlas/internal/nbt"
	"atlas/internal/proto"
	"atlas/internal/worldgen"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed fluid_container_sounds_1.21.1.json
var fluidContainerSoundJSON []byte

var fluidContainerSounds = func() map[string][2]float64 {
	var sounds map[string][2]float64
	if err := json.Unmarshal(fluidContainerSoundJSON, &sounds); err != nil {
		panic(err)
	}
	return sounds
}()

func (r *fluidRuntime) FluidSound(p worldgen.FeaturePosition, name string) error {
	sound, ok := fluidContainerSounds[name]
	if !ok {
		return fmt.Errorf("unknown fluid container sound %s", name)
	}
	// Level.playSound draws a seed even when no player receives the packet.
	seed := r.world.tickRandom.NextLong()
	// Registry holders encode the registered ID plus one. Coordinates are
	// eighths of a block; extinguishing uses its block's center and BLOCKS=4.
	payload := proto.NewWriter().VarInt(int32(sound[0]) + 1).VarInt(4).
		I32(p.X*8 + 4).I32(p.Y*8 + 4).I32(p.Z*8 + 4).F32(1).F32(1).I64(int64(seed)).Payload()
	if listeners, ok := r.world.hub.(interface {
		SoundListeners(float64, float64, float64, float64) []PacketSink
	}); ok {
		for _, sink := range listeners.SoundListeners(float64(p.X)+.5, float64(p.Y)+.5, float64(p.Z)+.5, sound[1]) {
			sink.SendPacket(proto.SSoundEffect, payload)
		}
	}
	return nil
}

func (r *fluidRuntime) FluidBlockChangeEvent(worldgen.FeaturePosition) error {
	// The existing level has no game-event listener engine. This callback is
	// kept distinct from neighbor updates: sculk/vibration propagation remains
	// part of the pending general block/entity tick implementation.
	return nil
}

type fluidCampfireDowse struct {
	p      worldgen.FeaturePosition
	result chan error
}

func (r *fluidRuntime) FluidCampfireDowse(p worldgen.FeaturePosition) error {
	result := make(chan error, 1)
	if err := r.world.post(context.Background(), p.X>>4, p.Z>>4, fluidCampfireDowse{p, result}); err != nil {
		return err
	}
	select {
	case err := <-result:
		return err
	case <-r.world.done:
		return ErrWorldClosed
	}
}

func (r *Region) dowseCampfire(p worldgen.FeaturePosition) error {
	c := r.chunks[chunkKey{p.X >> 4, p.Z >> 4}]
	if c == nil {
		return ErrChunkNotLoaded
	}
	root, _ := c.anvilRoot.Object()
	entities, _ := root["block_entities"].Elements()
	for _, tag := range entities {
		entity, _ := tag.Object()
		id, _ := entity.Text("id")
		x, _ := entity.Number("x")
		y, _ := entity.Number("y")
		z, _ := entity.Number("z")
		if id != "minecraft:campfire" || x != int64(p.X) || y != int64(p.Y) || z != int64(p.Z) {
			continue
		}
		// CampfireBlockEntity.dowse only marks the entity changed and requests
		// a client update. It does not clear food or reset cooking progress.
		c.dirty = true
		if err := r.world.store.vanilla.persistChunk(c); err != nil {
			return err
		}
		update, err := nbt.MarshalNetwork(blockEntityUpdateTag(entity))
		if err != nil {
			return err
		}
		payload := proto.NewWriter().I64(proto.PackPosition(p.X, p.Y, p.Z)).VarInt(generatedBlockEntityTypes[id]).Raw(update).Payload()
		if r.world.hub != nil {
			for _, sink := range r.world.hub.ViewersOf(p.X, p.Z) {
				if chunkDelivered(sink, c.cx, c.cz) {
					sink.SendPacket(proto.SBlockEntityData, payload)
				}
			}
		}
		return nil
	}
	return nil
}
