package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed motion_blocking_data.json
var motionBlockingData []byte

var motionBlockingOnce sync.Once
var motionBlockingKinds map[string]heightmapBlockKind
var motionBlockingError error

type heightmapBlockKind struct {
	Motion     string `json:"motion"`
	OceanFloor bool   `json:"ocean_floor"`
}

func heightmapKind(state SurfaceBlockState) (heightmapBlockKind, error) {
	motionBlockingOnce.Do(func() {
		motionBlockingError = json.Unmarshal(motionBlockingData, &motionBlockingKinds)
		if motionBlockingError == nil && len(motionBlockingKinds) != 1060 {
			motionBlockingError = fmt.Errorf("expected 1060 Minecraft 1.21.1 block classifiers, got %d", len(motionBlockingKinds))
		}
	})
	if motionBlockingError != nil {
		return heightmapBlockKind{}, motionBlockingError
	}
	kind, found := motionBlockingKinds[state.Name]
	if !found {
		return heightmapBlockKind{}, fmt.Errorf("unknown Minecraft 1.21.1 heightmap block %q", state.Name)
	}
	return kind, nil
}

func motionBlocking(state SurfaceBlockState) (bool, error) {
	kind, err := heightmapKind(state)
	if err != nil {
		return false, err
	}
	switch kind.Motion {
	case "always":
		return true, nil
	case "never":
		return false, nil
	case "waterlogged_default_true", "waterlogged_default_false":
		waterlogged, provided := state.Properties["waterlogged"]
		if !provided {
			return kind.Motion == "waterlogged_default_true", nil
		}
		if waterlogged != "true" && waterlogged != "false" {
			return false, fmt.Errorf("invalid waterlogged property %q for %s", waterlogged, state.Name)
		}
		return waterlogged == "true", nil
	default:
		return false, fmt.Errorf("invalid Minecraft 1.21.1 motion blocking kind %q", kind.Motion)
	}
}

func oceanFloorSolid(state SurfaceBlockState) (bool, error) {
	kind, err := heightmapKind(state)
	if err != nil {
		return false, err
	}
	return kind.OceanFloor, nil
}
