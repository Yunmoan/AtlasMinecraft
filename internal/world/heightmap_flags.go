package world

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"math/bits"
	"strings"
)

//go:embed heightmap_flags.txt
var heightmapFlagsText string

var motionFlags, oceanFloorFlags, faceFullFlags, fluidFlags []byte

func init() {
	fields := strings.Fields(heightmapFlagsText)
	if len(fields) != 12 || fields[0] != "motion_blocking_states" || fields[1] != "26684" ||
		fields[3] != "ocean_floor_states" || fields[4] != "26684" ||
		fields[6] != "face_full_states" || fields[7] != "26684" ||
		fields[9] != "fluid_states" || fields[10] != "26684" {
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
	faceFullFlags, err = base64.StdEncoding.DecodeString(fields[8])
	if err != nil || len(faceFullFlags) != len(motionFlags) {
		panic("invalid 1.21.1 face-full reference")
	}
	fluidFlags, err = base64.StdEncoding.DecodeString(fields[11])
	if err != nil || len(fluidFlags) != len(motionFlags) {
		panic("invalid 1.21.1 fluid reference")
	}
}

func stateFlag(flags []byte, state uint32) bool {
	return flags[state/8]&(1<<(state%8)) != 0
}

func (c *Chunk) anvilHeightmaps() (motion, surface, ocean []int64) {
	bitsPer := bits.Len32(uint32(c.maxY - c.minY))
	perLong := 64 / bitsPer
	longCount := (256 + perLong - 1) / perLong
	motion = make([]int64, longCount)
	surface = make([]int64, longCount)
	ocean = make([]int64, longCount)
	for z := 0; z < 16; z++ {
		for x := 0; x < 16; x++ {
			index := z*16 + x
			motionY, surfaceY, oceanY := int32(c.minY), int32(c.heights[index]), int32(c.minY)
			for y := surfaceY - 1; y >= c.minY; y-- {
				state := c.GetBlock(x, z, y)
				if motionY == c.minY && stateFlag(motionFlags, state) {
					motionY = y + 1
				}
				if oceanY == c.minY && stateFlag(oceanFloorFlags, state) {
					oceanY = y + 1
				}
				if motionY != c.minY && oceanY != c.minY {
					break
				}
			}
			shift := uint(index%perLong) * uint(bitsPer)
			motion[index/perLong] |= int64(motionY-c.minY) << shift
			surface[index/perLong] |= int64(surfaceY-c.minY) << shift
			ocean[index/perLong] |= int64(oceanY-c.minY) << shift
		}
	}
	return motion, surface, ocean
}
