package world

import (
	"atlas/internal/light"
	"errors"
	"fmt"
	"time"
)

type unloadRegion struct {
	cutoff time.Time
	result chan unloadResult
}

type unloadResult struct {
	count int
	err   error
}

// UnloadIdle releases saved chunks outside every player's view, together with
// their packet buffers and dense lighting columns. A one-chunk border remains
// available for lighting and block interactions. Recent direct API accesses
// are retained for maxIdle; dirty chunks are saved before anything is released.
func (w *World) UnloadIdle(maxIdle time.Duration) (int, error) {
	if maxIdle < 0 {
		return 0, fmt.Errorf("negative chunk idle timeout")
	}
	w.blockUpdateMu.Lock()
	defer w.blockUpdateMu.Unlock()
	w.saveMu.Lock()
	defer w.saveMu.Unlock()
	if w.store == nil {
		return 0, nil // memory-only chunks have no durable backing
	}
	w.mu.Lock()
	regions := make([]*Region, 0, len(w.regions))
	for _, r := range w.regions {
		regions = append(regions, r)
	}
	w.mu.Unlock()
	cutoff := time.Now().Add(-maxIdle)
	count := 0
	var combined error
	for _, r := range regions {
		result := make(chan unloadResult, 1)
		select {
		case r.mailbox <- unloadRegion{cutoff, result}:
		case <-w.done:
			return count, errors.Join(combined, ErrWorldClosed)
		}
		select {
		case unloaded := <-result:
			count += unloaded.count
			combined = errors.Join(combined, unloaded.err)
		case <-w.done:
			return count, errors.Join(combined, ErrWorldClosed)
		}
	}
	return count, combined
}

func (r *Region) chunkInUse(key chunkKey) bool {
	if r.genQueued[key] || len(r.pending[key])+len(r.pendingLoads[key])+len(r.pendingReady[key])+len(r.pendingEdits[key])+len(r.pendingSurfaces[key]) != 0 {
		return true
	}
	if r.world.hub != nil {
		for dx := int32(-1); dx <= 1; dx++ {
			for dz := int32(-1); dz <= 1; dz++ {
				if len(r.world.hub.ViewersOf((key.x+dx)*16, (key.z+dz)*16)) != 0 {
					return true
				}
			}
		}
	}
	return false
}

func (r *Region) unloadUnused(cutoff time.Time) (int, error) {
	count := 0
	var combined error
	for key, ch := range r.chunks {
		if ch.lastAccess.After(cutoff) || r.chunkInUse(key) {
			continue
		}
		persist := func(data *light.Data) error {
			r.world.store.vanilla.refreshStrongholdFlags(ch)
			r.applyLight(ch, data, false)
			if ch.dirty {
				return r.world.store.vanilla.persistChunk(ch)
			}
			return nil
		}
		var err error
		if r.world.lighting != nil {
			err = r.world.lighting.Uninstall(light.Position{X: key.x, Z: key.z}, persist)
		} else {
			err = persist(nil)
		}
		if err != nil {
			combined = errors.Join(combined, fmt.Errorf("unload chunk %d,%d: %w", key.x, key.z, err))
			continue
		}
		r.packets.forget(ch)
		delete(r.chunks, key)
		r.world.loadedChunks.Add(-1)
		count++
	}
	if len(r.chunks) == 0 {
		// Release map backing arrays after leaving a densely explored region.
		r.chunks = make(map[chunkKey]*Chunk)
		r.packets = chunkPacketCache{}
	}
	return count, combined
}
