package worldgen

import (
	"embed"
	"fmt"
	"strings"
	"sync"
)

type RuinedPortalVerticalPlacement string

const (
	RuinedPortalOnLandSurface RuinedPortalVerticalPlacement = "on_land_surface"
	RuinedPortalPartlyBuried  RuinedPortalVerticalPlacement = "partly_buried"
	RuinedPortalOnOceanFloor  RuinedPortalVerticalPlacement = "on_ocean_floor"
	RuinedPortalInMountain    RuinedPortalVerticalPlacement = "in_mountain"
	RuinedPortalUnderground   RuinedPortalVerticalPlacement = "underground"
	RuinedPortalInNether      RuinedPortalVerticalPlacement = "in_nether"
)

type RuinedPortalProperties struct {
	Cold                  bool    `json:"cold"`
	Mossiness             float32 `json:"mossiness"`
	AirPocket             bool    `json:"air_pocket"`
	Overgrown             bool    `json:"overgrown"`
	Vines                 bool    `json:"vines"`
	ReplaceWithBlackstone bool    `json:"replace_with_blackstone"`
}

//go:embed ruined_portal_data/*.nbt
var ruinedPortalAssets embed.FS

var loadRuinedPortalTemplates = sync.OnceValues(func() (map[string]*StructureTemplate, error) {
	result := make(map[string]*StructureTemplate)
	for _, family := range []struct {
		name  string
		count int
	}{{"portal", 10}, {"giant_portal", 3}} {
		for index := 1; index <= family.count; index++ {
			name := fmt.Sprintf("%s_%d", family.name, index)
			data, err := ruinedPortalAssets.ReadFile("ruined_portal_data/" + name + ".nbt")
			if err != nil {
				return nil, err
			}
			template, err := DecodeStructureTemplate(data)
			if err != nil {
				return nil, fmt.Errorf("ruined portal %s: %w", name, err)
			}
			result["minecraft:ruined_portal/"+name] = template
		}
	}
	return result, nil
})

func LoadRuinedPortalTemplate(name string) (*StructureTemplate, error) {
	if !strings.Contains(name, ":") {
		name = "minecraft:ruined_portal/" + strings.TrimPrefix(name, "ruined_portal/")
	}
	templates, err := loadRuinedPortalTemplates()
	if err != nil {
		return nil, err
	}
	template := templates[name]
	if template == nil {
		return nil, fmt.Errorf("unsupported ruined portal template %q", name)
	}
	return template, nil
}
