// Package server implements the connection lifecycle: handshake, status
// pings, offline-mode login, the 1.20.2+ configuration phase, and the play
// phase. Each connection runs its own read goroutine; writes are serialized
// through the player's writer goroutine.
package server

import (
	"atlas/internal/hub"
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/world"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ProtocolVersion is Minecraft Java 1.21.1.
const ProtocolVersion = 767

// Brand is reported to the client via the minecraft:brand plugin channel.
const Brand = "Atlas"

type Config struct {
	Addr               string
	Motd               string
	MaxPlayers         int
	OnlineMode         bool
	ViewDistance       int
	SimulationDistance int
	Debug              bool
	TracePackets       bool
	Ops                map[string]bool
	Operators          *OperatorStore
}

type Server struct {
	Cfg            Config
	Hub            *hub.Hub
	World          *world.World
	nextEID        atomic.Int32
	nextTeleportID atomic.Int32
	listenerMu     sync.Mutex
	listener       net.Listener
	connections    sync.Map
	clients        sync.WaitGroup
	closed         bool
	entityMu       sync.Mutex // position publication, spawn/despawn and observer movement ordering
	auth           onlineAuthentication
	performance    *performanceMetrics
}

func New(cfg Config, h *hub.Hub, w *world.World) *Server {
	if cfg.ViewDistance == 0 {
		cfg.ViewDistance = 10
	}
	if cfg.SimulationDistance == 0 {
		cfg.SimulationDistance = 10
	}
	return &Server{Cfg: cfg, Hub: h, World: w, performance: newPerformanceMetrics(time.Now())}
}

func (s *Server) Logf(format string, args ...any) {
	slog.Info(fmt.Sprintf(format, args...), "component", "server")
}

// Dbg logs protocol-level detail when -debug is on.
func (s *Server) Dbg(format string, args ...any) {
	if s.Cfg.Debug {
		slog.Debug(fmt.Sprintf(format, args...), "component", "protocol")
	}
}

func (s *Server) ListenAndServe() error {
	if s.Cfg.OnlineMode {
		if err := s.auth.ensureKey(); err != nil {
			return fmt.Errorf("generate online-mode keypair: %w", err)
		}
		s.Logf("Online mode enabled: authenticating players with Mojang")
	}
	ln, err := net.Listen("tcp", s.Cfg.Addr)
	if err != nil {
		return err
	}
	s.listenerMu.Lock()
	if s.closed {
		s.listenerMu.Unlock()
		ln.Close()
		return nil
	}
	s.listener = ln
	s.listenerMu.Unlock()
	s.Logf("listening on %s (protocol %d, online mode %t)", s.Cfg.Addr, ProtocolVersion, s.Cfg.OnlineMode)
	for {
		nc, err := ln.Accept()
		if err != nil {
			if s.isClosed() {
				return nil
			}
			return err
		}
		s.listenerMu.Lock()
		if s.closed {
			s.listenerMu.Unlock()
			nc.Close()
			return nil
		}
		s.connections.Store(nc, true)
		s.clients.Add(1)
		s.listenerMu.Unlock()
		go func() {
			defer s.clients.Done()
			defer s.connections.Delete(nc)
			s.handle(nc)
		}()
	}
}

func (s *Server) isClosed() bool {
	s.listenerMu.Lock()
	defer s.listenerMu.Unlock()
	return s.closed
}

// Close stops accepting clients and waits for their handlers to finish.
func (s *Server) Close() {
	s.listenerMu.Lock()
	s.closed = true
	if s.listener != nil {
		s.listener.Close()
	}
	s.connections.Range(func(k, _ any) bool {
		k.(net.Conn).Close()
		return true
	})
	s.listenerMu.Unlock()
	s.clients.Wait()
}

func (s *Server) savePlayer(p *player.Player) {
	p.SaveMu.Lock()
	defer p.SaveMu.Unlock()
	x, y, z, yaw, pitch, _ := p.Pos()
	position := world.PlayerPosition{X: x, Y: y, Z: z, Yaw: yaw, Pitch: pitch, Flying: p.Flying()}
	inventory, _ := p.Inventory()
	if err := s.World.SavePlayerState(p.UUID, position, inventory); err != nil {
		slog.Error("player save failed", "player", p.Name, "error", err)
	}
}

// SavePlayers writes online positions and inventories during periodic saves.
func (s *Server) SavePlayers() {
	for _, p := range s.Hub.Snapshot() {
		s.savePlayer(p)
	}
}

// handle runs the initial state machine for one connection.
func (s *Server) handle(nc net.Conn) {
	defer nc.Close()
	fc := proto.NewFramedConn(nc)
	nc.SetReadDeadline(time.Now().Add(30 * time.Second))

	id, payload, err := fc.ReadPacket()
	if err != nil {
		return
	}
	if id != proto.CHandshake {
		return
	}
	r := proto.NewReader(payload)
	protoVer, err := r.VarInt()
	if err != nil {
		return
	}
	host, err := r.String(255)
	if err != nil {
		return
	}
	_ = host
	if _, err = r.U16(); err != nil { // server port
		return
	}
	nextState, err := r.VarInt()
	if err != nil {
		return
	}
	// Protocol >= 1.20.5 appends a client-intent varint here; we do not
	// support transfers, so any login intent is treated as a normal join.

	switch nextState {
	case 1:
		s.handleStatus(fc)
	case 2:
		s.handleLogin(fc, protoVer)
	}
}

// handleStatus answers the server-list ping until the client disconnects.
func (s *Server) handleStatus(fc *proto.FramedConn) {
	fc.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	for {
		id, payload, err := fc.ReadPacket()
		if err != nil {
			return
		}
		switch id {
		case proto.CStatusRequest:
			if err := fc.WritePacket(proto.SStatusResponse,
				proto.NewWriter().String(s.statusJSON()).Payload()); err != nil {
				return
			}
		case proto.CStatusPing:
			v, err := proto.NewReader(payload).I64()
			if err != nil {
				return
			}
			fc.WritePacket(proto.SStatusPong, proto.NewWriter().I64(v).Payload())
			return
		default:
			return
		}
	}
}

func (s *Server) statusJSON() string {
	motd := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", "\\n").Replace(s.Cfg.Motd)
	return fmt.Sprintf(`{"version":{"name":"1.21.1","protocol":%d},"players":{"max":%d,"online":%d},`+
		`"description":{"text":"%s"},"enforcesSecureChat":false}`,
		ProtocolVersion, s.Cfg.MaxPlayers, s.Hub.Count(), motd)
}

// chatComponent encodes a bare string as a network NBT text component —
// the valid minimal form since 1.20.3. The root is anonymous (no name).
func chatComponent(text string) []byte {
	b := []byte(text)
	w := proto.NewWriter()
	w.U8(0x08) // TAG_String root, no name bytes
	w.U8(byte(len(b) >> 8)).U8(byte(len(b)))
	w.Raw(b)
	return w.Payload()
}

// systemChat builds a System Chat packet payload (component + overlay flag).
func systemChat(text string) []byte {
	return append(chatComponent(text), 0x00) // isActionBar = false
}

// loginKick sends a login-phase disconnect. Unlike configuration/play,
// the login disconnect reason is a JSON text STRING (not NBT).
func (s *Server) loginKick(fc *proto.FramedConn, msg string) {
	esc := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(msg)
	fc.WritePacket(proto.SLoginDisconnect,
		proto.NewWriter().String(`{"text":"`+esc+`"}`).Payload())
}

// kick sends a disconnect packet in whatever phase the connection is in.
func (s *Server) kick(fc *proto.FramedConn, packetID int32, msg string) {
	fc.WritePacket(packetID, chatComponent(msg))
}

// nextEntityID allocates a fresh play-phase entity id.
func (s *Server) nextEntityID() int32 { return s.nextEID.Add(1) }
