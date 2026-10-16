package worldgen

import (
	"fmt"
	"math"
)

// StructureBoundingBox uses inclusive vanilla block-coordinate bounds.
type StructureBoundingBox struct {
	MinX, MinY, MinZ int32
	MaxX, MaxY, MaxZ int32
}

// StructureRigidPiece is the terrain adjustment input attached to a generated
// structure piece. GroundLevelDelta is used by rigid jigsaw pieces; ordinary
// pieces use zero. This contains no templates or structure block placement.
type StructureRigidPiece struct {
	Box               StructureBoundingBox
	TerrainAdjustment string
	GroundLevelDelta  int32
}

// StructureJunction retains the source coordinates used by vanilla's jigsaw
// terrain blending. Destination projection and deltaY belong to piece assembly
// and do not affect the Beardifier density contribution.
type StructureJunction struct{ SourceX, SourceGroundY, SourceZ int32 }

// StructureTerrainPiece groups the density inputs produced by a structure
// start. Jigsaw pieces contribute their box only when projection is rigid;
// their junctions contribute for both rigid and terrain_matching projections.
type StructureTerrainPiece struct {
	RigidPiece StructureRigidPiece
	Jigsaw     bool
	Projection string
	Junctions  []StructureJunction
}

// StructureBeardifier implements the original 1.21.1 terrain adjustment kernel.
// Its copied inputs and stateless Sample method are safe to share between
// generation workers. The empty instance contributes exactly zero.
type StructureBeardifier struct {
	pieces    []StructureRigidPiece
	junctions []StructureJunction
}

func NewStructureBeardifier(pieces []StructureRigidPiece, junctions []StructureJunction) (*StructureBeardifier, error) {
	for _, piece := range pieces {
		b := piece.Box
		if b.MinX > b.MaxX || b.MinY > b.MaxY || b.MinZ > b.MaxZ {
			return nil, fmt.Errorf("invalid structure piece bounding box %v", b)
		}
		switch piece.TerrainAdjustment {
		case "none", "bury", "beard_thin", "beard_box", "encapsulate":
		default:
			return nil, fmt.Errorf("unsupported structure terrain adjustment %q", piece.TerrainAdjustment)
		}
	}
	return &StructureBeardifier{pieces: append([]StructureRigidPiece(nil), pieces...), junctions: append([]StructureJunction(nil), junctions...)}, nil
}

// NewStructureBeardifierForChunk follows Beardifier.forStructuresInChunk after
// the caller has loaded the relevant starts from the chunk's structure
// references. It cannot discover starts that have not yet been generated.
func NewStructureBeardifierForChunk(chunkX, chunkZ int32, terrainPieces []StructureTerrainPiece) (*StructureBeardifier, error) {
	minX, minZ := chunkX*16, chunkZ*16
	var pieces []StructureRigidPiece
	var junctions []StructureJunction
	for _, piece := range terrainPieces {
		rigid := piece.RigidPiece
		if rigid.TerrainAdjustment == "none" {
			continue
		}
		b := rigid.Box
		if b.MaxX < minX-12 || b.MinX > minX+15+12 || b.MaxZ < minZ-12 || b.MinZ > minZ+15+12 {
			continue
		}
		if !piece.Jigsaw {
			rigid.GroundLevelDelta = 0
			pieces = append(pieces, rigid)
			continue
		}
		if piece.Projection != "rigid" && piece.Projection != "terrain_matching" {
			return nil, fmt.Errorf("unsupported jigsaw projection %q", piece.Projection)
		}
		if piece.Projection == "rigid" {
			pieces = append(pieces, rigid)
		}
		for _, junction := range piece.Junctions {
			if junction.SourceX > minX-12 && junction.SourceZ > minZ-12 && junction.SourceX < minX+15+12 && junction.SourceZ < minZ+15+12 {
				junctions = append(junctions, junction)
			}
		}
	}
	return NewStructureBeardifier(pieces, junctions)
}

var structureBeardKernel = func() [24 * 24 * 24]float32 {
	var kernel [24 * 24 * 24]float32
	for z := int32(0); z < 24; z++ {
		for x := int32(0); x < 24; x++ {
			for y := int32(0); y < 24; y++ {
				dx, dy, dz := float64(x-12), float64(y-12)+0.5, float64(z-12)
				lengthSquared := dx*dx + dy*dy + dz*dz
				// Vanilla intentionally uses pow(E, ...) rather than exp(...),
				// then stores the kernel as float before promoting to double.
				kernel[z*24*24+x*24+y] = float32(math.Pow(math.E, -lengthSquared/16))
			}
		}
	}
	return kernel
}()

func structureBuryContribution(x, y, z float64) float64 {
	length := math.Sqrt(x*x + y*y + z*z)
	return clampedMap(length, 0, 6, 1, 0)
}

func structureBeardContribution(x, y, z, groundOffset int32) float64 {
	kx, ky, kz := x+12, y+12, z+12
	if kx < 0 || kx >= 24 || ky < 0 || ky >= 24 || kz < 0 || kz >= 24 {
		return 0
	}
	dy := float64(groundOffset) + 0.5
	dx, dz := float64(x), float64(z)
	lengthSquared := dx*dx + dy*dy + dz*dz
	factor := -dy * structureFastInvSqrt(lengthSquared/2) / 2
	return factor * float64(structureBeardKernel[kz*24*24+kx*24+ky])
}

// This deprecated approximation remains part of vanilla's Beardifier. Using
// 1/sqrt would produce different density and structure terrain blocks.
func structureFastInvSqrt(value float64) float64 {
	half := 0.5 * value
	bits := uint64(6910469410427058090) - (math.Float64bits(value) >> 1)
	result := math.Float64frombits(bits)
	result *= 1.5 - half*result*result
	return result
}

func (b *StructureBeardifier) Sample(x, y, z int32) float64 {
	if b == nil {
		return 0
	}
	value := float64(0)
	for _, piece := range b.pieces {
		box := piece.Box
		dx := max(int32(0), max(box.MinX-x, x-box.MaxX))
		dz := max(int32(0), max(box.MinZ-z, z-box.MaxZ))
		groundY := box.MinY + piece.GroundLevelDelta
		groundOffset := y - groundY
		switch piece.TerrainAdjustment {
		case "none":
		case "bury":
			value += structureBuryContribution(float64(dx), float64(groundOffset)/2, float64(dz))
		case "beard_thin":
			value += structureBeardContribution(dx, groundOffset, dz, groundOffset) * 0.8
		case "beard_box":
			dy := max(int32(0), max(groundY-y, y-box.MaxY))
			value += structureBeardContribution(dx, dy, dz, groundOffset) * 0.8
		case "encapsulate":
			dy := max(int32(0), max(box.MinY-y, y-box.MaxY))
			value += structureBuryContribution(float64(dx)/2, float64(dy)/2, float64(dz)/2) * 0.8
		}
	}
	for _, junction := range b.junctions {
		dx, dy, dz := x-junction.SourceX, y-junction.SourceGroundY, z-junction.SourceZ
		value += structureBeardContribution(dx, dy, dz, dy) * 0.4
	}
	return value
}
