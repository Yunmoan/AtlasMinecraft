package light

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
)

type Position struct{ X, Z int32 }

// Data is an immutable lighting snapshot, including one padding section above
// and below the dimension. Each non-nil layer is a 2048-byte nibble array.
type Data struct {
	Sky, Block [][]byte
	Complete   bool
	Preserved  bool // freshly computed Atlas data may have unloaded neighbors
	Revision   uint64
}

type column struct {
	states  []uint16
	light   *Data
	trusted bool
}

// Engine owns private block snapshots. It never calls back into chunk actors,
// so neighboring region actors can update lighting without waiting on each
// other. Published Data and its arrays are immutable across worker goroutines.
type Engine struct {
	mu           sync.Mutex
	minY, height int
	hasSky       bool
	rules        *ruleTable
	columns      map[Position]*column
	revision     uint64
	work         *relightWork
	columnCount  atomic.Int64
}

// Mutable scratch is owned under Engine.mu. It is never included in a
// published Data snapshot, so later relighting cannot alter queued packets.
type relightWork struct {
	states               []uint16
	sky, block           []byte
	loaded               [48 * 48]bool
	skyQueue, blockQueue [16][]int
}

func New(minY, height int, hasSky bool) (*Engine, error) {
	if minY%16 != 0 || height%16 != 0 || height <= 0 {
		return nil, fmt.Errorf("invalid light dimension height")
	}
	rules, err := loadRules()
	if err != nil {
		return nil, err
	}
	return &Engine{minY: minY, height: height, hasSky: hasSky, rules: rules, columns: make(map[Position]*column)}, nil
}

func (e *Engine) Snapshot(p Position) *Data {
	e.mu.Lock()
	defer e.mu.Unlock()
	if c := e.columns[p]; c != nil {
		return c.light
	}
	return nil
}

// Uninstall persists the latest immutable light snapshot before releasing its
// dense block column. The callback must not call back into Engine or actors.
// A failed save retains the column. Neighbor snapshots remain valid for their
// unchanged blocks and are reconciled when the neighborhood is loaded again.
func (e *Engine) Uninstall(p Position, persist func(*Data) error) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	c := e.columns[p]
	if c == nil {
		return persist(nil)
	}
	if err := persist(c.light); err != nil {
		return err
	}
	delete(e.columns, p)
	e.columnCount.Add(-1)
	return nil
}

func (e *Engine) LoadedColumns() int {
	return int(e.columnCount.Load())
}

// InitializeArea installs a fresh 3x3 generation neighborhood and solves its
// lighting once. SPAWN needs one complete neighborhood, so solving eight
// intermediate, incomplete neighborhoods provides no useful snapshots.
func (e *Engine) InitializeArea(center Position, states map[Position][]uint16) (map[Position]*Data, error) {
	if _, ok := states[center]; !ok {
		return nil, fmt.Errorf("initial light area missing center")
	}
	for position, values := range states {
		if position.X < center.X-1 || position.X > center.X+1 || position.Z < center.Z-1 || position.Z > center.Z+1 {
			return nil, fmt.Errorf("initial light column outside 3x3 area")
		}
		if len(values) != e.height*256 {
			return nil, fmt.Errorf("incomplete chunk light snapshot")
		}
		for _, state := range values {
			if int(state) >= len(e.rules.states) {
				return nil, fmt.Errorf("unknown light block state %d", state)
			}
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.columns) != 0 {
		return nil, fmt.Errorf("initial light area requires a fresh engine")
	}
	for position, values := range states {
		e.columns[position] = &column{states: append([]uint16(nil), values...)}
	}
	e.columnCount.Store(int64(len(states)))
	return e.relight(center), nil
}

func (e *Engine) Install(p Position, states []uint16, imported *Data) (map[Position]*Data, error) {
	if imported != nil {
		layers := e.height/16 + 2
		if len(imported.Sky) != layers || len(imported.Block) != layers {
			return nil, fmt.Errorf("invalid imported light section count")
		}
		// Own copies even when callers supply mutable imported arrays.
		owned := &Data{Sky: make([][]byte, layers), Block: make([][]byte, layers), Complete: imported.Complete, Preserved: imported.Preserved}
		for i := 0; i < layers; i++ {
			if imported.Sky[i] != nil && len(imported.Sky[i]) != 2048 || imported.Block[i] != nil && len(imported.Block[i]) != 2048 {
				return nil, fmt.Errorf("invalid imported light layer")
			}
			owned.Sky[i] = append([]byte(nil), imported.Sky[i]...)
			owned.Block[i] = append([]byte(nil), imported.Block[i]...)
		}
		imported = owned
	}
	if len(states) != e.height*256 {
		return nil, fmt.Errorf("incomplete chunk light snapshot")
	}
	for _, state := range states {
		if int(state) >= len(e.rules.states) {
			return nil, fmt.Errorf("unknown light block state %d", state)
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.columns[p] != nil {
		return nil, nil
	}
	e.columns[p] = &column{states: append([]uint16(nil), states...), light: imported, trusted: imported != nil && (imported.Complete || imported.Preserved)}
	e.columnCount.Add(1)
	return e.relight(p), nil
}

type Change struct {
	X, Y, Z int
	State   uint16
}

func (e *Engine) Edit(p Position, changes []Change) (map[Position]*Data, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	c := e.columns[p]
	if c == nil {
		return nil, fmt.Errorf("light chunk is not installed")
	}
	for _, change := range changes {
		if change.X < 0 || change.X >= 16 || change.Z < 0 || change.Z >= 16 || change.Y < e.minY || change.Y >= e.minY+e.height || int(change.State) >= len(e.rules.states) {
			return nil, fmt.Errorf("invalid light block edit")
		}
	}
	relight := false
	for _, change := range changes {
		index := (change.Y-e.minY)*256 + change.Z*16 + change.X
		// The original table includes emission, opacity and all six directional
		// occlusion shapes. Equal rules cannot change either light channel;
		// still retain the new block state for later edits and neighboring solves.
		relight = relight || e.rules.states[c.states[index]] != e.rules.states[change.State]
		c.states[index] = change.State
	}
	if !relight {
		return nil, nil
	}
	for dx := int32(-1); dx <= 1; dx++ {
		for dz := int32(-1); dz <= 1; dz++ {
			if neighbor := e.columns[Position{p.X + dx, p.Z + dz}]; neighbor != nil {
				neighbor.trusted = false
			}
		}
	}
	return e.relight(p), nil
}

func lightEqual(a, b *Data) bool {
	if a == nil || a.Complete != b.Complete {
		return false
	}
	for i := range b.Sky {
		if !bytes.Equal(a.Sky[i], b.Sky[i]) || !bytes.Equal(a.Block[i], b.Block[i]) {
			return false
		}
	}
	return true
}

func packLayer(values []byte, width, layer, x0, z0 int) []byte {
	var packed [2048]byte
	nonzero := false
	for y := 0; y < 16; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				v := values[(layer*16+y)*width*width+(z0+z)*width+x0+x]
				i := y*256 + z*16 + x
				packed[i/2] |= v << uint((i&1)*4)
				nonzero = nonzero || v != 0
			}
		}
	}
	if !nonzero {
		return nil
	}
	return append([]byte(nil), packed[:]...)
}

// Changes can influence at most the immediately adjacent chunks: sideways
// propagation loses at least one level and the maximum source level is 15.
// Outside boundary samples remain valid when solving this local 3x3 area.
func (e *Engine) relight(center Position) map[Position]*Data {
	const width = 48
	plane := width * width
	lightHeight := e.height + 32
	minX, minZ := (center.X-1)*16, (center.Z-1)*16
	work := e.work
	if work == nil {
		work = &relightWork{states: make([]uint16, plane*lightHeight), sky: make([]byte, plane*lightHeight), block: make([]byte, plane*lightHeight)}
		e.work = work
	} else {
		clear(work.states)
		clear(work.sky)
		clear(work.block)
		clear(work.loaded[:])
		for i := range work.skyQueue {
			work.skyQueue[i] = work.skyQueue[i][:0]
			work.blockQueue[i] = work.blockQueue[i][:0]
		}
	}
	states, loaded := work.states, work.loaded[:]
	var columns [3][3]*column
	for dx := 0; dx < 3; dx++ {
		for dz := 0; dz < 3; dz++ {
			c := e.columns[Position{center.X + int32(dx) - 1, center.Z + int32(dz) - 1}]
			columns[dx][dz] = c
			if c == nil {
				continue
			}
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					loaded[(dz*16+z)*width+dx*16+x] = true
				}
				for y := 0; y < e.height; y++ {
					copy(states[(y+16)*plane+(dz*16+z)*width+dx*16:][:16], c.states[y*256+z*16:][:16])
				}
			}
		}
	}
	sky, block := work.sky, work.block
	// Bucket queues process brightest nodes first, avoiding repeated visits to
	// cells reached through competing light paths. Entries encode a grid index.
	skyQueue, blockQueue := &work.skyQueue, &work.blockQueue
	if e.hasSky {
		for z := 0; z < width; z++ {
			for x := 0; x < width; x++ {
				if !loaded[z*width+x] {
					continue
				}
				previous := e.rules.states[0]
				for y := lightHeight - 1; y >= 0; y-- {
					index := y*plane + z*width + x
					current := e.rules.states[states[index]]
					if current.Opacity > 0 || e.rules.blocked[previous.Faces[0]][current.Faces[1]] {
						break
					}
					sky[index] = 15
					previous = current
				}
			}
		}
		// Interior direct-sun cells already have their final level. Queue only
		// edges next to a lower level, not hundreds of thousands of sky blocks.
		for y := 0; y < lightHeight; y++ {
			for z := 0; z < width; z++ {
				for x := 0; x < width; x++ {
					i := y*plane + z*width + x
					if sky[i] != 15 {
						continue
					}
					if (y > 0 && sky[i-plane] < 14) || (x > 0 && loaded[z*width+x-1] && sky[i-1] < 14) || (x+1 < width && loaded[z*width+x+1] && sky[i+1] < 14) || (z > 0 && loaded[(z-1)*width+x] && sky[i-width] < 14) || (z+1 < width && loaded[(z+1)*width+x] && sky[i+width] < 14) {
						skyQueue[15] = append(skyQueue[15], i)
					}
				}
			}
		}
	}
	for y := 16; y < lightHeight-16; y++ {
		for z := 0; z < width; z++ {
			for x := 0; x < width; x++ {
				if !loaded[z*width+x] {
					continue
				}
				i := y*plane + z*width + x
				if level := e.rules.states[states[i]].Emission; level > 0 {
					block[i] = level
					blockQueue[level] = append(blockQueue[level], i)
				}
			}
		}
	}
	// Pull unchanged light from loaded neighbors outside the solved area.
	for z := 0; z < width; z++ {
		for x := 0; x < width; x++ {
			if !loaded[z*width+x] || (x != 0 && x != width-1 && z != 0 && z != width-1) {
				continue
			}
			for y := 0; y < lightHeight; y++ {
				for direction, offset := range [6][3]int{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}} {
					nx, nz := x+offset[0], z+offset[2]
					if nx >= 0 && nx < width && nz >= 0 && nz < width {
						continue
					}
					wx, wz := minX+int32(nx), minZ+int32(nz)
					neighbor := e.columns[Position{wx >> 4, wz >> 4}]
					if neighbor == nil || neighbor.light == nil {
						continue
					}
					i := y*plane + z*width + x
					state := uint16(0)
					if y >= 16 && y < lightHeight-16 {
						state = neighbor.states[(y-16)*256+int(wz&15)*16+int(wx&15)]
					}
					if e.rules.blocked[e.rules.states[state].Faces[direction^1]][e.rules.states[states[i]].Faces[direction]] {
						continue
					}
					attenuation := max(byte(1), e.rules.states[states[i]].Opacity)
					for kind := 0; kind < 2; kind++ {
						layers, values, queue := neighbor.light.Sky, sky, skyQueue
						if kind == 1 {
							layers, values, queue = neighbor.light.Block, block, blockQueue
						}
						if kind == 0 && !e.hasSky {
							continue
						}
						layer := layers[y>>4]
						if layer == nil {
							continue
						}
						ni := (y&15)*256 + int(wz&15)*16 + int(wx&15)
						level := (layer[ni/2] >> uint((ni&1)*4)) & 15
						if level > attenuation && level-attenuation > values[i] {
							values[i] = level - attenuation
							queue[values[i]] = append(queue[values[i]], i)
						}
					}
				}
			}
		}
	}
	propagate := func(values []byte, queue *[16][]int) {
		for level := 15; level > 1; level-- {
			for _, i := range queue[level] {
				if int(values[i]) != level {
					continue
				}
				y, flat := i/plane, i%plane
				z, x := flat/width, flat%width
				source := e.rules.states[states[i]]
				for direction, offset := range [6][3]int{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}} {
					nx, ny, nz := x+offset[0], y+offset[1], z+offset[2]
					if nx < 0 || nx >= width || ny < 0 || ny >= lightHeight || nz < 0 || nz >= width || !loaded[nz*width+nx] {
						continue
					}
					ni := ny*plane + nz*width + nx
					target := e.rules.states[states[ni]]
					incoming := level - max(1, int(target.Opacity))
					if incoming <= int(values[ni]) || e.rules.blocked[source.Faces[direction]][target.Faces[direction^1]] {
						continue
					}
					values[ni] = byte(incoming)
					if incoming > 1 {
						queue[incoming] = append(queue[incoming], ni)
					}
				}
			}
		}
	}
	if e.hasSky {
		propagate(sky, skyQueue)
	}
	propagate(block, blockQueue)
	updates := make(map[Position]*Data)
	e.revision++
	for dx := 0; dx < 3; dx++ {
		for dz := 0; dz < 3; dz++ {
			c := columns[dx][dz]
			if c == nil {
				continue
			}
			p := Position{center.X + int32(dx) - 1, center.Z + int32(dz) - 1}
			data := &Data{Sky: make([][]byte, lightHeight/16), Block: make([][]byte, lightHeight/16), Complete: true, Revision: e.revision}
			for nx := int32(-1); nx <= 1; nx++ {
				for nz := int32(-1); nz <= 1; nz++ {
					if e.columns[Position{p.X + nx, p.Z + nz}] == nil {
						data.Complete = false
					}
				}
			}
			// A valid Java save already contains lighting for unloaded neighbors.
			// Preserve it until we have its full neighborhood, or an edit makes
			// it necessary to recompute the affected area.
			if c.trusted && !data.Complete {
				continue
			}
			for layer := range data.Sky {
				data.Sky[layer] = packLayer(sky, width, layer, dx*16, dz*16)
				data.Block[layer] = packLayer(block, width, layer, dx*16, dz*16)
			}
			if !lightEqual(c.light, data) {
				c.light = data
				updates[p] = data
			}
		}
	}
	return updates
}
