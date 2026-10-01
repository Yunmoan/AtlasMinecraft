package server

import (
	"atlas/internal/player"
	"atlas/internal/proto"
	"atlas/internal/registry"
	"crypto/md5"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// handleLogin drives the login phase through Login Success and the
// Login Acknowledged handshake, then hands off to configuration.
func (s *Server) handleLogin(fc *proto.FramedConn, protoVer int32) {
	if protoVer != ProtocolVersion {
		s.loginKick(fc, "Atlas is running Minecraft 1.21.1 (protocol 767)")
		return
	}

	fc.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	id, payload, err := fc.ReadPacket()
	if err != nil || id != proto.CLoginStart {
		return
	}
	r := proto.NewReader(payload)
	name, err := r.String(16)
	if err != nil || !validOperatorName(name) {
		return
	}
	clientUUID, err := r.UUID() // untrusted; session server supplies online UUID
	if err != nil {
		return
	}
	_ = clientUUID
	uuid := offlineUUID(name)
	var properties []player.ProfileProperty

	if s.Hub.Count() >= s.Cfg.MaxPlayers {
		s.loginKick(fc, "Server is full")
		return
	}
	if s.Cfg.OnlineMode {
		profile, err := s.auth.authenticate(fc, name)
		if err != nil {
			slog.Warn("online authentication failed", "player", name, "error", err)
			if errors.Is(err, errInvalidSession) {
				s.loginKick(fc, "Failed to verify username!")
			} else {
				s.loginKick(fc, "Authentication failed. Please try again later.")
			}
			return
		}
		name, uuid, properties = profile.Name, profile.UUID, profile.Properties
		s.Logf("login verified: %s (UUID %s)", name, formatUUID(uuid))
	} else {
		s.Logf("login start: %s (offline UUID %s)", name, formatUUID(uuid))
	}
	s.Dbg("login: compression threshold 256 -> success -> awaiting acknowledge")

	// Compression must be enabled before Login Success.
	if err := fc.WritePacket(proto.SLoginSetCompress,
		proto.NewWriter().VarInt(256).Payload()); err != nil {
		return
	}
	fc.EnableCompression(256)
	// 1.20.5+: UUID, username, authenticated profile properties, and
	// strictErrorHandling (false = lenient about unexpected packets).
	success := proto.NewWriter().UUID(uuid).String(name).VarInt(int32(len(properties)))
	for _, property := range properties {
		success.String(property.Name).String(property.Value).Bool(property.Signature != "")
		if property.Signature != "" {
			success.String(property.Signature)
		}
	}
	if err := fc.WritePacket(proto.SLoginSuccess, success.Bool(false).Payload()); err != nil {
		return
	}

	// The client must acknowledge before configuration starts.
	for {
		fc.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		id, _, err := fc.ReadPacket()
		if err != nil {
			return
		}
		if id == proto.CLoginAcknowledged {
			s.handleConfiguration(fc, name, uuid, properties)
			return
		}
		// cookie/plugin responses: we never request any, ignore.
	}
}

// offlineUUID derives the vanilla offline-mode UUID (MD5 of "OfflinePlayer:name").
func offlineUUID(name string) [16]byte {
	h := md5.Sum([]byte("OfflinePlayer:" + name))
	h[6] = (h[6] & 0x0F) | 0x30 // version 3
	h[8] = (h[8] & 0x3F) | 0x80 // IETF variant
	return h
}

// handleConfiguration sends the 1.20.2+ configuration payload and waits for
// the client's Finish Configuration acknowledgement.
func (s *Server) handleConfiguration(fc *proto.FramedConn, name string, uuid [16]byte, properties []player.ProfileProperty) {
	// Brand (plugin message; the payload itself is a length-prefixed string).
	brand := proto.NewWriter().String("minecraft:brand").
		Raw(proto.NewWriter().String(Brand).Payload()).
		Payload()
	fc.WritePacket(proto.SCfgCustomPayload, brand)

	// Feature flags + known packs + the full tag set: vanilla payloads.
	fc.WritePacket(proto.SCfgFeatureFlags, registry.FeatureFlags())
	fc.WritePacket(proto.SCfgSelectKnownPacks, registry.KnownPacks())

	// Registry data: since 1.20.5 the client wants ONE packet per registry,
	// each carrying the registry identifier and its ordered entry list.
	for _, payload := range registry.Packets() {
		if err := fc.WritePacket(proto.SCfgRegistryData, payload); err != nil {
			return
		}
	}

	// Tags reference registry entries and must always be sent in full.
	fc.WritePacket(proto.SCfgTags, registry.Tags())

	// Finish configuration.
	fc.WritePacket(proto.SCfgFinish, nil)
	s.Dbg("config: brand/feature_flags/known_packs/11 registries(%d bytes)/tags sent, finish",
		len(registry.Packets()[0])+len(registry.Tags()))

	viewDist := int32(s.Cfg.ViewDistance)
	skinParts := byte(0x7f)
	for {
		fc.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		id, payload, err := fc.ReadPacket()
		if err != nil {
			return
		}
		switch id {
		case proto.CCfgClientSettings:
			vd, parts, err := clientSettings(payload)
			if err != nil {
				return
			}
			viewDist, skinParts = vd, parts
		case proto.CCfgFinishAck:
			s.Dbg("config: client acknowledged finish (view distance %d)", viewDist)
			s.handlePlay(fc, name, uuid, viewDist, skinParts, properties)
			return
		default:
			// keep-alive echo, pong, brand, select_known_packs: ignorable
			s.Dbg("config: received packet 0x%02X (%d bytes)", id, len(payload))
		}
	}
}

// Client Information begins with locale, view distance, chat mode, chat
// colors, then the displayed skin-parts bit mask. Later fields are not needed
// for these two settings and may evolve without shifting the skin byte.
func clientSettings(payload []byte) (int32, byte, error) {
	r := proto.NewReader(payload)
	if _, err := r.String(16); err != nil {
		return 0, 0, err
	}
	view, err := r.U8()
	if err != nil {
		return 0, 0, err
	}
	if _, err := r.VarInt(); err != nil {
		return 0, 0, err
	}
	if _, err := r.Bool(); err != nil {
		return 0, 0, err
	}
	parts, err := r.U8()
	if err != nil {
		return 0, 0, fmt.Errorf("client skin parts: %w", err)
	}
	return int32(view), parts & 0x7f, nil
}
