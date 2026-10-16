package worldgen

import (
	"fmt"
	"math"
)

// PortalBorder uses the actual WorldBorder bounds, including its absolute
// coordinate limit. Portal targets are floored after clamping to max-1.
type PortalBorder struct {
	MinX, MaxX, MinZ, MaxZ float64
}

func NewPortalBorder(centerX, centerZ, size float64) (PortalBorder, error) {
	if math.IsNaN(centerX) || math.IsNaN(centerZ) || math.IsNaN(size) || math.IsInf(centerX, 0) || math.IsInf(centerZ, 0) || math.IsInf(size, 0) || size < 1 {
		return PortalBorder{}, fmt.Errorf("invalid portal world border")
	}
	const limit = 29999984.0
	return PortalBorder{
		MinX: max(centerX-size/2, -limit), MaxX: min(centerX+size/2, limit),
		MinZ: max(centerZ-size/2, -limit), MaxZ: min(centerZ+size/2, limit),
	}, nil
}

func (b PortalBorder) Contains(p FeaturePosition) bool {
	return float64(p.X) >= b.MinX && float64(p.X) < b.MaxX && float64(p.Z) >= b.MinZ && float64(p.Z) < b.MaxZ
}

func (b PortalBorder) Clamp(x, y, z float64) FeaturePosition {
	return FeaturePosition{
		X: int32(math.Floor(javaClamp(x, b.MinX, b.MaxX-1))),
		Y: int32(math.Floor(y)),
		Z: int32(math.Floor(javaClamp(z, b.MinZ, b.MaxZ-1))),
	}
}

func javaClamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	return min(value, high)
}

// NetherPortalSearchTarget is NetherPortalBlock's target before the destination
// portal search. It deliberately retains Y; PortalForcer handles build bounds.
func NetherPortalSearchTarget(border PortalBorder, x, y, z, sourceScale, destinationScale float64) (FeaturePosition, error) {
	for _, v := range []float64{x, y, z, sourceScale, destinationScale} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return FeaturePosition{}, fmt.Errorf("invalid portal target")
		}
	}
	if sourceScale <= 0 || destinationScale <= 0 || math.Abs(y) > 30000000 {
		return FeaturePosition{}, fmt.Errorf("invalid portal target")
	}
	scale := sourceScale / destinationScale
	return border.Clamp(x*scale, y, z*scale), nil
}

func EndSpawnPoint() FeaturePosition { return FeaturePosition{100, 50, 0} }

// PlaceEndPlatform shares EndPlatformFeature's block geometry and iteration
// order. Entity drops on replacement belong to the runtime entity system.
func PlaceEndPlatform(world FeatureBlockWorld, origin FeaturePosition) error {
	for z := int32(-2); z <= 2; z++ {
		for x := int32(-2); x <= 2; x++ {
			for y := int32(-1); y < 3; y++ {
				p := treeOffset(origin, x, y, z)
				name := "minecraft:air"
				if y == -1 {
					name = "minecraft:obsidian"
				}
				state, err := world.BlockAt(p)
				if err != nil {
					return err
				}
				if state.Name != name {
					if err := setFeatureBlock(world, p, SurfaceBlockState{Name: name}); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
