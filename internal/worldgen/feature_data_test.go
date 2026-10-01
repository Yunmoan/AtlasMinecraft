package worldgen

import (
	"encoding/json"
	"testing"
)

func TestBuiltInFeatureDefinitionsCoverOverworldOrder(t *testing.T) {
	order, err := loadOverworldFeatureOrder()
	if err != nil {
		t.Fatal(err)
	}
	for _, feature := range order.features {
		placed, err := LoadPlacedFeatureDefinition(feature.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := LoadConfiguredFeatureDefinition(placed.Feature); err != nil {
			t.Fatal(err)
		}
	}
	placed, err := LoadPlacedFeatureDefinition("minecraft:ore_iron_upper")
	if err != nil {
		t.Fatal(err)
	}
	if placed.Feature != "minecraft:ore_iron" || len(placed.Placement) != 4 {
		t.Fatalf("unexpected iron ore placement: %+v", placed)
	}
	configured, err := LoadConfiguredFeatureDefinition(placed.Feature)
	if err != nil {
		t.Fatal(err)
	}
	if configured.Type != "minecraft:ore" || !json.Valid(configured.Config) {
		t.Fatalf("unexpected iron ore configuration: %+v", configured)
	}
	if _, err := LoadPlacedFeatureDefinition("minecraft:not_a_feature"); err == nil {
		t.Fatal("unknown placed feature accepted")
	}
}
