package worldgen

import "fmt"

// NetherPortalRectangle ports BlockUtil.getLargestRectangleAround with the
// portal block's exact state predicate and the stock 21-block scan limits.
func NetherPortalRectangle(world FeatureBlockWorld, origin FeaturePosition) (PortalRectangle, error) {
	state, err := world.BlockAt(origin)
	if err != nil {
		return PortalRectangle{}, err
	}
	axis := state.Properties["axis"]
	if state.Name != "minecraft:nether_portal" || axis != "x" && axis != "z" {
		return PortalRectangle{}, fmt.Errorf("source block is not a Nether portal")
	}
	dx, dz := int32(1), int32(0)
	if axis == "z" {
		dx, dz = 0, 1
	}
	distance := func(p FeaturePosition, sx, sy, sz, limit int32) (int32, error) {
		var count int32
		for count < limit {
			p = treeOffset(p, sx, sy, sz)
			s, err := world.BlockAt(p)
			if err != nil {
				return 0, err
			}
			if s.Name != state.Name || s.Properties["axis"] != axis {
				break
			}
			count++
		}
		return count, nil
	}
	left, err := distance(origin, -dx, 0, -dz, 21)
	if err != nil {
		return PortalRectangle{}, err
	}
	right, err := distance(origin, dx, 0, dz, 21)
	if err != nil {
		return PortalRectangle{}, err
	}
	type vertical struct{ down, up int32 }
	spans := make([]vertical, left+right+1)
	spans[left].down, err = distance(origin, 0, -1, 0, 21)
	if err != nil {
		return PortalRectangle{}, err
	}
	spans[left].up, err = distance(origin, 0, 1, 0, 21)
	if err != nil {
		return PortalRectangle{}, err
	}
	for n := int32(1); n <= left; n++ {
		p := treeOffset(origin, -dx*n, 0, -dz*n)
		prev := spans[left-n+1]
		spans[left-n].down, err = distance(p, 0, -1, 0, prev.down)
		if err != nil {
			return PortalRectangle{}, err
		}
		spans[left-n].up, err = distance(p, 0, 1, 0, prev.up)
		if err != nil {
			return PortalRectangle{}, err
		}
	}
	for n := int32(1); n <= right; n++ {
		p := treeOffset(origin, dx*n, 0, dz*n)
		prev := spans[left+n-1]
		spans[left+n].down, err = distance(p, 0, -1, 0, prev.down)
		if err != nil {
			return PortalRectangle{}, err
		}
		spans[left+n].up, err = distance(p, 0, 1, 0, prev.up)
		if err != nil {
			return PortalRectangle{}, err
		}
	}
	bottom := spans[left].down
	heights := make([]int32, len(spans))
	var bestX, bestY, bestWidth, bestHeight int32
	for row := bottom; row >= 0; row-- {
		for i, v := range spans {
			low, high := bottom-v.down, bottom+v.up
			heights[i] = 0
			if row >= low && row <= high {
				heights[i] = high + 1 - row
			}
		}
		x, width, height := largestPortalHistogram(heights)
		if width*height > bestWidth*bestHeight {
			bestX, bestY, bestWidth, bestHeight = x, row, width, height
		}
	}
	return PortalRectangle{Origin: treeOffset(origin, dx*(bestX-left), bestY-bottom, dz*(bestX-left)), Axis: axis, Width: bestWidth, Height: bestHeight}, nil
}

func largestPortalHistogram(heights []int32) (int32, int32, int32) {
	left, right, height := 0, 0, int32(0)
	stack := []int{0}
	for i := 1; i <= len(heights); i++ {
		next := int32(0)
		if i < len(heights) {
			next = heights[i]
		}
		for len(stack) > 0 {
			h := heights[stack[len(stack)-1]]
			if next >= h {
				stack = append(stack, i)
				break
			}
			stack = stack[:len(stack)-1]
			x := 0
			if len(stack) > 0 {
				x = stack[len(stack)-1] + 1
			}
			if h*int32(i-x) > height*int32(right-left) {
				left, right, height = x, i, h
			}
		}
		if len(stack) == 0 {
			stack = append(stack, i)
		}
	}
	return int32(left), int32(right - left), height
}

// PortalRelativePosition is PortalShape.getRelativePosition. Living players
// subsequently reset its forward offset to zero before destination placement.
func PortalRelativePosition(rect PortalRectangle, position [3]float64, width, height float32, living bool) [3]float64 {
	w, h := float64(width), float64(height)
	axis := 0
	if rect.Axis == "z" {
		axis = 2
	}
	base := float64(rect.Origin.X)
	other := float64(rect.Origin.Z)
	if axis == 2 {
		base, other = other, base
	}
	relative := [3]float64{0.5, 0, position[2-axis] - (other + 0.5)}
	if room := float64(rect.Width) - w; room > 0 {
		relative[0] = javaClamp((position[axis]-(base+w/2))/room, 0, 1)
	}
	if room := float64(rect.Height) - h; room > 0 {
		relative[1] = javaClamp((position[1]-float64(rect.Origin.Y))/room, 0, 1)
	}
	if living {
		relative[2] = 0
	}
	return relative
}

func PortalExitPosition(rect PortalRectangle, sourceAxis string, relative [3]float64, width, height float32, yaw float32, velocity [3]float64) (position [3]float64, exitYaw float32, exitVelocity [3]float64) {
	along := float64(width)/2 + (float64(rect.Width)-float64(width))*relative[0]
	up := (float64(rect.Height) - float64(height)) * relative[1]
	forward := 0.5 + relative[2]
	position = [3]float64{float64(rect.Origin.X) + along, float64(rect.Origin.Y) + up, float64(rect.Origin.Z) + forward}
	if rect.Axis == "z" {
		position[0], position[2] = float64(rect.Origin.X)+forward, float64(rect.Origin.Z)+along
	}
	exitYaw, exitVelocity = yaw, velocity
	if rect.Axis != sourceAxis {
		exitYaw += 90
		exitVelocity = [3]float64{velocity[2], velocity[1], -velocity[0]}
	}
	return
}
