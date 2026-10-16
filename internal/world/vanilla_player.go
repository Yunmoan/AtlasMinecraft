package world

import (
	"atlas/internal/item"
	"atlas/internal/nbt"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (v *VanillaSave) playerPath(uuid [16]byte) string {
	return filepath.Join(v.dir, "playerdata", fmt.Sprintf("%x-%x-%x-%x-%x.dat",
		uuid[:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]))
}

func (v *VanillaSave) loadPlayer(uuid [16]byte) (PlayerPosition, bool, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	data, err := readGzip(v.playerPath(uuid))
	if errors.Is(err, os.ErrNotExist) {
		return PlayerPosition{}, false, nil
	}
	if err != nil {
		return PlayerPosition{}, false, err
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		return PlayerPosition{}, false, err
	}
	o, _ := root.Object()
	if version, ok := o.Number("DataVersion"); !ok || version != 3955 {
		return PlayerPosition{}, false, fmt.Errorf("unsupported player DataVersion %d", version)
	}
	dimensionName, _ := o.Text("Dimension")
	dimension, err := playerDimension(dimensionName)
	if err != nil {
		return PlayerPosition{}, false, err
	}
	pos, ok := o["Pos"].Elements()
	if !ok || len(pos) != 3 || pos[0].Type != nbt.Double || pos[1].Type != nbt.Double || pos[2].Type != nbt.Double {
		return PlayerPosition{}, false, errors.New("invalid player Pos list")
	}
	rotation, ok := o["Rotation"].Elements()
	if !ok || len(rotation) != 2 || rotation[0].Type != nbt.Float || rotation[1].Type != nbt.Float {
		return PlayerPosition{}, false, errors.New("invalid player Rotation list")
	}
	p := PlayerPosition{
		X: pos[0].Value.(float64), Y: pos[1].Value.(float64), Z: pos[2].Value.(float64),
		Yaw: rotation[0].Value.(float32), Pitch: rotation[1].Value.(float32),
		Dimension: dimension.name,
	}
	if abilities, ok := o.Object("abilities"); ok {
		flying, _ := abilities.Number("flying")
		p.Flying = flying != 0
	}
	if cooldown, ok := o.Number("PortalCooldown"); ok {
		if cooldown < 0 || cooldown > 2147483647 {
			return PlayerPosition{}, false, errors.New("invalid player portal cooldown")
		}
		p.PortalCooldown = int32(cooldown)
	}
	if !validPlayerPosition(p) {
		return PlayerPosition{}, false, errors.New("invalid vanilla player position")
	}
	return p, true, nil
}

func (v *VanillaSave) savePlayer(uuid [16]byte, p PlayerPosition) error {
	return v.savePlayerState(uuid, p, nil)
}

// setPlayerSpawn writes the vanilla personal respawn point tags into the
// player's .dat, preserving any other stored state. /spawnpoint forces the
// point (SpawnForced=1), as vanilla's SetSpawnCommand does.
func (v *VanillaSave) setPlayerSpawn(uuid [16]byte, x, y, z int32, angle float32) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	path := v.playerPath(uuid)
	name := ""
	root := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"DataVersion": {Type: nbt.Int, Value: int32(3955)},
		"Dimension":   {Type: nbt.String, Value: v.dimension.name},
	}}
	if data, err := readGzip(path); err == nil {
		name, root, err = nbt.Parse(data)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	o, _ := root.Object()
	if version, ok := o.Number("DataVersion"); !ok || version != 3955 {
		return fmt.Errorf("refusing to overwrite player DataVersion %d", version)
	}
	if dimension, _ := o.Text("Dimension"); dimension != "" {
		if _, err := playerDimension(dimension); err != nil {
			return err
		}
	}
	o["SpawnX"] = nbt.Tag{Type: nbt.Int, Value: x}
	o["SpawnY"] = nbt.Tag{Type: nbt.Int, Value: y}
	o["SpawnZ"] = nbt.Tag{Type: nbt.Int, Value: z}
	o["SpawnAngle"] = nbt.Tag{Type: nbt.Float, Value: angle}
	o["SpawnForced"] = nbt.Tag{Type: nbt.Byte, Value: int8(1)}
	o["SpawnDimension"] = nbt.Tag{Type: nbt.String, Value: v.dimension.name}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return writeGzipNBTAtomic(path, name, root)
}

func (v *VanillaSave) loadInventory(uuid [16]byte) (item.Inventory, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	data, err := readGzip(v.playerPath(uuid))
	if errors.Is(err, os.ErrNotExist) {
		return item.Inventory{}, nil
	}
	if err != nil {
		return item.Inventory{}, err
	}
	_, root, err := nbt.Parse(data)
	if err != nil {
		return item.Inventory{}, err
	}
	o, _ := root.Object()
	return item.InventoryFromNBT(o)
}

func (v *VanillaSave) savePlayerState(uuid [16]byte, p PlayerPosition, inventory *item.Inventory) error {
	destination, err := playerDimension(p.Dimension)
	if err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	path := v.playerPath(uuid)
	name := ""
	root := nbt.Tag{Type: nbt.Compound, Value: nbt.Object{
		"DataVersion": {Type: nbt.Int, Value: int32(3955)},
		"Dimension":   {Type: nbt.String, Value: "minecraft:overworld"},
	}}
	if data, err := readGzip(path); err == nil {
		name, root, err = nbt.Parse(data)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	o, _ := root.Object()
	if version, ok := o.Number("DataVersion"); !ok || version != 3955 {
		return fmt.Errorf("refusing to overwrite player DataVersion %d", version)
	}
	if dimension, _ := o.Text("Dimension"); dimension != "" {
		if _, err := playerDimension(dimension); err != nil {
			return err
		}
	}
	o["Dimension"] = nbt.Tag{Type: nbt.String, Value: destination.name}
	o["PortalCooldown"] = nbt.Tag{Type: nbt.Int, Value: p.PortalCooldown}
	o["Pos"] = nbt.Tag{Type: nbt.List, ListType: nbt.Double, Value: []nbt.Tag{
		{Type: nbt.Double, Value: p.X}, {Type: nbt.Double, Value: p.Y}, {Type: nbt.Double, Value: p.Z},
	}}
	o["Rotation"] = nbt.Tag{Type: nbt.List, ListType: nbt.Float, Value: []nbt.Tag{
		{Type: nbt.Float, Value: p.Yaw}, {Type: nbt.Float, Value: p.Pitch},
	}}
	abilities, ok := o.Object("abilities")
	if !ok {
		abilities = nbt.Object{}
		o["abilities"] = nbt.Tag{Type: nbt.Compound, Value: abilities}
	}
	flying := int8(0)
	if p.Flying {
		flying = 1
	}
	abilities["flying"] = nbt.Tag{Type: nbt.Byte, Value: flying}
	if inventory != nil {
		if err := inventory.SaveNBT(o); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return writeGzipNBTAtomic(path, name, root)
}
