package server

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// OperatorStore reads and writes the vanilla ops.json entry shape.
type OperatorStore struct {
	mu      sync.RWMutex
	path    string
	entries map[string]OperatorEntry
}

type OperatorEntry struct {
	UUID                string `json:"uuid"`
	Name                string `json:"name"`
	Level               int    `json:"level"`
	BypassesPlayerLimit bool   `json:"bypassesPlayerLimit"`
}

func OpenOperators(path string) (*OperatorStore, error) {
	s := &OperatorStore{path: path, entries: make(map[string]OperatorEntry)}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, s.save()
	}
	if err != nil {
		return nil, err
	}
	var entries []OperatorEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	for _, entry := range entries {
		if !validOperatorName(entry.Name) || entry.Level < 1 || entry.Level > 4 || !validUUID(entry.UUID) {
			return nil, fmt.Errorf("invalid operator entry for %q", entry.Name)
		}
		s.entries[strings.ToLower(entry.Name)] = entry
	}
	return s, nil
}

func validUUID(s string) bool {
	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	_, err := hex.DecodeString(strings.ReplaceAll(s, "-", ""))
	return err == nil
}

func validOperatorName(name string) bool {
	if len(name) < 1 || len(name) > 16 {
		return false
	}
	for _, r := range name {
		if r != '_' && (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

func (s *OperatorStore) Has(name string) bool {
	return s.Level(name) > 0
}

func (s *OperatorStore) Level(name string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[strings.ToLower(name)].Level
}

func (s *OperatorStore) LevelForPlayer(name string, uuid [16]byte) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[strings.ToLower(name)]
	if !ok || !strings.EqualFold(entry.UUID, formatUUID(uuid)) {
		return 0
	}
	return entry.Level
}

func (s *OperatorStore) Add(name string) (bool, error) {
	return s.AddProfile(name, offlineUUID(name))
}

// AddProfile persists the verified UUID in online mode. This also upgrades
// an existing entry for the same name that held an offline UUID.
func (s *OperatorStore) AddProfile(name string, uuid [16]byte) (bool, error) {
	if !validOperatorName(name) {
		return false, errors.New("operator name must be a 1–16 character Minecraft username")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.ToLower(name)
	previous, exists := s.entries[key]
	if exists && strings.EqualFold(previous.UUID, formatUUID(uuid)) {
		return false, nil
	}
	s.entries[key] = OperatorEntry{UUID: formatUUID(uuid), Name: name, Level: 4}
	if err := s.save(); err != nil {
		if exists {
			s.entries[key] = previous
		} else {
			delete(s.entries, key)
		}
		return false, err
	}
	return true, nil
}

func (s *OperatorStore) Remove(name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := strings.ToLower(name)
	previous, exists := s.entries[key]
	if !exists {
		return false, nil
	}
	delete(s.entries, key)
	if err := s.save(); err != nil {
		s.entries[key] = previous
		return false, err
	}
	return true, nil
}

func (s *OperatorStore) save() error {
	entries := make([]OperatorEntry, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name) })
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), "atlas-ops-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(append(data, '\n')); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

func formatUUID(uuid [16]byte) string {
	var b [36]byte
	hex.Encode(b[0:8], uuid[0:4])
	b[8] = '-'
	hex.Encode(b[9:13], uuid[4:6])
	b[13] = '-'
	hex.Encode(b[14:18], uuid[6:8])
	b[18] = '-'
	hex.Encode(b[19:23], uuid[8:10])
	b[23] = '-'
	hex.Encode(b[24:36], uuid[10:16])
	return string(b[:])
}
