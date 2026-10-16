package worldgen

import "atlas/internal/nbt"

func isOverworldVillageStructure(name string) bool {
	switch name {
	case "minecraft:village_desert", "minecraft:village_plains", "minecraft:village_savanna", "minecraft:village_snowy", "minecraft:village_taiga":
		return true
	default:
		return false
	}
}

func (g *OverworldGenerator) finalizeVillageRequest(region *OverworldFeatureRegion, request *PendingStructureEntity, random *Xoroshiro) error {
	p := pendingEntityBlockPosition(*request)
	qx, qy, qz := BiomeQuartAtBlock(g.seed, p.X, p.Y, p.Z)
	biome, err := region.biomeAtQuart(qx, qy, qz)
	if err != nil {
		return err
	}
	actorSeed := g.seed ^ int64(positionSeed(p.X, p.Y, p.Z))
	finalized, err := FinalizeVillageTemplateEntity(*request, random, NewLegacyRandom(actorSeed), biome, g.structureFullMoon(), false)
	if err != nil {
		return err
	}
	request.Data = finalized.Data
	if request.ID == "minecraft:camel" {
		request.Data["LastPoseTick"] = nbt.Tag{Type: nbt.Long, Value: max(int64(0), g.structureAge()-53)}
	}
	if finalized.ChickenJockey {
		request.ID = "minecraft:chicken"
		request.Persistent = false
		request.Data = finalized.ChickenData
		request.Data["Passengers"] = nbt.Tag{Type: nbt.List, ListType: nbt.Compound, Value: []nbt.Tag{{Type: nbt.Compound, Value: finalized.Data}}}
		request.Data["EggLayTime"] = nbt.Tag{Type: nbt.Int, Value: NewLegacyRandom(actorSeed).NextIntBound(6000) + 6000}
	}
	return nil
}
