package world

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type SpawnPreparationProgress struct {
	Completed, Total int
}

// ensureChunk uses the actor's existing generation coalescing without encoding
// a network packet. Loaded chunks and their edits are retained as-is.
func (w *World) ensureChunk(ctx context.Context, cx, cz int32) error {
	result := make(chan error, 1)
	if err := w.post(ctx, cx, cz, chunkEnsure{cx: cx, cz: cz, result: result}); err != nil {
		return err
	}
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-w.done:
		return ErrWorldClosed
	}
}

func spawnPreparationChunks(cx, cz int32, radius int) []chunkKey {
	chunks := make([]chunkKey, 0, (2*radius+1)*(2*radius+1))
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			chunks = append(chunks, chunkKey{cx + int32(x), cz + int32(z)})
		}
	}
	// Nearest first, stable coordinate tie-breaking. Neighboring jobs share
	// the generator's immutable terrain cache rather than independent halos.
	sort.Slice(chunks, func(i, j int) bool {
		a, b := chunks[i], chunks[j]
		ax, az, bx, bz := int64(a.x-cx), int64(a.z-cz), int64(b.x-cx), int64(b.z-cz)
		ad, bd := ax*ax+az*az, bx*bx+bz*bz
		if ad != bd {
			return ad < bd
		}
		if a.x != b.x {
			return a.x < b.x
		}
		return a.z < b.z
	})
	return chunks
}

// PrepareSpawn waits for the entire square around spawn to be loaded and saved
// before players can join. Radius is in chunks; zero prepares the center only.
// At most w.workers chunk requests are in flight, regardless of the radius.
// progress is called sequentially by the calling goroutine, including 0/total.
func (w *World) PrepareSpawn(ctx context.Context, cx, cz int32, radius int, progress func(SpawnPreparationProgress)) error {
	if radius < 0 || radius > 16 {
		return fmt.Errorf("spawn preload radius must be 0..16 chunks")
	}
	if cx < -1874998+int32(radius) || cx >= 1874998-int32(radius) || cz < -1874998+int32(radius) || cz >= 1874998-int32(radius) {
		return fmt.Errorf("spawn preparation neighborhood is outside world bounds")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chunks := spawnPreparationChunks(cx, cz, radius)
	if progress != nil {
		progress(SpawnPreparationProgress{Total: len(chunks)})
	}
	// Center readiness also updates a new save's SpawnY before other work.
	if err := w.ensureChunk(ctx, cx, cz); err != nil {
		return fmt.Errorf("prepare spawn chunk %d,%d: %w", cx, cz, err)
	}
	completed := 1
	if progress != nil {
		progress(SpawnPreparationProgress{Completed: completed, Total: len(chunks)})
	}
	if err := w.prepareChunkSet(ctx, chunks[1:], func() {
		completed++
		if progress != nil {
			progress(SpawnPreparationProgress{Completed: completed, Total: len(chunks)})
		}
	}); err != nil {
		return err
	}
	// Save the final neighbor-derived light snapshots before reporting readiness.
	return w.Save()
}

// PrepareDestination loads the immediate player neighborhood in parallel.
// Unlike startup preparation, its center can share terrain work with neighbors
// immediately; the saved spawn height does not depend on this operation.
func (w *World) PrepareDestination(ctx context.Context, cx, cz int32) error {
	if cx < -1875000 || cx >= 1875000 || cz < -1875000 || cz >= 1875000 {
		return fmt.Errorf("destination is outside world bounds")
	}
	chunks := spawnPreparationChunks(cx, cz, 1)
	valid := chunks[:0]
	for _, chunk := range chunks {
		// An imported full center can be loaded all the way to the world
		// border, even where new normal terrain lacks a complete 5x5 halo.
		if chunk == (chunkKey{cx, cz}) || chunk.x >= -1874998 && chunk.x < 1874998 && chunk.z >= -1874998 && chunk.z < 1874998 {
			valid = append(valid, chunk)
		}
	}
	return w.prepareChunkSet(ctx, valid, nil)
}

func (w *World) prepareChunkSet(ctx context.Context, chunks []chunkKey, onReady func()) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan chunkKey)
	results := make(chan error, w.workers)
	var wg sync.WaitGroup
	for i := 0; i < min(w.workers, len(chunks)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				err := w.ensureChunk(ctx, p.x, p.z)
				if err != nil {
					err = fmt.Errorf("prepare chunk %d,%d: %w", p.x, p.z, err)
				}
				select {
				case results <- err:
				case <-ctx.Done():
					return
				}
				if err != nil {
					return
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs)
		for _, p := range chunks {
			select {
			case jobs <- p:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { wg.Wait(); close(results) }()
	var firstErr error
	for err := range results {
		if err != nil {
			if firstErr == nil {
				firstErr = err
				cancel()
			}
			continue
		}
		if onReady != nil {
			onReady()
		}
	}
	if firstErr != nil {
		return firstErr
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
