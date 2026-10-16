package worldgen

import "fmt"

// NetherPortalShape preserves PortalShape's bottomLeft and rightDir ordering:
// X frames extend west from Origin; Z frames extend south from Origin.
type NetherPortalShape struct {
	Origin        FeaturePosition
	Axis          string
	Width, Height int32
	PortalBlocks  int32
}

func (s NetherPortalShape) Valid() bool {
	return s.Width >= 2 && s.Width <= 21 && s.Height >= 3 && s.Height <= 21
}
func (s NetherPortalShape) Complete() bool { return s.Valid() && s.PortalBlocks == s.Width*s.Height }

func portalEmpty(state SurfaceBlockState) bool {
	switch state.Name {
	case "minecraft:air", "minecraft:cave_air", "minecraft:void_air", "minecraft:fire", "minecraft:soul_fire", "minecraft:nether_portal":
		return true
	}
	return false
}

func InspectNetherPortalShape(world FeatureBlockWorld, minY int32, origin FeaturePosition, axis string) (NetherPortalShape, error) {
	s := NetherPortalShape{Origin: origin, Axis: axis}
	dx, dz := int32(-1), int32(0)
	if axis == "z" {
		dx, dz = 0, 1
	} else if axis != "x" {
		return s, fmt.Errorf("invalid Nether portal axis %q", axis)
	}
	read := func(p FeaturePosition) (SurfaceBlockState, error) { return world.BlockAt(p) }
	p := origin
	for floor := max(minY, origin.Y-21); p.Y > floor; {
		below, err := read(treeOffset(p, 0, -1, 0))
		if err != nil {
			return s, err
		}
		if !portalEmpty(below) {
			break
		}
		p.Y--
	}
	distance := func(start FeaturePosition, stepX, stepZ int32) (int32, error) {
		for i := int32(0); i <= 21; i++ {
			p := treeOffset(start, stepX*i, 0, stepZ*i)
			state, err := read(p)
			if err != nil {
				return 0, err
			}
			if !portalEmpty(state) {
				if state.Name == "minecraft:obsidian" {
					return i, nil
				}
				break
			}
			below, err := read(treeOffset(p, 0, -1, 0))
			if err != nil {
				return 0, err
			}
			if below.Name != "minecraft:obsidian" {
				break
			}
		}
		return 0, nil
	}
	edge, err := distance(p, -dx, -dz)
	if err != nil {
		return s, err
	}
	if edge-1 < 0 {
		return s, nil
	}
	s.Origin = treeOffset(p, -dx*(edge-1), 0, -dz*(edge-1))
	width, err := distance(s.Origin, dx, dz)
	if err != nil {
		return s, err
	}
	if width < 2 || width > 21 {
		return s, nil
	}
	s.Width = width
	height := int32(21)
heightScan:
	for y := int32(0); y < 21; y++ {
		for _, x := range []int32{-1, width} {
			state, err := read(treeOffset(s.Origin, dx*x, y, dz*x))
			if err != nil {
				return s, err
			}
			if state.Name != "minecraft:obsidian" {
				height = y
				break heightScan
			}
		}
		for x := int32(0); x < width; x++ {
			state, err := read(treeOffset(s.Origin, dx*x, y, dz*x))
			if err != nil {
				return s, err
			}
			if !portalEmpty(state) {
				height = y
				break heightScan
			}
			if state.Name == "minecraft:nether_portal" {
				s.PortalBlocks++
			}
		}
	}
	if height < 3 || height > 21 {
		return s, nil
	}
	for x := int32(0); x < width; x++ {
		state, err := read(treeOffset(s.Origin, dx*x, height, dz*x))
		if err != nil {
			return s, err
		}
		if state.Name != "minecraft:obsidian" {
			return s, nil
		}
	}
	s.Height = height
	return s, nil
}

func FindEmptyNetherPortalShape(world FeatureBlockWorld, minY int32, origin FeaturePosition, preferredAxis string) (NetherPortalShape, bool, error) {
	other := "z"
	if preferredAxis == "z" {
		other = "x"
	}
	for _, axis := range []string{preferredAxis, other} {
		s, err := InspectNetherPortalShape(world, minY, origin, axis)
		if err != nil {
			return s, false, err
		}
		if s.Valid() && s.PortalBlocks == 0 {
			return s, true, nil
		}
	}
	return NetherPortalShape{}, false, nil
}

func FillNetherPortalShape(world FeatureBlockWorld, s NetherPortalShape) error {
	if !s.Valid() {
		return fmt.Errorf("cannot fill invalid Nether portal")
	}
	dx, dz := int32(-1), int32(0)
	if s.Axis == "z" {
		dx, dz = 0, 1
	} else if s.Axis != "x" {
		return fmt.Errorf("invalid Nether portal axis %q", s.Axis)
	}
	for x := int32(0); x < s.Width; x++ {
		for y := int32(0); y < s.Height; y++ {
			if !world.InWorld(treeOffset(s.Origin, dx*x, y, dz*x)) {
				return fmt.Errorf("Nether portal outside dimension bounds")
			}
		}
	}
	state := SurfaceBlockState{Name: "minecraft:nether_portal", Properties: map[string]string{"axis": s.Axis}}
	// BlockPos.betweenClosed iterates X before Y before Z.
	for z := int32(0); z < max(int32(1), dz*s.Width); z++ {
		for y := int32(0); y < s.Height; y++ {
			for x := int32(0); x < max(int32(1), -dx*s.Width); x++ {
				p := treeOffset(s.Origin, 0, y, z)
				if s.Axis == "x" {
					p.X -= s.Width - 1
					p.X += x
				}
				if err := world.SetBlock(p, state); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
