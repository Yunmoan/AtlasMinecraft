package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type ruinedPortalProbeWorld struct {
	*hutProbeWorld
	columnEdits map[FeaturePosition]int32
}

func (world *ruinedPortalProbeWorld) SetBlock(p FeaturePosition, state SurfaceBlockState) error {
	if world.columnEdits == nil {
		world.columnEdits = make(map[FeaturePosition]int32)
	}
	column := FeaturePosition{p.X, 0, p.Z}
	world.columnEdits[column] = max(world.columnEdits[column], p.Y)
	return world.hutProbeWorld.SetBlock(p, state)
}

func (world *ruinedPortalProbeWorld) HeightmapAt(kind string, x, z int32) (int32, error) {
	y := int32(63)
	if world.scenario == 1 {
		y = 60 + (x+2*z)%5
		if y < 60 {
			y += 5
		}
	}
	if world.scenario == 2 || world.scenario == 3 {
		y = 69
	}
	if world.scenario == 4 || world.scenario == 5 {
		y = 319
	}
	y = max(y, world.columnEdits[FeaturePosition{x, 0, z}])
	for ; y >= -64; y-- {
		state, err := world.BlockAt(FeaturePosition{x, y, z})
		if err != nil {
			return 0, err
		}
		occupied := !surfaceAir(state)
		if kind == "OCEAN_FLOOR_WG" {
			occupied, err = oceanFloorSolid(state)
			if err != nil {
				return 0, err
			}
		}
		if occupied {
			return y + 1, nil
		}
	}
	return -64, nil
}

func (world *ruinedPortalProbeWorld) BlockAt(p FeaturePosition) (SurfaceBlockState, error) {
	if state, found := world.edits[p]; found {
		return state, nil
	}
	floor := int32(63)
	if world.scenario == 1 {
		floor = 60 + (p.X+2*p.Z)%5
		if floor < 60 {
			floor += 5
		}
	}
	mod := func(value, divisor int32) int32 {
		result := value % divisor
		if result < 0 {
			result += divisor
		}
		return result
	}
	if world.scenario == 4 && mod(p.X+p.Y+p.Z, 7) == 0 {
		return structureState("bedrock"), nil
	}
	if world.scenario == 5 && mod(p.X+p.Y+p.Z, 3) == 0 {
		return structureState("obsidian"), nil
	}
	if p.Y <= floor {
		if p.Y == floor {
			return structureState("grass_block"), nil
		}
		return structureState("stone"), nil
	}
	if world.scenario == 2 && p.Y <= 69 {
		return structureState("water"), nil
	}
	if world.scenario == 3 && p.Y <= 69 {
		return structureState("lava"), nil
	}
	return structureState("air"), nil
}

func ruinedPortalProbeProperties(variant int) RuinedPortalProperties {
	props := RuinedPortalProperties{Cold: variant == 1, Mossiness: 0.8, AirPocket: variant == 5, Overgrown: variant == 3, Vines: variant == 3, ReplaceWithBlackstone: variant == 4}
	if variant == 0 {
		props.Mossiness = 0
	} else if variant == 2 {
		props.Mossiness = 1
	}
	return props
}

func TestRuinedPortalOriginalTemplatesLoad(t *testing.T) {
	templates, err := loadRuinedPortalTemplates()
	if err != nil {
		t.Fatal(err)
	}
	for name, template := range templates {
		names := make(map[string]bool)
		for _, block := range template.Palettes[0] {
			names[block.State.Name] = true
			if block.State.Name == "minecraft:jigsaw" {
				t.Logf("%s jigsaw %v: %v", name, block.Position, block.NBT)
			}
		}
		t.Logf("%s: size %v, blocks %d, entities %d, palette %v", name, template.Size, len(template.Palettes[0]), len(template.Entities), names)
	}
}

func TestRuinedPortalProcessorsAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "docs", "worldgen", "ruined_portal_processors_1.21.1.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) != 11 || f[0] != "processors" {
			t.Fatalf("invalid portal vector %q", scanner.Text())
		}
		parse := func(index int) int32 {
			value, err := strconv.ParseInt(f[index], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			return int32(value)
		}
		template, err := LoadRuinedPortalTemplate(f[1])
		if err != nil {
			t.Fatal(err)
		}
		variant := int(parse(4))
		placement := RuinedPortalOnLandSurface
		if variant == 2 {
			placement = RuinedPortalOnOceanFloor
		} else if variant == 4 {
			placement = RuinedPortalInNether
		}
		world := &ruinedPortalProbeWorld{hutProbeWorld: &hutProbeWorld{scenario: int(parse(5)), edits: make(map[FeaturePosition]SurfaceBlockState)}}
		blocks, err := ProcessRuinedPortalTemplateBlocks(world, template, FeaturePosition{parse(6), parse(7), parse(8)}, placement, ruinedPortalProbeProperties(variant), TemplateMirror(parse(3)), TemplateRotation(parse(2)))
		if err != nil {
			t.Fatalf("%s: %v", scanner.Text(), err)
		}
		rows := make([]string, 0, len(blocks))
		for _, block := range blocks {
			id, err := blockstate.BlockStateID(block.State.Name, block.State.Properties)
			if err != nil {
				t.Fatal(err)
			}
			data := "null"
			if block.NBT != nil {
				data = iglooCanonicalNBT(t, nbt.Tag{Type: nbt.Compound, Value: block.NBT})
			}
			p := block.Position
			rows = append(rows, fmt.Sprintf("%d %d %d %d %s", p.X, p.Y, p.Z, id, data))
		}
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(rows, "\n"))))
		if len(blocks) != int(parse(9)) || hash != f[10] {
			t.Fatalf("%s: got %d processed blocks/%s", scanner.Text(), len(blocks), hash)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 11232 {
		t.Fatalf("got %d portal processor vectors", count)
	}
	t.Logf("matched %d full original portal processor-chain cases", count)
}
