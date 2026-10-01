// Package hub tracks online players and provides coordinate-based lookups
// used by region actors to find packet sinks (viewers) for broadcasts.
package hub

import (
	"atlas/internal/player"
	"atlas/internal/world"
	"sort"
	"strings"
	"sync"
)

type Hub struct {
	mu     sync.RWMutex
	byEID  map[int32]*player.Player
	byUUID map[[16]byte]*player.Player
}

func New() *Hub {
	return &Hub{
		byEID:  make(map[int32]*player.Player),
		byUUID: make(map[[16]byte]*player.Player),
	}
}

// Join registers a player; returns false if their UUID is already online.
func (h *Hub) Join(p *player.Player) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, dup := h.byUUID[p.UUID]; dup {
		return false
	}
	for _, current := range h.byEID {
		if strings.EqualFold(current.Name, p.Name) {
			return false
		}
	}
	h.byEID[p.EID] = p
	h.byUUID[p.UUID] = p
	return true
}

func (h *Hub) Leave(p *player.Player) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.byEID, p.EID)
	delete(h.byUUID, p.UUID)
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byEID)
}

func (h *Hub) FindName(name string) *player.Player {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range h.byEID {
		if strings.EqualFold(p.Name, name) {
			return p
		}
	}
	return nil
}

func (h *Hub) Names() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	names := make([]string, 0, len(h.byEID))
	for _, p := range h.byEID {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return names
}

func (h *Hub) Snapshot() []*player.Player {
	h.mu.RLock()
	defer h.mu.RUnlock()
	players := make([]*player.Player, 0, len(h.byEID))
	for _, p := range h.byEID {
		players = append(players, p)
	}
	return players
}

// Each calls fn for every online player.
func (h *Hub) Each(fn func(*player.Player)) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range h.byEID {
		fn(p)
	}
}

// Others calls fn for every online player except exclude.
func (h *Hub) Others(exclude *player.Player, fn func(*player.Player)) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range h.byEID {
		if p.EID != exclude.EID {
			fn(p)
		}
	}
}

// Broadcast sends a packet to every online player.
func (h *Hub) Broadcast(id int32, payload []byte) {
	h.Each(func(p *player.Player) { p.SendPacket(id, payload) })
}

// ViewersOf returns every player whose view distance covers block (x, z).
// Satisfies world.HubView.
func (h *Hub) ViewersOf(x, z int32) []world.PacketSink {
	cx, cz := x>>4, z>>4
	h.mu.RLock()
	defer h.mu.RUnlock()
	sinks := make([]world.PacketSink, 0, len(h.byEID))
	for _, p := range h.byEID {
		pcx, pcz := p.ChunkPos()
		vd := p.ViewDistance() + 1 // match entity spawn radius
		dx, dz := int64(pcx)-int64(cx), int64(pcz)-int64(cz)
		if dx*dx+dz*dz <= int64(vd)*int64(vd) {
			sinks = append(sinks, p)
		}
	}
	return sinks
}
