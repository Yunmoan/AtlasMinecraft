package worldgen

import (
	"atlas/internal/blockstate"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
)

//go:embed portal_rules_1.21.1.txt
var portalRuleText string

var portalRuleFlags = func() []byte {
	f := strings.Fields(portalRuleText)
	if len(f) != 3 || f[0] != "portal_block_rules" || f[1] != "26684" {
		panic("invalid Nether portal block rules")
	}
	flags, err := base64.StdEncoding.DecodeString(f[2])
	if err != nil || len(flags) != (26684+3)/4 {
		panic("invalid Nether portal rule flags")
	}
	return flags
}()

type PortalRectangle struct {
	Origin        FeaturePosition
	Axis          string
	Width, Height int32
}

type PortalForcerWorld interface {
	FeatureBlockWorld
	HeightmapAt(heightmap string, x, z int32) (int32, error)
}

func portalRules(s SurfaceBlockState) (replace, solid bool, err error) {
	id, err := blockstate.BlockStateID(s.Name, s.Properties)
	if err != nil {
		return false, false, err
	}
	flags := portalRuleFlags[id/4] >> ((id % 4) * 2)
	return flags&1 != 0, flags&2 != 0, nil
}

// portalSpiral reproduces BlockPos.spiralAround(center,radius,EAST,SOUTH),
// including its iterator's movement-before-leg-update ordering.
func portalSpiral(center FeaturePosition, radius int32) []FeaturePosition {
	p := treeOffset(center, 0, 0, 1)
	directions := [][2]int32{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	leg, length, step := int32(-1), int32(0), int32(0)
	positions := make([]FeaturePosition, 0, (radius*2+1)*(radius*2+1))
	for {
		d := directions[(leg+4)%4]
		p = treeOffset(p, d[0], 0, d[1])
		if step >= length {
			if leg >= 4*radius {
				return positions
			}
			leg++
			step = 0
			length = leg/2 + 1
		}
		step++
		positions = append(positions, p)
	}
}

func portalDistance(a, b FeaturePosition) int64 {
	dx, dy, dz := int64(a.X)-int64(b.X), int64(a.Y)-int64(b.Y), int64(a.Z)-int64(b.Z)
	return dx*dx + dy*dy + dz*dz
}

// FindClosestNetherPortal applies PortalForcer's square POI radius, border,
// current portal-state predicate and distance/Y ordering to a supplied index.
func FindClosestNetherPortal(world FeatureBlockWorld, candidates []FeaturePosition, target FeaturePosition, nether bool, border PortalBorder) (FeaturePosition, bool, error) {
	radius := int32(128)
	if nether {
		radius = 16
	}
	var best FeaturePosition
	found := false
	for _, p := range candidates {
		if p.X < target.X-radius || p.X > target.X+radius || p.Z < target.Z-radius || p.Z > target.Z+radius || !border.Contains(p) {
			continue
		}
		s, err := world.BlockAt(p)
		if err != nil {
			return best, false, err
		}
		if s.Name != "minecraft:nether_portal" || s.Properties["axis"] != "x" && s.Properties["axis"] != "z" {
			continue
		}
		d := portalDistance(p, target)
		if !found || d < portalDistance(best, target) || d == portalDistance(best, target) && p.Y < best.Y {
			best = p
			found = true
		}
	}
	return best, found, nil
}

func CreateNetherPortal(world PortalForcerWorld, target FeaturePosition, axis string, minY, maxY, logicalHeight int32, border PortalBorder) (PortalRectangle, bool, error) {
	rect := PortalRectangle{Axis: axis, Width: 2, Height: 3}
	dx, dz := int32(1), int32(0)
	if axis == "z" {
		dx, dz = 0, 1
	} else if axis != "x" {
		return rect, false, fmt.Errorf("invalid Nether portal axis %q", axis)
	}
	// Direction.getClockWise(): EAST->SOUTH, SOUTH->WEST.
	px, pz := -dz, dx
	ceiling := min(maxY, minY+logicalHeight) - 1
	if logicalHeight <= 0 || maxY <= minY {
		return rect, false, fmt.Errorf("invalid portal dimension bounds")
	}
	replace := func(p FeaturePosition) (bool, error) {
		s, err := world.BlockAt(p)
		if err != nil {
			return false, err
		}
		r, _, err := portalRules(s)
		return r, err
	}
	canHost := func(p FeaturePosition, depth int32) (bool, error) {
		for x := int32(-1); x < 3; x++ {
			for y := int32(-1); y < 4; y++ {
				s, err := world.BlockAt(treeOffset(p, dx*x+px*depth, y, dz*x+pz*depth))
				if err != nil {
					return false, err
				}
				r, solid, err := portalRules(s)
				if err != nil {
					return false, err
				}
				if y < 0 && !solid || y >= 0 && !r {
					return false, nil
				}
			}
		}
		return true, nil
	}
	var best, secondary FeaturePosition
	bestDistance, secondaryDistance := int64(-1), int64(-1)
	for _, column := range portalSpiral(target, 16) {
		top, err := world.HeightmapAt("MOTION_BLOCKING", column.X, column.Z)
		if err != nil {
			return rect, false, err
		}
		if !border.Contains(column) || !border.Contains(treeOffset(column, dx, 0, dz)) {
			continue
		}
		for y := min(ceiling, top); y >= minY; y-- {
			p := column
			p.Y = y
			canReplace, err := replace(p)
			if err != nil {
				return rect, false, err
			}
			if !canReplace {
				continue
			}
			initialY := y
			for y > minY {
				below := p
				below.Y = y - 1
				canReplace, err = replace(below)
				if err != nil {
					return rect, false, err
				}
				if !canReplace {
					break
				}
				y--
			}
			p.Y = y
			if y+4 > ceiling || initialY-y > 0 && initialY-y < 3 {
				continue
			}
			host, err := canHost(p, 0)
			if err != nil {
				return rect, false, err
			}
			if !host {
				continue
			}
			distance := portalDistance(target, p)
			left, err := canHost(p, -1)
			if err != nil {
				return rect, false, err
			}
			right := false
			if left {
				right, err = canHost(p, 1)
				if err != nil {
					return rect, false, err
				}
			}
			if left && right && (bestDistance == -1 || bestDistance > distance) {
				best, bestDistance = p, distance
			}
			if bestDistance == -1 && (secondaryDistance == -1 || secondaryDistance > distance) {
				secondary, secondaryDistance = p, distance
			}
		}
	}
	if bestDistance == -1 && secondaryDistance != -1 {
		best, bestDistance = secondary, secondaryDistance
	}
	if bestDistance == -1 {
		low, high := max(minY+1, int32(70)), ceiling-9
		if high < low {
			return rect, false, nil
		}
		best = treeOffset(target, -dx, 0, -dz)
		best.Y = max(low, min(best.Y, high))
		best = border.Clamp(float64(best.X), float64(best.Y), float64(best.Z))
		for depth := int32(-1); depth < 2; depth++ {
			for x := int32(0); x < 2; x++ {
				for y := int32(-1); y < 3; y++ {
					name := "minecraft:air"
					if y < 0 {
						name = "minecraft:obsidian"
					}
					if err := world.SetBlock(treeOffset(best, x*dx+depth*px, y, x*dz+depth*pz), SurfaceBlockState{Name: name}); err != nil {
						return rect, false, err
					}
				}
			}
		}
	}
	for x := int32(-1); x < 3; x++ {
		for y := int32(-1); y < 4; y++ {
			if x != -1 && x != 2 && y != -1 && y != 3 {
				continue
			}
			if err := world.SetBlock(treeOffset(best, x*dx, y, x*dz), SurfaceBlockState{Name: "minecraft:obsidian"}); err != nil {
				return rect, false, err
			}
		}
	}
	state := SurfaceBlockState{Name: "minecraft:nether_portal", Properties: map[string]string{"axis": axis}}
	for x := int32(0); x < 2; x++ {
		for y := int32(0); y < 3; y++ {
			if err := world.SetBlock(treeOffset(best, x*dx, y, x*dz), state); err != nil {
				return rect, false, err
			}
		}
	}
	rect.Origin = best
	return rect, true, nil
}
