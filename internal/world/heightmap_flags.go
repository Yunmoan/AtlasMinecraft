package world

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
)

//go:embed heightmap_flags.txt
var heightmapFlagsText string

var motionFlags, oceanFloorFlags []byte

func init() {
	fields := strings.Fields(heightmapFlagsText)
	if len(fields) != 6 || fields[0] != "motion_blocking_states" || fields[1] != "26684" ||
		fields[3] != "ocean_floor_states" || fields[4] != "26684" {
		panic("invalid 1.21.1 heightmap reference")
	}
	var err error
	motionFlags, err = base64.StdEncoding.DecodeString(fields[2])
	if err != nil {
		panic(fmt.Sprintf("decode motion heightmap reference: %v", err))
	}
	oceanFloorFlags, err = base64.StdEncoding.DecodeString(fields[5])
	if err != nil || len(motionFlags) != (MaxBlockState+8)/8 || len(oceanFloorFlags) != len(motionFlags) {
		panic("invalid 1.21.1 heightmap reference length")
	}
}

func stateFlag(flags []byte, state uint32) bool {
	return flags[state/8]&(1<<(state%8)) != 0
}

func (c *Chunk) anvilHeightmaps() (motion, surface, ocean []int64) {
	motion = make([]int64, 37)
	surface = make([]int64, 37)
	ocean = make([]int64, 37)
	for z := 0; z < 16; z++ {
		for x := 0; x < 16; x++ {
			index := z*16 + x
			motionY, surfaceY, oceanY := int32(MinY), int32(c.heights[index]), int32(MinY)
			for y := surfaceY - 1; y >= MinY; y-- {
				state := c.GetBlock(x, z, y)
				if motionY == MinY && stateFlag(motionFlags, state) {
					motionY = y + 1
				}
				if oceanY == MinY && stateFlag(oceanFloorFlags, state) {
					oceanY = y + 1
				}
				if motionY != MinY && oceanY != MinY {
					break
				}
			}
			shift := uint(index%7) * 9
			motion[index/7] |= int64(motionY-MinY) << shift
			surface[index/7] |= int64(surfaceY-MinY) << shift
			ocean[index/7] |= int64(oceanY-MinY) << shift
		}
	}
	return motion, surface, ocean
}
