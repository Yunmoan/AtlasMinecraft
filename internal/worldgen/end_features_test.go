package worldgen

import (
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func endFeatureProbeBase(p, origin FeaturePosition, spikes bool, mode int) SurfaceBlockState {
	s := func(name string) SurfaceBlockState { return SurfaceBlockState{Name: "minecraft:" + name} }
	mod := func(n, d int32) int32 { return (n%d + d) % d }
	if p.Y < 0 || p.Y >= 256 || mode == 3 {
		return s("air")
	}
	if mode == 4 {
		return s("stone")
	}
	dx, dy, dz := p.X-origin.X, p.Y-origin.Y, p.Z-origin.Z
	floor := origin.Y - 1
	if spikes {
		floor = 66
	}
	if mode == 10 {
		floor += mod(dx+dz, 5) - 2
	}
	if p.Y <= floor {
		switch mode {
		case 5:
			return s("stone")
		case 6:
			return s("water")
		case 7:
			return s("obsidian")
		}
		return s("end_stone")
	}
	if mode == 8 && dy == 0 {
		return s("water")
	}
	if mode == 9 && dy == 0 {
		return s("lava")
	}
	if mode == 11 && dy >= 2 && dy <= 5 && abs32(dx) <= 2 && abs32(dz) <= 2 {
		return s("stone")
	}
	if mode == 16 && dy >= 0 && dy <= 6 && dx == 2 {
		return s("chorus_plant")
	}
	if mode == 17 && dy >= 0 && dy <= 6 && mod(dx+dz, 3) == 0 {
		return s("end_stone")
	}
	return s("air")
}

func endFeatureProbeRegion(t *testing.T, origin FeaturePosition, spikes bool, mode int) *OverworldFeatureRegion {
	t.Helper()
	var chunks []*NoiseChunkResult
	for cz := (origin.Z >> 4) - 1; cz <= (origin.Z>>4)+1; cz++ {
		for cx := (origin.X >> 4) - 1; cx <= (origin.X>>4)+1; cx++ {
			ch := &NoiseChunkResult{X: cx, Z: cz, MinY: 0, Height: 256, Blocks: make([]NoiseMaterial, 256*256), SurfaceBlocks: map[int]SurfaceBlockState{}}
			for y := int32(0); y < 256; y++ {
				for z := int32(0); z < 16; z++ {
					for x := int32(0); x < 16; x++ {
						s := endFeatureProbeBase(FeaturePosition{cx*16 + x, y, cz*16 + z}, origin, spikes, mode)
						i, _ := ch.Index(x, y, z)
						switch s.Name {
						case "minecraft:air":
						case "minecraft:end_stone":
							ch.Blocks[i] = NoiseEndStone
						case "minecraft:stone":
							ch.Blocks[i] = NoiseStone
						case "minecraft:water":
							ch.Blocks[i] = NoiseWater
						case "minecraft:lava":
							ch.Blocks[i] = NoiseLava
						default:
							ch.SurfaceBlocks[i] = s
						}
					}
				}
			}
			chunks = append(chunks, ch)
		}
	}
	region, err := NewDimensionFeatureRegionWithSeaLevel(128, 0, chunks...)
	if err != nil {
		t.Fatal(err)
	}
	region.borrowed = map[featureChunkPosition]bool{}
	for key := range region.chunks {
		region.borrowed[key] = true
	}
	return region
}

func compileEndFeatureProbePlan(name string) (configuredFeaturePlacer, error) {
	var def ConfiguredFeatureDefinition
	var err error
	switch name {
	case "minecraft:end_gateway_none":
		def = ConfiguredFeatureDefinition{Type: "minecraft:end_gateway", Config: json.RawMessage(`{"exact":false}`)}
	case "minecraft:end_gateway_inexact":
		def = ConfiguredFeatureDefinition{Type: "minecraft:end_gateway", Config: json.RawMessage(`{"exact":false,"exit":[-100,50,123]}`)}
	case "minecraft:end_spike_guarded":
		def = ConfiguredFeatureDefinition{Type: "minecraft:end_spike", Config: json.RawMessage(`{"spikes":[{"centerX":8,"centerZ":8,"radius":2,"height":76,"guarded":true}],"crystal_invulnerable":false}`)}
	case "minecraft:end_spike_beam":
		def = ConfiguredFeatureDefinition{Type: "minecraft:end_spike", Config: json.RawMessage(`{"spikes":[{"centerX":8,"centerZ":8,"radius":2,"height":76,"guarded":true},{"centerX":10,"centerZ":10,"radius":3,"height":80,"guarded":false}],"crystal_invulnerable":true,"crystal_beam_target":[0,128,0]}`)}
	default:
		def, err = LoadConfiguredFeatureDefinition(name)
		if err != nil {
			return nil, err
		}
	}
	switch def.Type {
	case "minecraft:end_island":
		return compileEndIslandPlan(def)
	case "minecraft:chorus_plant":
		return compileChorusPlantPlan(def)
	case "minecraft:end_gateway":
		return compileEndGatewayPlan(def)
	case "minecraft:end_platform":
		return compileEndPlatformPlan(def)
	case "minecraft:end_spike":
		return compileEndSpikePlan(def)
	}
	return nil, fmt.Errorf("unsupported End probe %s", name)
}

func TestEndFeaturesAgainstVanillaJAR(t *testing.T) {
	file, err := os.Open("../../docs/worldgen/end_features_1.21.1.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	parse := func(s string) int64 {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	parseFloat := func(s string) float64 {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	regions := map[string]*OverworldFeatureRegion{}
	var current *OverworldFeatureRegion
	cases, positive, layouts, gateways, crystals := 0, 0, 0, 0, 0
	entityIndex := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		switch f[0] {
		case "end_spike_layout":
			if len(f) != 8 {
				t.Fatal(scanner.Text())
			}
			s := endSpikesForSeed(parse(f[1]))[parse(f[2])]
			want := endSpike{int32(parse(f[3])), int32(parse(f[4])), int32(parse(f[5])), int32(parse(f[6])), f[7] == "1"}
			if s != want {
				t.Errorf("%s got %+v", scanner.Text(), s)
			}
			layouts++
		case "end_feature":
			if len(f) != 11 {
				t.Fatal(scanner.Text())
			}
			seed, mode := parse(f[2]), int(parse(f[3]))
			origin := FeaturePosition{int32(parse(f[4])), int32(parse(f[5])), int32(parse(f[6]))}
			spikes := strings.HasPrefix(f[1], "minecraft:end_spike")
			key := fmt.Sprintf("%v/%d/%v", spikes, mode, origin)
			if regions[key] == nil {
				regions[key] = endFeatureProbeRegion(t, origin, spikes, mode)
			}
			region := regions[key]
			current = region
			entityIndex = 0
			region.overlays = nil
			region.blockEntities = nil
			region.pendingEntities = nil
			region.biomeSeed = seed
			place, err := compileEndFeatureProbePlan(f[1])
			if err != nil {
				t.Fatal(err)
			}
			r := NewWorldgenRandom(seed)
			placed, err := place(region, r, origin)
			if err != nil {
				t.Fatal(err)
			}
			var output []string
			for key, overlay := range region.overlays {
				for i, s := range overlay.blocks {
					id, err := blockstate.BlockStateID(s.Name, s.Properties)
					if err != nil {
						t.Fatal(err)
					}
					output = append(output, fmt.Sprintf("%d %d %d %d", key.chunkX*16+int32(i&15), int32(i/256), key.chunkZ*16+int32((i>>4)&15), id))
				}
			}
			sort.Strings(output)
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(output, "\n"))))
			next := int64(r.NextLong())
			if placed != (f[7] == "1") || len(output) != int(parse(f[8])) || next != parse(f[9]) || hash != f[10] {
				t.Errorf("%s seed=%d mode=%d placed=%v writes=%d next=%d hash=%s; want %s", f[1], seed, mode, placed, len(output), next, hash, scanner.Text())
			}
			cases++
			if placed {
				positive++
			}
		case "end_gateway_nbt":
			if len(f) != 8 || current == nil {
				t.Fatal(scanner.Text())
			}
			p := FeaturePosition{int32(parse(f[4])), int32(parse(f[5])), int32(parse(f[6]))}
			entity, ok := current.blockEntities[p]
			if !ok {
				t.Fatalf("missing gateway %s", scanner.Text())
			}
			data, err := hex.DecodeString(f[7])
			if err != nil {
				t.Fatal(err)
			}
			expected, used, err := nbt.ParseNetwork(data)
			if err != nil || used != len(data) {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(nbt.Tag{Type: nbt.Compound, Value: entity.data}, expected) {
				t.Errorf("gateway %s got %v want %v", scanner.Text(), entity.data, expected)
			}
			gateways++
		case "end_crystal":
			if len(f) != 10 || current == nil || entityIndex >= len(current.pendingEntities) {
				t.Fatalf("invalid or missing crystal %s", scanner.Text())
			}
			entity := current.pendingEntities[entityIndex]
			entityIndex++
			pos, _ := entity.Data["Pos"].Elements()
			rot, _ := entity.Data["Rotation"].Elements()
			inv, _ := entity.Data["Invulnerable"].Number()
			show, _ := entity.Data["ShowBottom"].Number()
			if entity.ID != "minecraft:end_crystal" || len(pos) != 3 || len(rot) != 2 || pos[0].Value != parseFloat(f[4]) || pos[1].Value != parseFloat(f[5]) || pos[2].Value != parseFloat(f[6]) || math.Float32bits(rot[0].Value.(float32)) != uint32(parse(f[7])) || inv != parse(f[8]) || show != 1 {
				t.Errorf("%s actual crystal %v", scanner.Text(), entity)
			}
			beam, hasBeam := entity.Data["beam_target"]
			if f[9] == "none" {
				if hasBeam {
					t.Error("unexpected crystal beam target")
				}
			} else {
				coords := strings.Split(f[9], ",")
				want := []int32{int32(parse(coords[0])), int32(parse(coords[1])), int32(parse(coords[2]))}
				if !hasBeam || !reflect.DeepEqual(beam.Value, want) {
					t.Error("incorrect crystal beam target")
				}
			}
			if _, err := nbt.Marshal("", nbt.Tag{Type: nbt.Compound, Value: entity.Data}); err != nil {
				t.Fatal(err)
			}
			crystals++
		default:
			t.Fatal(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if cases != 1296 || layouts != 80 || gateways != 432 || crystals != 488 {
		t.Fatalf("coverage features=%d layouts=%d gateways=%d crystals=%d", cases, layouts, gateways, crystals)
	}
	t.Logf("matched %d JAR feature cases (%d positive), %d seed-derived spike entries, %d gateway NBT and %d actual crystals", cases, positive, layouts, gateways, crystals)
}
