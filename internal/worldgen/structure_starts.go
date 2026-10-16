package worldgen

import (
	"atlas/internal/nbt"
	"fmt"
	"sort"
)

// StructurePiece retains the original saved identity, orientation and piece
// data. Generating a start does not postprocess its blocks; only complete
// supported piece placers may be wired into the final chunk pipeline.
type StructurePiece struct {
	ID              string
	Box             StructureBoundingBox
	Orientation     string
	GenerationDepth int32
	Data            nbt.Object
}

// StructureStart is the persisted STRUCTURE_STARTS-stage result. References
// counts player /locate use; it is distinct from the neighboring chunks' packed
// structure-start references.
type StructureStart struct {
	Structure         string
	Chunk             StructureChunkPosition
	References        int32
	TerrainAdjustment string
	Pieces            []StructurePiece
}

// Clone detaches every saved piece tag for private cross-chunk postprocessing.
// Cached STRUCTURE_STARTS must remain immutable while workers replay their
// block placement, vertical alignment and saved chest/entity flags.
func (start StructureStart) Clone() StructureStart {
	start.Pieces = append([]StructurePiece(nil), start.Pieces...)
	for index := range start.Pieces {
		piece := &start.Pieces[index]
		if piece.Data != nil {
			piece.Data = cloneFeatureNBT(nbt.Tag{Type: nbt.Compound, Value: piece.Data}).Value.(nbt.Object)
		}
	}
	return start
}

// StructureNBTForChunk encodes ChunkSerializer's exact compound field names:
// lowercase "starts" and capitalized "References". Only starts owned by this
// chunk are saved; neighboring ownership is expressed by packed LongArrays.
func StructureNBTForChunk(chunkX, chunkZ int32, starts []StructureStart, references map[string][]int64) (nbt.Object, error) {
	savedStarts, savedReferences := make(nbt.Object), make(nbt.Object)
	for _, start := range starts {
		if start.Chunk.X != chunkX || start.Chunk.Z != chunkZ {
			continue
		}
		if start.Structure == "" {
			return nil, fmt.Errorf("saved structure start has no registry identity")
		}
		data, err := start.NBT()
		if err != nil {
			return nil, err
		}
		savedStarts[start.Structure] = nbt.Tag{Type: nbt.Compound, Value: data}
	}
	for structure, positions := range references {
		if len(positions) == 0 {
			continue
		}
		copy := append([]int64(nil), positions...)
		sort.Slice(copy, func(i, j int) bool { return copy[i] < copy[j] })
		unique := copy[:1]
		for _, position := range copy[1:] {
			if position != unique[len(unique)-1] {
				unique = append(unique, position)
			}
		}
		savedReferences[structure] = nbt.Tag{Type: nbt.LongArray, Value: unique}
	}
	return nbt.Object{
		"starts":     {Type: nbt.Compound, Value: savedStarts},
		"References": {Type: nbt.Compound, Value: savedReferences},
	}, nil
}

func (piece StructurePiece) NBT() (nbt.Object, error) {
	orientation := int32(-1)
	switch piece.Orientation {
	case "":
	case "south":
		orientation = 0
	case "west":
		orientation = 1
	case "north":
		orientation = 2
	case "east":
		orientation = 3
	default:
		return nil, fmt.Errorf("unsupported saved structure orientation %q", piece.Orientation)
	}
	result := make(nbt.Object, len(piece.Data)+4)
	for name, value := range piece.Data {
		result[name] = value
	}
	b := piece.Box
	result["id"] = nbt.Tag{Type: nbt.String, Value: piece.ID}
	result["BB"] = nbt.Tag{Type: nbt.IntArray, Value: []int32{b.MinX, b.MinY, b.MinZ, b.MaxX, b.MaxY, b.MaxZ}}
	result["O"] = nbt.Tag{Type: nbt.Int, Value: orientation}
	result["GD"] = nbt.Tag{Type: nbt.Int, Value: piece.GenerationDepth}
	return result, nil
}

func (start StructureStart) NBT() (nbt.Object, error) {
	if len(start.Pieces) == 0 {
		return nbt.Object{"id": {Type: nbt.String, Value: "INVALID"}}, nil
	}
	children := make([]nbt.Tag, 0, len(start.Pieces))
	for _, piece := range start.Pieces {
		data, err := piece.NBT()
		if err != nil {
			return nil, err
		}
		children = append(children, nbt.Tag{Type: nbt.Compound, Value: data})
	}
	return nbt.Object{
		"id":         {Type: nbt.String, Value: start.Structure},
		"ChunkX":     {Type: nbt.Int, Value: start.Chunk.X},
		"ChunkZ":     {Type: nbt.Int, Value: start.Chunk.Z},
		"references": {Type: nbt.Int, Value: start.References},
		"Children":   {Type: nbt.List, ListType: nbt.Compound, Value: children},
	}, nil
}

// BoundingBox includes the original extra twelve blocks for structures with
// terrain adaptation. Individual piece bounds remain unexpanded for placement.
func (start StructureStart) BoundingBox() (StructureBoundingBox, bool) {
	if len(start.Pieces) == 0 {
		return StructureBoundingBox{}, false
	}
	box := start.Pieces[0].Box
	for _, piece := range start.Pieces[1:] {
		b := piece.Box
		box.MinX, box.MinY, box.MinZ = min(box.MinX, b.MinX), min(box.MinY, b.MinY), min(box.MinZ, b.MinZ)
		box.MaxX, box.MaxY, box.MaxZ = max(box.MaxX, b.MaxX), max(box.MaxY, b.MaxY), max(box.MaxZ, b.MaxZ)
	}
	if start.TerrainAdjustment != "" && start.TerrainAdjustment != "none" {
		box.MinX, box.MinY, box.MinZ = box.MinX-12, box.MinY-12, box.MinZ-12
		box.MaxX, box.MaxY, box.MaxZ = box.MaxX+12, box.MaxY+12, box.MaxZ+12
	}
	return box, true
}

// BuildStructureReferences checks the same eight-chunk start neighborhood and
// inclusive horizontal bounding-box intersection used by ChunkGenerator.
// The caller must provide all starts in that neighborhood before generating
// noise or postprocessing; missing neighbors must not be assumed empty.
func BuildStructureReferences(chunkX, chunkZ int32, starts []StructureStart) map[string][]int64 {
	result := make(map[string][]int64)
	seen := make(map[string]map[int64]bool)
	minX, minZ := chunkX*16, chunkZ*16
	for _, start := range starts {
		if start.Chunk.X < chunkX-8 || start.Chunk.X > chunkX+8 || start.Chunk.Z < chunkZ-8 || start.Chunk.Z > chunkZ+8 {
			continue
		}
		box, valid := start.BoundingBox()
		if !valid || box.MaxX < minX || box.MinX > minX+15 || box.MaxZ < minZ || box.MinZ > minZ+15 {
			continue
		}
		packed := int64(uint64(uint32(start.Chunk.X)) | uint64(uint32(start.Chunk.Z))<<32)
		if seen[start.Structure] == nil {
			seen[start.Structure] = make(map[int64]bool)
		}
		if !seen[start.Structure][packed] {
			result[start.Structure] = append(result[start.Structure], packed)
			seen[start.Structure][packed] = true
		}
	}
	for _, references := range result {
		sort.Slice(references, func(i, j int) bool { return references[i] < references[j] })
	}
	return result
}

// StructureBaseOccupiedHeightAt is ChunkGenerator.getFirstOccupiedHeight for
// WORLD_SURFACE_WG: the Y of the highest non-air base-noise block, before
// surface, structures or features, rather than the first empty block above it.
type StructureBaseOccupiedHeightAt func(x, z int32) (int32, error)

// GenerateScatteredStructureStart implements the exact vanilla initial start
// and saved piece constructor for desert/jungle temples and swamp huts. A nil result
// means terrain or biome rejected the start. Their block placement, archaeology
// and entity postprocessing are separate stages and are not implied here.
func GenerateScatteredStructureStart(seed int64, structure string, chunkX, chunkZ int32, seaLevel int32,
	biomeAtQuart StructureBiomeAtQuart, occupiedHeightAt StructureBaseOccupiedHeightAt) (*StructureStart, error) {
	structure = densityName(structure)
	if structure != "minecraft:desert_pyramid" && structure != "minecraft:swamp_hut" && structure != "minecraft:jungle_pyramid" {
		return nil, fmt.Errorf("unsupported scattered structure start %q", structure)
	}
	if biomeAtQuart == nil || occupiedHeightAt == nil {
		return nil, fmt.Errorf("scattered structure starts require base height and biome sources")
	}
	minX, minZ := chunkX*16, chunkZ*16
	if structure == "minecraft:desert_pyramid" || structure == "minecraft:jungle_pyramid" {
		width, depth := int32(21), int32(21)
		if structure == "minecraft:jungle_pyramid" {
			width, depth = 12, 15
		}
		for _, offset := range [][2]int32{{0, 0}, {0, depth}, {width, 0}, {width, depth}} {
			y, err := occupiedHeightAt(minX+offset[0], minZ+offset[1])
			if err != nil {
				return nil, err
			}
			if y < seaLevel {
				return nil, nil
			}
		}
	}
	y, err := occupiedHeightAt(minX+8, minZ+8)
	if err != nil {
		return nil, err
	}
	biome, err := biomeAtQuart((minX+8)>>2, y>>2, (minZ+8)>>2)
	if err != nil {
		return nil, err
	}
	biomes, err := StructureBiomeNames(structure)
	if err != nil {
		return nil, err
	}
	validBiome := false
	for _, name := range biomes {
		validBiome = validBiome || name == biome
	}
	if !validBiome {
		return nil, nil
	}
	random := NewLegacyRandom(seed)
	xMultiplier, zMultiplier := random.NextLong(), random.NextLong()
	mixed := uint64(int64(chunkX))*xMultiplier ^ uint64(int64(chunkZ))*zMultiplier ^ uint64(seed)
	random = NewLegacyRandom(int64(mixed))
	orientation := [...]string{"north", "east", "south", "west"}[random.NextIntBound(4)]
	width, height, depth, pieceID := int32(21), int32(15), int32(21), "minecraft:tedp"
	if structure == "minecraft:swamp_hut" {
		width, height, depth, pieceID = 7, 7, 9, "minecraft:tesh"
	} else if structure == "minecraft:jungle_pyramid" {
		width, height, depth, pieceID = 12, 10, 15, "minecraft:tejp"
	}
	xSize, zSize := width, depth
	if orientation == "east" || orientation == "west" {
		xSize, zSize = depth, width
	}
	piece := StructurePiece{
		ID: pieceID, Box: StructureBoundingBox{minX, 64, minZ, minX + xSize - 1, 64 + height - 1, minZ + zSize - 1}, Orientation: orientation,
		Data: nbt.Object{
			"Width": {Type: nbt.Int, Value: width}, "Height": {Type: nbt.Int, Value: height},
			"Depth": {Type: nbt.Int, Value: depth}, "HPos": {Type: nbt.Int, Value: int32(-1)},
		},
	}
	if structure == "minecraft:desert_pyramid" {
		for index := 0; index < 4; index++ {
			piece.Data[fmt.Sprintf("hasPlacedChest%d", index)] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		}
	} else if structure == "minecraft:swamp_hut" {
		piece.Data["Witch"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		piece.Data["Cat"] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
	} else {
		for _, field := range []string{"placedMainChest", "placedHiddenChest", "placedTrap1", "placedTrap2"} {
			piece.Data[field] = nbt.Tag{Type: nbt.Byte, Value: int8(0)}
		}
	}
	return &StructureStart{Structure: structure, Chunk: StructureChunkPosition{chunkX, chunkZ}, TerrainAdjustment: "none", Pieces: []StructurePiece{piece}}, nil
}
