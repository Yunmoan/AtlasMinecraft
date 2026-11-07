// Compare complete frozen vanilla chunks with the production Go generator.
package main

import (
	"atlas/internal/anvil"
	"atlas/internal/blockstate"
	"atlas/internal/nbt"
	"atlas/internal/worldgen"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/bits"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Case struct {
	Seed int64 `json:"seed"`
	X    int32 `json:"x"`
	Z    int32 `json:"z"`
}
type Counts struct {
	Total                int64
	Equal                int64
	NonAirUnion          int64
	NonAirEqual          int64
	SameBlock            int64
	SameOccupancy        int64
	SurfaceColumns       int64
	SameSurface          int64
	SurfaceAbsoluteError int64
}
type Difference struct {
	X, Y, Z     int32
	Vanilla, Go blockstate.BlockState
}
type Result struct {
	Case
	Counts
	Dimension                      string
	MinY, Height                   int32
	Seconds                        float64
	Pairs                          map[string]int64
	Examples                       []Difference
	SkippedFeatures, SkippedStages []string
	Biomes                         map[string]int
	VanillaSHA256, GoSHA256        string
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func vanilla(regionDir string, c Case, minY, height int32) []uint32 {
	_, tag, exists, err := anvil.ReadChunk(regionDir, c.X, c.Z)
	must(err)
	if !exists {
		panic(fmt.Sprintf("missing vanilla chunk %+v", c))
	}
	root, ok := tag.Object()
	if !ok {
		panic("invalid chunk compound")
	}
	status, _ := root.Text("Status")
	version, _ := root.Number("DataVersion")
	if status != "minecraft:full" || version != 3955 {
		panic(fmt.Sprintf("unfinished/wrong-version chunk: %s %d", status, version))
	}
	out := make([]uint32, height*256)
	sections, _ := root["sections"].Elements()
	for _, st := range sections {
		section, _ := st.Object()
		sy, _ := section.Number("Y")
		if sy < int64(minY/16) || sy >= int64((minY+height)/16) {
			continue
		}
		states, _ := section.Object("block_states")
		palette, _ := states["palette"].Elements()
		if len(palette) == 0 {
			panic("missing palette")
		}
		ids := make([]uint32, len(palette))
		for i, entry := range palette {
			state, _ := entry.Object()
			name, _ := state.Text("Name")
			props, _ := state.Object("Properties")
			p := map[string]string{}
			for k := range props {
				p[k], _ = props.Text(k)
			}
			ids[i], err = blockstate.BlockStateID(name, p)
			must(err)
		}
		longs, _ := states["data"].Longs()
		width := max(4, bits.Len(uint(len(palette)-1)))
		perLong := 64 / width
		if len(palette) > 1 && len(longs) < (4096+perLong-1)/perLong {
			panic("truncated palette array")
		}
		for i := 0; i < 4096; i++ {
			pi := 0
			if len(palette) > 1 {
				pi = int(uint64(longs[i/perLong]) >> ((i % perLong) * width) & ((1 << width) - 1))
			}
			if pi >= len(ids) {
				panic("invalid palette index")
			}
			out[int(sy-int64(minY/16))*4096+i] = ids[pi]
		}
	}
	return out
}

func air(name string) bool {
	return name == "minecraft:air" || name == "minecraft:cave_air" || name == "minecraft:void_air"
}

func stateBytes(states []uint32) []byte {
	raw := make([]byte, len(states)*2)
	for i, id := range states {
		if id > 65535 {
			panic("block state overflows u16 audit array")
		}
		binary.BigEndian.PutUint16(raw[i*2:], uint16(id))
	}
	return raw
}

func digest(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }

func reusable(path, output string, c Case, dimension string, minY, height int32, want []uint32) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var result Result
	if json.Unmarshal(data, &result) != nil || result.Case != c || result.Total != int64(len(want)) {
		return false
	}
	// Earlier Overworld audits did not include explicit dimension fields.
	if result.Dimension != dimension && !(dimension == "overworld" && result.Dimension == "") {
		return false
	}
	if result.Dimension != "" && (result.MinY != minY || result.Height != height) {
		return false
	}
	if result.VanillaSHA256 != digest(stateBytes(want)) {
		return false
	}
	for label, hash := range map[string]string{"vanilla": result.VanillaSHA256, "go": result.GoSHA256} {
		raw, err := os.ReadFile(filepath.Join(output, fmt.Sprintf("%d_%d.%s.u16be", c.X, c.Z, label)))
		if err != nil || len(raw) != len(want)*2 || digest(raw) != hash {
			return false
		}
	}
	return true
}

func main() {
	manifest := flag.String("cases", "", "JSON list of seed/x/z cases")
	world := flag.String("world", "", "frozen official world directory")
	output := flag.String("output", "", "per-chunk results directory")
	workers := flag.Int("workers", 2, "concurrent final chunks")
	dimension := flag.String("dimension", "overworld", "overworld, nether or end")
	verifyVanilla := flag.Bool("verify-vanilla", false, "verify saved reference arrays against the final stopped vanilla save")
	validateVanilla := flag.Bool("validate-vanilla", false, "validate all official chunks are FULL and decode every block state")
	flag.Parse()
	if *workers < 1 {
		panic("workers must be positive")
	}
	minY, height, regionDir := int32(-64), int32(384), filepath.Join(*world, "region")
	switch *dimension {
	case "overworld":
	case "nether":
		minY, height, regionDir = 0, 256, filepath.Join(*world, "DIM-1", "region")
	case "end":
		minY, height, regionDir = 0, 256, filepath.Join(*world, "DIM1", "region")
	default:
		panic("unsupported dimension")
	}
	data, err := os.ReadFile(*manifest)
	must(err)
	var cases []Case
	must(json.Unmarshal(data, &cases))
	if len(cases) == 0 {
		panic("empty cases")
	}
	f, err := os.Open(filepath.Join(*world, "level.dat"))
	must(err)
	gz, err := gzip.NewReader(f)
	must(err)
	metadata, err := io.ReadAll(gz)
	must(err)
	must(gz.Close())
	must(f.Close())
	_, levelTag, err := nbt.Parse(metadata)
	must(err)
	levelRoot, _ := levelTag.Object()
	level, _ := levelRoot.Object("Data")
	settings, _ := level.Object("WorldGenSettings")
	seed, _ := settings.Number("seed")
	structures, _ := settings.Number("generate_features")
	if seed != cases[0].Seed || structures != 1 {
		panic("vanilla seed or structure settings differ from experiment")
	}
	if *verifyVanilla || *validateVanilla {
		for _, c := range cases {
			if c.Seed != seed {
				panic("mixed seeds")
			}
			want := vanilla(regionDir, c, minY, height)
			if *validateVanilla {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(*output, fmt.Sprintf("%d_%d.vanilla.u16be", c.X, c.Z)))
			must(err)
			if len(raw) != len(want)*2 {
				panic("invalid saved reference array length")
			}
			for i, state := range want {
				if state != uint32(binary.BigEndian.Uint16(raw[i*2:])) {
					panic(fmt.Sprintf("final vanilla save changed: %+v index=%d", c, i))
				}
			}
		}
		if *validateVanilla {
			fmt.Printf("VALIDATED_FULL_VANILLA %d %d chunks\n", seed, len(cases))
		} else {
			fmt.Printf("VERIFIED_FINAL_VANILLA %d %d chunks\n", seed, len(cases))
		}
		return
	}
	must(os.MkdirAll(*output, 0755))
	names := make([]string, 26684)
	stateTable := make([]blockstate.BlockState, len(names))
	for i := range names {
		stateTable[i], err = blockstate.BlockStateByID(uint32(i))
		must(err)
		names[i] = stateTable[i].Name
	}
	var generate func(int32, int32) (*worldgen.GeneratedOverworldChunk, error)
	if *dimension == "overworld" {
		generator, err := worldgen.NewOverworldGeneratorWithOptions(cases[0].Seed, worldgen.OverworldGeneratorOptions{GenerateStructures: true})
		must(err)
		generator.SetTerrainWorkers(6)
		generate = generator.Generate
	} else {
		generator, err := worldgen.NewDimensionGeneratorWithOptions(cases[0].Seed, *dimension, worldgen.DimensionGeneratorOptions{GenerateStructures: true})
		must(err)
		generator.SetTerrainWorkers(6)
		generate = func(x, z int32) (*worldgen.GeneratedOverworldChunk, error) {
			chunk, err := generator.Generate(x, z)
			if err != nil {
				return nil, err
			}
			result := &worldgen.GeneratedOverworldChunk{Terrain: chunk.Terrain, SkippedFeatures: chunk.SkippedFeatures, SkippedStages: chunk.SkippedStages}
			copy(result.Biomes[:], chunk.Biomes)
			return result, nil
		}
	}
	jobs := make(chan Case)
	var wg sync.WaitGroup
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				if c.Seed != cases[0].Seed {
					panic("mixed seeds")
				}
				path := filepath.Join(*output, fmt.Sprintf("%d_%d.json", c.X, c.Z))
				start := time.Now()
				want := vanilla(regionDir, c, minY, height)
				if reusable(path, *output, c, *dimension, minY, height, want) {
					continue
				}
				generated, err := generate(c.X, c.Z)
				must(err)
				if generated.Terrain.MinY != minY || generated.Terrain.Height != height {
					panic("generated dimension bounds differ")
				}
				got := make([]uint32, len(want))
				cache := map[string]uint32{}
				for i := range got {
					s, err := generated.Terrain.BlockStateAt(int32(i&15), int32(i/256)+minY, int32((i>>4)&15))
					must(err)
					id, ok := cache[s.Name]
					if !ok || len(s.Properties) != 0 {
						id, err = blockstate.BlockStateID(s.Name, s.Properties)
						must(err)
						if len(s.Properties) == 0 {
							cache[s.Name] = id
						}
					}
					got[i] = id
				}
				result := Result{Case: c, Dimension: *dimension, MinY: minY, Height: height, Pairs: map[string]int64{}, Biomes: map[string]int{}, SkippedFeatures: generated.SkippedFeatures, SkippedStages: generated.SkippedStages}
				for _, section := range generated.Biomes[:height/16] {
					for _, biome := range section {
						result.Biomes[biome]++
					}
				}
				for i, a := range want {
					b := got[i]
					result.Total++
					if a == b {
						result.Equal++
					}
					aa, ba := air(names[a]), air(names[b])
					if !aa || !ba {
						result.NonAirUnion++
						if a == b {
							result.NonAirEqual++
						}
					}
					if names[a] == names[b] {
						result.SameBlock++
					}
					if aa == ba {
						result.SameOccupancy++
					}
					if a != b {
						result.Pairs[names[a]+" -> "+names[b]]++
						if len(result.Examples) < 12 {
							result.Examples = append(result.Examples, Difference{c.X*16 + int32(i&15), int32(i/256) + minY, c.Z*16 + int32((i>>4)&15), stateTable[a], stateTable[b]})
						}
					}
				}
				for col := 0; col < 256; col++ {
					ha, hb := int(minY)-1, int(minY)-1
					for y := int(minY+height) - 1; y >= int(minY); y-- {
						i := (y-int(minY))*256 + col
						if ha == int(minY)-1 && !air(names[want[i]]) {
							ha = y
						}
						if hb == int(minY)-1 && !air(names[got[i]]) {
							hb = y
						}
						if ha != int(minY)-1 && hb != int(minY)-1 {
							break
						}
					}
					result.SurfaceColumns++
					if ha == hb {
						result.SameSurface++
					}
					delta := ha - hb
					if delta < 0 {
						delta = -delta
					}
					result.SurfaceAbsoluteError += int64(delta)
				}
				// Preserve exact arrays so aggregate results can be independently audited.
				for label, array := range map[string][]uint32{"vanilla": want, "go": got} {
					raw := stateBytes(array)
					must(os.WriteFile(filepath.Join(*output, fmt.Sprintf("%d_%d.%s.u16be", c.X, c.Z, label)), raw, 0644))
					hash := digest(raw)
					if label == "vanilla" {
						result.VanillaSHA256 = hash
					} else {
						result.GoSHA256 = hash
					}
				}
				result.Seconds = time.Since(start).Seconds()
				encoded, err := json.MarshalIndent(result, "", "  ")
				must(err)
				must(os.WriteFile(path+".tmp", encoded, 0644))
				must(os.Rename(path+".tmp", path))
				fmt.Printf("COMPARED %d %d %d %.6f%% %.1fs\n", c.Seed, c.X, c.Z, 100*float64(result.Equal)/float64(result.Total), result.Seconds)
			}
		}()
	}
	for _, c := range cases {
		jobs <- c
	}
	close(jobs)
	wg.Wait()
}
