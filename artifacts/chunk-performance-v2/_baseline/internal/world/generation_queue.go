package world

import (
	"math"
	"sync"
)

type generationTask struct {
	job      genJob
	sinks    []PacketSink
	required bool
	order    uint64
}

// Priorities use current player positions, so a distant teleport never waits
// behind hundreds of obsolete requests from the previous view.
type generationQueue struct {
	mu      sync.Mutex
	pending map[chunkKey]*generationTask
	wake    chan struct{}
	order   uint64
}

func newGenerationQueue() *generationQueue {
	return &generationQueue{pending: map[chunkKey]*generationTask{}, wake: make(chan struct{}, 1)}
}

func chunkWanted(sink PacketSink, cx, cz int32) bool {
	if viewer, ok := sink.(interface{ WantsChunk(int32, int32) bool }); ok {
		return viewer.WantsChunk(cx, cz)
	}
	return true
}

func (q *generationQueue) enqueue(job genJob, sinks []PacketSink, required, create bool) {
	q.mu.Lock()
	key := chunkKey{job.cx, job.cz}
	task := q.pending[key]
	if task == nil && create {
		q.order++
		task = &generationTask{job: job, order: q.order}
		q.pending[key] = task
	}
	if task != nil {
		task.sinks = append([]PacketSink(nil), sinks...)
		task.required = required
	}
	q.mu.Unlock()
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

func (q *generationQueue) next(done <-chan struct{}) (genJob, bool) {
	for {
		q.mu.Lock()
		var best *generationTask
		bestPriority := int64(math.MaxInt64)
		for key, task := range q.pending {
			priority := int64(math.MaxInt64)
			if task.required {
				priority = -1
			} else {
				for _, sink := range task.sinks {
					if !chunkWanted(sink, key.x, key.z) {
						continue
					}
					distance := int64(0)
					if viewer, ok := sink.(interface{ ChunkPos() (int32, int32) }); ok {
						cx, cz := viewer.ChunkPos()
						dx, dz := int64(key.x)-int64(cx), int64(key.z)-int64(cz)
						distance = dx*dx + dz*dz
					}
					priority = min(priority, distance)
				}
			}
			if priority == math.MaxInt64 {
				best = task
				best.job.canceled = true
				break
			}
			if best == nil || priority < bestPriority || priority == bestPriority && task.order < best.order {
				best = task
				bestPriority = priority
			}
		}
		if best != nil {
			delete(q.pending, chunkKey{best.job.cx, best.job.cz})
			more := len(q.pending) > 0
			q.mu.Unlock()
			if more {
				select {
				case q.wake <- struct{}{}:
				default:
				}
			}
			select {
			case <-done:
				return genJob{}, false
			default:
				return best.job, true
			}
		}
		q.mu.Unlock()
		select {
		case <-done:
			return genJob{}, false
		case <-q.wake:
		}
	}
}
