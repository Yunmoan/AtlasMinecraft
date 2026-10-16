package world

import (
	"atlas/internal/item"
	"errors"
	"math"
)

// Store owns a Java Edition save. Chunk and player edits are written to
// vanilla Anvil/NBT files rather than an Atlas-specific overlay.
type Store struct {
	vanilla *VanillaSave
}

func OpenStore(dir string) (*Store, error) {
	vanilla, err := openVanilla(dir)
	if err != nil {
		return nil, err
	}
	return &Store{vanilla: vanilla}, nil
}

type worldMeta struct {
	Age     int64
	DayTime int64
	Weather string
}

// PlayerPosition remains separate from inventory so position-only tools can
// update it without overwriting existing player items.
type PlayerPosition struct {
	X, Y, Z        float64
	Yaw, Pitch     float32
	Flying         bool
	Dimension      string
	PortalCooldown int32
}

func validPlayerPosition(p PlayerPosition) bool {
	if _, err := playerDimension(p.Dimension); err != nil {
		return false
	}
	return !math.IsNaN(p.X) && !math.IsNaN(p.Y) && !math.IsNaN(p.Z) &&
		!math.IsInf(p.X, 0) && !math.IsInf(p.Y, 0) && !math.IsInf(p.Z, 0) &&
		math.Abs(p.X) <= 30000000 && math.Abs(p.Z) <= 30000000 &&
		math.Abs(p.Y) <= 30000000 &&
		!math.IsNaN(float64(p.Yaw)) && !math.IsNaN(float64(p.Pitch)) &&
		!math.IsInf(float64(p.Yaw), 0) && !math.IsInf(float64(p.Pitch), 0)
}

func playerDimension(name string) (dimensionSettings, error) {
	if name == "" {
		name = "minecraft:overworld"
	}
	return builtinDimension(name)
}

func (s *Store) LoadPlayer(uuid [16]byte) (PlayerPosition, bool, error) {
	return s.vanilla.loadPlayer(uuid)
}

func (s *Store) SavePlayer(uuid [16]byte, position PlayerPosition) error {
	if !validPlayerPosition(position) {
		return errors.New("invalid player position")
	}
	return s.vanilla.savePlayer(uuid, position)
}

func (s *Store) SavePlayerSpawn(uuid [16]byte, x, y, z int32, angle float32) error {
	return s.vanilla.setPlayerSpawn(uuid, x, y, z, angle)
}

func (s *Store) LoadInventory(uuid [16]byte) (item.Inventory, error) {
	return s.vanilla.loadInventory(uuid)
}
func (s *Store) SavePlayerState(uuid [16]byte, position PlayerPosition, inventory item.Inventory) error {
	if !validPlayerPosition(position) || inventory.Selected < 0 || inventory.Selected > 8 {
		return errors.New("invalid player state")
	}
	return s.vanilla.savePlayerState(uuid, position, &inventory)
}

func (s *Store) LoadMeta() (worldMeta, error)  { return s.vanilla.loadMeta() }
func (s *Store) SaveMeta(meta worldMeta) error { return s.vanilla.saveMeta(meta) }
func (s *Store) Close() error                  { return nil }
