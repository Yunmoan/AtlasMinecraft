package world

import (
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"fmt"
)

type portalBlockWorld struct{ world *World }

func (a portalBlockWorld) BlockAt(p worldgen.FeaturePosition) (worldgen.SurfaceBlockState, error) {
	id, err := a.world.BlockAt(p.X, p.Y, p.Z)
	if err != nil {
		return worldgen.SurfaceBlockState{}, err
	}
	state, err := BlockStateByID(id)
	return worldgen.SurfaceBlockState{Name: state.Name, Properties: state.Properties}, err
}
func (a portalBlockWorld) SetBlock(p worldgen.FeaturePosition, state worldgen.SurfaceBlockState) error {
	id, err := BlockStateID(state.Name, state.Properties)
	if err != nil {
		return err
	}
	return a.world.SetBlock(p.X, p.Y, p.Z, id)
}
func (a portalBlockWorld) InWorld(p worldgen.FeaturePosition) bool {
	return p.Y >= a.world.MinBuildHeight() && p.Y < a.world.MaxBuildHeight() && p.X >= -30000000 && p.X < 30000000 && p.Z >= -30000000 && p.Z < 30000000
}

// IgniteNetherPortal activates an empty valid frame in either vanilla axis.
// The caller supplies the ignition position; using the item belongs to server.
func (w *World) IgniteNetherPortal(x, y, z int32) (bool, error) {
	if w.DimensionName() != "minecraft:overworld" && w.DimensionName() != "minecraft:the_nether" {
		return false, nil
	}
	a := portalBlockWorld{w}
	p := worldgen.FeaturePosition{X: x, Y: y, Z: z}
	if !a.InWorld(p) {
		return false, fmt.Errorf("portal ignition outside world bounds")
	}
	shape, found, err := worldgen.FindEmptyNetherPortalShape(a, w.MinBuildHeight(), p, "x")
	if err != nil || !found {
		return false, err
	}
	return true, worldgen.FillNetherPortalShape(a, shape)
}

func (w *World) PortalBorder() (worldgen.PortalBorder, error) {
	centerX, centerZ, size := 0.0, 0.0, 59999968.0
	if w.store != nil && w.store.vanilla != nil {
		v := w.store.vanilla
		v.mu.Lock()
		root, _ := v.root.Object()
		data, _ := root.Object("Data")
		read := func(name string, fallback float64) float64 {
			tag := data[name]
			if tag.Type == nbt.Double {
				if value, ok := tag.Value.(float64); ok {
					return value
				}
			}
			return fallback
		}
		centerX, centerZ, size = read("BorderCenterX", centerX), read("BorderCenterZ", centerZ), read("BorderSize", size)
		v.mu.Unlock()
	}
	return worldgen.NewPortalBorder(centerX, centerZ, size)
}

// NetherPortalSearchTarget uses the destination border and vanilla 8:1 scale.
// This computes the search anchor, not a collision-safe entity arrival.
func (w *World) NetherPortalSearchTarget(sourceDimension string, x, y, z float64) (BlockPosition, error) {
	source, err := builtinDimension(sourceDimension)
	if err != nil {
		return BlockPosition{}, err
	}
	if source.generator == "end" || w.DimensionName() == "minecraft:the_end" {
		return BlockPosition{}, fmt.Errorf("End dimension has no Nether portal transition")
	}
	border, err := w.PortalBorder()
	if err != nil {
		return BlockPosition{}, err
	}
	sourceScale, targetScale := 1.0, 1.0
	if source.generator == "nether" {
		sourceScale = 8
	}
	if w.DimensionName() == "minecraft:the_nether" {
		targetScale = 8
	}
	p, err := worldgen.NetherPortalSearchTarget(border, x, y, z, sourceScale, targetScale)
	return BlockPosition{p.X, p.Y, p.Z}, err
}

// PrepareEndArrival rebuilds the stock End entry platform before a player
// transfer. Runtime drops from removed blocks remain unimplemented.
func (w *World) PrepareEndArrival() (PlayerPosition, error) {
	if w.DimensionName() != "minecraft:the_end" {
		return PlayerPosition{}, fmt.Errorf("End arrival requires the End dimension")
	}
	spawn := worldgen.EndSpawnPoint()
	origin := spawn
	origin.Y--
	if err := worldgen.PlaceEndPlatform(portalBlockWorld{w}, origin); err != nil {
		return PlayerPosition{}, err
	}
	return PlayerPosition{X: float64(spawn.X) + 0.5, Y: float64(spawn.Y) - 1, Z: float64(spawn.Z) + 0.5, Yaw: 90, Dimension: w.DimensionName()}, nil
}
