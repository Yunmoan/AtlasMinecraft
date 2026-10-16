package worldgen

import "fmt"

// FeatureBrightnessReader supplies an existing light engine's raw brightness
// during FEATURES. It includes sky light and receives vanilla's sky-darkening
// argument. A fully generated/lit neighboring chunk can provide this reader.
type FeatureBrightnessReader func(FeaturePosition, int32) (int32, error)

func (region *OverworldFeatureRegion) BindFeatureBrightness(reader FeatureBrightnessReader) {
	region.readBrightness = reader
}

func (region *OverworldFeatureRegion) FeatureRawBrightness(p FeaturePosition, skyDarkening int32) (int32, error) {
	if region.readBrightness != nil {
		return region.readBrightness(p, skyDarkening)
	}
	// In a fresh Overworld region, FEATURES precedes INITIALIZE_LIGHT.
	// SkyLightSectionStorage.getLightValue reads an absent light column as 15;
	// it does not scan the terrain height or trace shadows. Block light is 0.
	// The official MushroomFeatureProbe executes the real LevelLightEngine
	// without light sections and checks these values, including underground.
	return max(0, 15-skyDarkening), nil
}

func featureRawBrightness(world FeatureBlockWorld, p FeaturePosition, skyDarkening int32) (int32, error) {
	reader, ok := world.(interface {
		FeatureRawBrightness(FeaturePosition, int32) (int32, error)
	})
	if !ok {
		return 0, fmt.Errorf("feature survival requires generation-time brightness")
	}
	return reader.FeatureRawBrightness(p, skyDarkening)
}

func featureMushroomSurvives(world FeatureBlockWorld, p FeaturePosition, below SurfaceBlockState) (bool, error) {
	if vegetationTags["mushroom_grow_block"][below.Name] {
		return true, nil
	}
	brightness, err := featureRawBrightness(world, p, 0)
	if err != nil || brightness >= 13 {
		return false, err
	}
	e := featureEditor{world: world}
	return multifaceRules(&e, below).SolidRender, e.err
}
