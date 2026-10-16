package worldgen

import (
	"encoding/json"
	"fmt"
)

type RuinedPortalStartSources struct {
	MinY         int32
	BiomeAtQuart StructureBiomeAtQuart
	BaseColumnAt StructureBaseNoiseColumnAt
}

type ruinedPortalSetup struct {
	Placement             RuinedPortalVerticalPlacement `json:"placement"`
	AirPocketProbability  float32                       `json:"air_pocket_probability"`
	Mossiness             float32                       `json:"mossiness"`
	Overgrown             bool                          `json:"overgrown"`
	Vines                 bool                          `json:"vines"`
	CanBeCold             bool                          `json:"can_be_cold"`
	ReplaceWithBlackstone bool                          `json:"replace_with_blackstone"`
	Weight                float32                       `json:"weight"`
}

// GenerateRuinedPortalStructureStart preserves all seven original setup
// variants, selection draws, giant templates, four-corner column descent and
// final biome/cold checks. Candidate placement is checked by the caller.
func GenerateRuinedPortalStructureStart(seed int64, chunkX, chunkZ int32, structure string, sources RuinedPortalStartSources) (*StructureStart, error) {
	if sources.BiomeAtQuart == nil || sources.BaseColumnAt == nil {
		return nil, fmt.Errorf("ruined portal start requires biomes and base columns")
	}
	structure = densityName(structure)
	definition, err := BuiltinStructureJSON(structure)
	if err != nil {
		return nil, err
	}
	var settings struct {
		Type              string              `json:"type"`
		Setups            []ruinedPortalSetup `json:"setups"`
		TerrainAdjustment string              `json:"terrain_adaptation"`
	}
	if err := json.Unmarshal(definition, &settings); err != nil {
		return nil, err
	}
	if settings.Type != "minecraft:ruined_portal" || len(settings.Setups) == 0 {
		return nil, fmt.Errorf("%s is not an original ruined portal structure", structure)
	}
	random := NewLegacyRandom(seed)
	xm, zm := random.NextLong(), random.NextLong()
	random = NewLegacyRandom(int64(uint64(int64(chunkX))*xm ^ uint64(int64(chunkZ))*zm ^ uint64(seed)))
	setup := settings.Setups[0]
	if len(settings.Setups) > 1 {
		var total float32
		for _, item := range settings.Setups {
			total += item.Weight
		}
		value := random.NextFloat()
		selected := false
		for _, item := range settings.Setups {
			value -= item.Weight / total
			if value < 0 {
				setup = item
				selected = true
				break
			}
		}
		if !selected {
			return nil, fmt.Errorf("ruined portal setup selection did not select an entry")
		}
	}
	properties := RuinedPortalProperties{Mossiness: setup.Mossiness, Overgrown: setup.Overgrown, Vines: setup.Vines, ReplaceWithBlackstone: setup.ReplaceWithBlackstone}
	switch setup.AirPocketProbability {
	case 0:
	case 1:
		properties.AirPocket = true
	default:
		properties.AirPocket = random.NextFloat() < setup.AirPocketProbability
	}
	name := ""
	if random.NextFloat() < 0.05 {
		name = fmt.Sprintf("minecraft:ruined_portal/giant_portal_%d", random.NextIntBound(3)+1)
	} else {
		name = fmt.Sprintf("minecraft:ruined_portal/portal_%d", random.NextIntBound(10)+1)
	}
	template, err := LoadRuinedPortalTemplate(name)
	if err != nil {
		return nil, err
	}
	rotation := TemplateRotation(random.NextIntBound(4))
	mirror := TemplateMirrorFrontBack
	if random.NextFloat() < 0.5 {
		mirror = TemplateMirrorNone
	}
	origin := FeaturePosition{chunkX * 16, 0, chunkZ * 16}
	box, err := template.BoundingBox(origin, FeaturePosition{template.Size.X / 2, 0, template.Size.Z / 2}, mirror, rotation)
	if err != nil {
		return nil, err
	}
	center := FeaturePosition{box.MinX + (box.MaxX-box.MinX+1)/2, 0, box.MinZ + (box.MaxZ-box.MinZ+1)/2}
	column, err := sources.BaseColumnAt(center.X, center.Z)
	if err != nil {
		return nil, err
	}
	if column == nil {
		return nil, fmt.Errorf("ruined portal base column is nil")
	}
	oceanFloor := setup.Placement == RuinedPortalOnOceanFloor
	surface := column.OccupiedHeight(oceanFloor)
	lowest := sources.MinY + 15
	interval := func(low, high int32) int32 {
		if low < high {
			return low + random.NextIntBound(high-low+1)
		}
		return high
	}
	y := surface
	switch setup.Placement {
	case RuinedPortalInNether:
		if properties.AirPocket {
			y = interval(32, 100)
		} else if random.NextFloat() < 0.5 {
			y = interval(27, 29)
		} else {
			y = interval(29, 100)
		}
	case RuinedPortalInMountain:
		y = interval(70, surface-template.Size.Y)
	case RuinedPortalUnderground:
		y = interval(lowest, surface-template.Size.Y)
	case RuinedPortalPartlyBuried:
		y = surface - template.Size.Y + interval(2, 8)
	}
	var corners []*StructureBaseNoiseColumn
	for _, p := range []FeaturePosition{{box.MinX, 0, box.MinZ}, {box.MaxX, 0, box.MinZ}, {box.MinX, 0, box.MaxZ}, {box.MaxX, 0, box.MaxZ}} {
		column, err := sources.BaseColumnAt(p.X, p.Z)
		if err != nil {
			return nil, err
		}
		if column == nil {
			return nil, fmt.Errorf("ruined portal corner column is nil")
		}
		corners = append(corners, column)
	}
	for ; y > lowest; y-- {
		count := 0
		for _, column := range corners {
			material := column.MaterialAt(y)
			if material != NoiseAir && (!oceanFloor || material != NoiseWater && material != NoiseLava) {
				count++
			}
		}
		if count >= 3 {
			break
		}
	}
	origin.Y = y
	biome, err := sources.BiomeAtQuart(origin.X>>2, y>>2, origin.Z>>2)
	if err != nil {
		return nil, err
	}
	allowed, err := StructureBiomeNames(structure)
	if err != nil {
		return nil, err
	}
	valid := false
	for _, name := range allowed {
		valid = valid || name == biome
	}
	if !valid {
		return nil, nil
	}
	if setup.CanBeCold {
		properties.Cold, err = BiomeColdAt(biome, origin.X, y, origin.Z)
		if err != nil {
			return nil, err
		}
	}
	piece, err := NewRuinedPortalTemplatePiece(name, origin, setup.Placement, properties, mirror, rotation)
	if err != nil {
		return nil, err
	}
	adjustment := settings.TerrainAdjustment
	if adjustment == "" {
		adjustment = "none"
	}
	return &StructureStart{Structure: structure, Chunk: StructureChunkPosition{chunkX, chunkZ}, TerrainAdjustment: adjustment, Pieces: []StructurePiece{*piece}}, nil
}
