package worldgen

import (
	"atlas/internal/nbt"
	"encoding/json"
	"fmt"
	"sync"
)

type jigsawFeatureEntry struct {
	load func() (configuredFeaturePlacer, error)
}

var jigsawFeatureCache sync.Map

func jigsawFeaturePlan(raw json.RawMessage) (configuredFeaturePlacer, error) {
	key := string(raw)
	if entry, found := jigsawFeatureCache.Load(key); found {
		return entry.(jigsawFeatureEntry).load()
	}
	entry, _ := jigsawFeatureCache.LoadOrStore(key, jigsawFeatureEntry{sync.OnceValues(func() (configuredFeaturePlacer, error) { return compileNestedFeaturePlacement(raw, 0) })})
	return entry.(jigsawFeatureEntry).load()
}

// FeaturePoolElement deliberately ignores the template clip, pivot, rotation,
// and liquid settings. Its original PlacedFeature runs at the piece origin
// and can write into neighboring chunks.
func (world *structureRecordingWorld) PlaceJigsawFeature(raw json.RawMessage, random StructurePieceRandom, origin FeaturePosition) (bool, error) {
	stream, ok := random.(*WorldgenRandom)
	if !ok {
		return false, fmt.Errorf("Jigsaw feature requires the decoration WorldgenRandom")
	}
	plan, err := jigsawFeaturePlan(raw)
	if err != nil {
		return false, err
	}
	previous := world.region.structureRecording
	world.region.structureRecording = world.result
	defer func() { world.region.structureRecording = previous }()
	return plan(world.region, stream, origin)
}

func jigsawFeatureOrigins(piece StructurePiece) ([]FeaturePosition, error) {
	origin, _, _, err := jigsawSavedOriginRotation(&piece)
	if err != nil {
		return nil, err
	}
	data, ok := piece.Data.Object("pool_element")
	if !ok {
		return nil, fmt.Errorf("Jigsaw piece lacks pool_element")
	}
	count, err := jigsawFeatureElementCount(data)
	if err != nil || count == 0 {
		return nil, err
	}
	positions := make([]FeaturePosition, count)
	for i := range positions {
		positions[i] = origin
	}
	return positions, nil
}

func jigsawFeatureElementCount(element nbt.Object) (int, error) {
	kind, _ := element.Text("element_type")
	if kind == "minecraft:feature_pool_element" {
		return 1, nil
	}
	if kind != "minecraft:list_pool_element" {
		return 0, nil
	}
	if _, exists := element["elements"]; !exists {
		return 0, nil
	}
	children, ok := element["elements"].Elements()
	if !ok {
		return 0, fmt.Errorf("invalid saved Jigsaw list elements")
	}
	count := 0
	for _, child := range children {
		data, ok := child.Object()
		if !ok {
			return 0, fmt.Errorf("invalid saved Jigsaw child element")
		}
		n, err := jigsawFeatureElementCount(data)
		if err != nil {
			return 0, err
		}
		count += n
	}
	return count, nil
}
